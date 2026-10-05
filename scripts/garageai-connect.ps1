<#
.SYNOPSIS
  garageai-connect.ps1 - connect a Windows garage (GPU machine) to the GarageAI mesh.

.DESCRIPTION
  Windows version of garageai-connect.sh, for Ollama and LM Studio (and any other
  OpenAI-compatible runtime that runs natively on Windows). It:
    1. installs the NetBird client if it is missing (open source, NetBird GmbH, Berlin)
    2. joins the private GarageAI WireGuard mesh using your setup key
    3. checks that your runtime answers an OpenAI-compatible API
    4. makes sure the gateway can reach it: the runtime must listen on the network, and a
       Windows Firewall rule lets ONLY the mesh (100.64.0.0/10) reach the runtime's port
    5. registers the models you offer; GarageAI tests each one before it is sold
    6. installs a heartbeat (a scheduled task) that reports your models every 5 minutes

  Your runtime is never exposed to the internet or to your home network by this script.
  Run it in a PowerShell window opened with "Run as administrator".

  Every option can also be given as an environment variable (GARAGEAI_SETUP_KEY,
  GARAGEAI_MANAGEMENT_URL, GARAGEAI_RUNTIME, GARAGEAI_PORT, GARAGEAI_NODE_NAME,
  GARAGEAI_RUNTIME_API_KEY, GARAGEAI_MODELS, GARAGEAI_REGISTER_URL, GARAGEAI_REGISTER_TOKEN).
  Secrets are best passed that way, so they do not end up in the command history.

.EXAMPLE
  $env:GARAGEAI_SETUP_KEY = "..."; $env:GARAGEAI_REGISTER_TOKEN = "grg_..."
  .\garageai-connect.ps1 -Name my-garage -Runtime ollama -ManagementUrl https://netbird.garageai.eu `
      -RegisterUrl https://<portal>/functions/v1/register-node

.EXAMPLE
  .\garageai-connect.ps1 -Doctor        # check this garage and say exactly what to fix
  .\garageai-connect.ps1 -Uninstall     # remove the heartbeat and firewall rule, leave the mesh
#>
[CmdletBinding()]
param(
  [string]$SetupKey = $env:GARAGEAI_SETUP_KEY,
  [string]$ManagementUrl = $env:GARAGEAI_MANAGEMENT_URL,
  [string]$Runtime = $env:GARAGEAI_RUNTIME,
  [int]$Port = 0,
  [string]$Name = $env:GARAGEAI_NODE_NAME,
  [string]$RuntimeApiKey = $env:GARAGEAI_RUNTIME_API_KEY,
  [string]$Models = $env:GARAGEAI_MODELS,
  [string]$RegisterUrl = $env:GARAGEAI_REGISTER_URL,
  [string]$RegisterToken = $env:GARAGEAI_REGISTER_TOKEN,
  [switch]$SkipInstall,
  [switch]$NoHeartbeat,
  [switch]$Doctor,
  [switch]$Uninstall,
  [switch]$Yes,
  [switch]$Help
)

$ErrorActionPreference = 'Stop'
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

$DataDir       = Join-Path $env:ProgramData 'GarageAI'
$HeartbeatPs1  = Join-Path $DataDir 'heartbeat.ps1'
$HeartbeatConf = Join-Path $DataDir 'heartbeat.json'
$HeartbeatLog  = Join-Path $DataDir 'heartbeat.log'
$TaskName      = 'GarageAI heartbeat'
$FirewallRule  = 'GarageAI runtime (mesh only)'
$MeshRange     = '100.64.0.0/10'   # the address range NetBird hands out; your LAN is not in it
$NonChatRe     = 'embed|bge-|bge:|e5-|minilm|rerank|colbert|gte-'

function Write-Head($t) { Write-Host $t -ForegroundColor White }
function Write-Ok($t)   { Write-Host "  [ok] $t" -ForegroundColor Green }
function Write-Info($t) { Write-Host "  $t" }
function Write-Warn2($t) { Write-Host "  [!] $t" -ForegroundColor Yellow }
function Stop-With($t)  { Write-Host "  [x] $t" -ForegroundColor Red; exit 1 }

function Confirm-Step($question) {
  if ($Yes) { return $true }
  $answer = Read-Host "  $question [y/N]"
  return ($answer -match '^(y|yes)$')
}

function Test-Admin {
  $id = [Security.Principal.WindowsIdentity]::GetCurrent()
  return (New-Object Security.Principal.WindowsPrincipal($id)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Get-DefaultPort($rt) {
  switch ($rt) {
    'ollama'   { 11434 }
    'lmstudio' { 1234 }
    'llamacpp' { 8080 }
    default    { 0 }
  }
}

function Get-NetbirdExe {
  $cmd = Get-Command netbird -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd.Source }
  $default = Join-Path $env:ProgramFiles 'NetBird\netbird.exe'
  if (Test-Path $default) { return $default }
  return $null
}

function Get-MeshStatus {
  # @{ Connected = bool; Ip = '100.x.y.z' } or $null when NetBird is missing or not running.
  $nb = Get-NetbirdExe
  if (-not $nb) { return $null }
  try {
    $s = (& $nb status --json 2>$null | Out-String) | ConvertFrom-Json
    $ip = ''
    if ($s.netbirdIp) { $ip = ($s.netbirdIp -split '/')[0] }
    return @{ Connected = [bool]$s.management.connected; Ip = $ip }
  } catch { return $null }
}

function Get-RuntimeModels($hostName) {
  # Model ids served by the runtime, or $null if no OpenAI-compatible API answers.
  $headers = @{}
  if ($RuntimeApiKey) { $headers['Authorization'] = "Bearer $RuntimeApiKey" }
  try {
    $r = Invoke-RestMethod -Uri "http://${hostName}:$Port/v1/models" -Headers $headers -TimeoutSec 5
    return @($r.data | ForEach-Object { $_.id })
  } catch { return $null }
}

function Get-ListenAddresses {
  # Local addresses the runtime listens on for $Port ('0.0.0.0' and '::' mean all interfaces).
  try {
    return @(Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction Stop |
             Select-Object -ExpandProperty LocalAddress -Unique)
  } catch { return @() }
}

function Test-ListensOnNetwork($addresses, $meshIp) {
  foreach ($a in $addresses) { if ($a -eq '0.0.0.0' -or $a -eq '::' -or ($meshIp -and $a -eq $meshIp)) { return $true } }
  return $false
}

function Select-OfferedModels($all) {
  if ($Models) {
    $wanted = @($Models -split ',' | ForEach-Object { $_.Trim() } | Where-Object { $_ })
    foreach ($w in $wanted) { if ($all -notcontains $w) { Write-Warn2 "-Models: '$w' is not served by the runtime; skipped" } }
    return @($wanted | Where-Object { $all -contains $_ })
  }
  return @($all | Where-Object { $_ -notmatch $NonChatRe })
}

function Show-RuntimeHint {
  switch ($Runtime) {
    'ollama' {
      Write-Info 'Ollama listens on 127.0.0.1 by default. To make it listen on the network:'
      Write-Info '    setx OLLAMA_HOST 0.0.0.0'
      Write-Info '  then quit Ollama from the system tray (right-click the icon -> Quit) and start it again.'
      Write-Info '  Optional, to serve several buyers at once:  setx OLLAMA_NUM_PARALLEL 4'
    }
    'lmstudio' {
      Write-Info "LM Studio: Developer tab -> start the server on port $Port and turn on"
      Write-Info "  `"Serve on Local Network`". Headless:  lms server start --bind 0.0.0.0 --port $Port"
    }
    'llamacpp' {
      Write-Info "llama.cpp:  llama-server.exe -m model.gguf --host 0.0.0.0 --port $Port -np 4 --jinja"
    }
    default {
      Write-Info "Start your OpenAI-compatible server on port $Port, bound to 0.0.0.0 instead of 127.0.0.1."
    }
  }
  Write-Info 'This script adds a firewall rule that only lets the GarageAI mesh reach that port.'
}

function Set-MeshFirewallRule {
  Get-NetFirewallRule -DisplayName $FirewallRule -ErrorAction SilentlyContinue | Remove-NetFirewallRule
  New-NetFirewallRule -DisplayName $FirewallRule -Direction Inbound -Action Allow -Protocol TCP `
    -LocalPort $Port -RemoteAddress $MeshRange -Profile Any `
    -Description 'Lets the GarageAI gateway reach the inference runtime over the NetBird mesh only.' | Out-Null
}

function Remove-Heartbeat {
  Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false -ErrorAction SilentlyContinue
  Remove-Item -Force -ErrorAction SilentlyContinue $HeartbeatPs1, $HeartbeatConf, $HeartbeatLog
}

function Install-Heartbeat($meshIp) {
  New-Item -ItemType Directory -Force -Path $DataDir | Out-Null
  # The register token and runtime key are secrets: only SYSTEM and Administrators may read the folder.
  & icacls $DataDir /inheritance:r /grant:r 'SYSTEM:(OI)(CI)F' 'Administrators:(OI)(CI)F' | Out-Null

  $heartbeatUrl = ($RegisterUrl -replace '/register-node$', '') + '/node-heartbeat'
  @{ url = $heartbeatUrl; token = $RegisterToken; name = $Name; runtime = $Runtime; port = $Port
     runtimeApiKey = $RuntimeApiKey; meshIp = $meshIp; models = $Models; nonChatRe = $NonChatRe } |
    ConvertTo-Json | Set-Content -Path $HeartbeatConf -Encoding UTF8

  @'
# GarageAI heartbeat - reports this garage's current models to GarageAI every 5 minutes.
# Installed by garageai-connect.ps1; remove with: garageai-connect.ps1 -Uninstall
$ErrorActionPreference = 'Stop'
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
$dir = Split-Path -Parent $MyInvocation.MyCommand.Path
$c = Get-Content (Join-Path $dir 'heartbeat.json') -Raw | ConvertFrom-Json
$log = Join-Path $dir 'heartbeat.log'
$headers = @{}
if ($c.runtimeApiKey) { $headers['Authorization'] = "Bearer $($c.runtimeApiKey)" }
$models = @()
foreach ($h in @('127.0.0.1', $c.meshIp)) {
  if (-not $h) { continue }
  try {
    $r = Invoke-RestMethod -Uri "http://${h}:$($c.port)/v1/models" -Headers $headers -TimeoutSec 5
    $models = @($r.data | ForEach-Object { $_.id }); break
  } catch { }
}
# If the runtime does not answer, report no models so buyers are not routed here.
if ($c.models) {
  $allow = @($c.models -split ',' | ForEach-Object { $_.Trim() } | Where-Object { $_ })
  $models = @($models | Where-Object { $allow -contains $_ })
} else {
  $models = @($models | Where-Object { $_ -notmatch $c.nonChatRe })
}
$payload = @{ name = $c.name; port = [int]$c.port; runtime = $c.runtime; models = $models }
if ($c.runtimeApiKey) { $payload['runtime_api_key'] = $c.runtimeApiKey }
$stamp = (Get-Date).ToUniversalTime().ToString('s') + 'Z'
try {
  $res = Invoke-RestMethod -Method Post -Uri $c.url -Headers @{ Authorization = "Bearer $($c.token)" } `
    -ContentType 'application/json' -Body (ConvertTo-Json $payload -Depth 5 -Compress) -TimeoutSec 180
  "$stamp ok " + (ConvertTo-Json $res -Depth 5 -Compress) | Set-Content -Path $log -Encoding UTF8
} catch {
  "$stamp error $($_.Exception.Message)" | Set-Content -Path $log -Encoding UTF8
  exit 1
}
'@ | Set-Content -Path $HeartbeatPs1 -Encoding UTF8

  $action    = New-ScheduledTaskAction -Execute 'powershell.exe' -Argument "-NoProfile -NonInteractive -ExecutionPolicy Bypass -File `"$HeartbeatPs1`""
  $trigger   = New-ScheduledTaskTrigger -Once -At (Get-Date).AddMinutes(1) -RepetitionInterval (New-TimeSpan -Minutes 5) -RepetitionDuration (New-TimeSpan -Days 3650)
  $boot      = New-ScheduledTaskTrigger -AtStartup
  $principal = New-ScheduledTaskPrincipal -UserId 'SYSTEM' -LogonType ServiceAccount -RunLevel Highest
  $settings  = New-ScheduledTaskSettingsSet -StartWhenAvailable -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit (New-TimeSpan -Minutes 5)
  Register-ScheduledTask -TaskName $TaskName -Action $action -Trigger @($trigger, $boot) -Principal $principal -Settings $settings -Force | Out-Null
}

function Invoke-Doctor {
  function Report-Bad($what, $fix) { Write-Host "  [x] $what" -ForegroundColor Red; Write-Info "      -> $fix"; $script:DoctorProblems++ }
  $script:DoctorProblems = 0
  Write-Head "GarageAI doctor - runtime $Runtime, port $Port"

  # Level 1: the tunnel
  $mesh = Get-MeshStatus
  if (-not (Get-NetbirdExe)) { Report-Bad 'NetBird is not installed' 'run the connect command from the GarageAI portal' }
  elseif ($mesh -and $mesh.Connected -and $mesh.Ip) { Write-Ok "Tunnel: connected to the GarageAI mesh as $($mesh.Ip)" }
  else { Report-Bad 'Tunnel: NetBird is installed but not connected' 'run "netbird up" in an administrator PowerShell (or get a new command from the portal)' }
  $meshIp = ''
  if ($mesh) { $meshIp = $mesh.Ip }

  # Level 2: the runtime
  $found = Get-RuntimeModels '127.0.0.1'
  if ($null -eq $found -and $meshIp) { $found = Get-RuntimeModels $meshIp }
  if ($null -eq $found) {
    Report-Bad "Runtime: nothing answers on port $Port" "start $Runtime (and check -Runtime / -Port if you use another one)"
  } else {
    Write-Ok "Runtime: $Runtime answers on port $Port with $($found.Count) model(s): $($found -join ' ')"
    $listen = Get-ListenAddresses
    if ($listen.Count -eq 0) { Write-Info '(could not read which address it listens on)' }
    elseif (Test-ListensOnNetwork $listen $meshIp) { Write-Ok "Runtime: listens on the network ($($listen -join ' ')), so the gateway can reach it" }
    else { Report-Bad "Runtime: only listens on $($listen -join ' '), so the gateway cannot reach it" 'see below'; Show-RuntimeHint }
    $rule = Get-NetFirewallRule -DisplayName $FirewallRule -ErrorAction SilentlyContinue
    if ($rule) { Write-Ok 'Firewall: the mesh-only rule is in place' }
    else { Report-Bad 'Firewall: no rule lets the mesh reach the runtime' 'run the connect command from the portal again (as administrator)' }
  }

  if (-not (Test-Admin)) {
    Write-Info 'Heartbeat: not checked (open PowerShell with "Run as administrator" for the complete check)'
    Write-Host ''
    if ($script:DoctorProblems -eq 0) { Write-Ok 'No problems found in the checks that could run.'; exit 0 }
    Write-Warn2 "$($script:DoctorProblems) problem(s) found - fix the lines marked [x] from the top down, then run -Doctor again."
    exit 1
  }

  # Heartbeat
  $task = Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue
  if (-not $task) {
    Report-Bad 'Heartbeat: not installed, so model changes and outages are noticed late' 'run the connect command from the portal again ("New command")'
  } else {
    if ($task.State -eq 'Disabled') { Report-Bad 'Heartbeat: the scheduled task is disabled' "enable '$TaskName' in Task Scheduler" }
    else { Write-Ok 'Heartbeat: installed and scheduled' }
    $last = ''
    try { $last = (Get-Content $HeartbeatLog -Tail 1 -ErrorAction Stop) } catch { }
    if ($last -match '"ok":true') { Write-Ok 'Heartbeat: last report was accepted by GarageAI' }
    elseif ($last -match '401|nauthorized') { Report-Bad "Heartbeat: GarageAI rejects this garage's token (it was replaced or revoked)" 'My garages -> New command, and run that command here' }
    elseif ($last) { Write-Info "last heartbeat output: $($last.Substring(0, [Math]::Min(120, $last.Length)))" }
  }

  Write-Info 'Note: a sleeping PC is offline for buyers. For a garage that should stay up:'
  Write-Info '      -> Settings -> System -> Power -> Sleep: Never (when plugged in)'
  Write-Host ''
  if ($script:DoctorProblems -eq 0) { Write-Ok 'No problems found. If the portal still shows the garage as offline, use Retest there.'; exit 0 }
  Write-Warn2 "$($script:DoctorProblems) problem(s) found - fix the lines marked [x] from the top down, then run -Doctor again."
  exit 1
}

# ---------------------------------------------------------------------------------------------

if ($Help) { Get-Help $MyInvocation.MyCommand.Path -Detailed; exit 0 }

if (-not $Runtime) { $Runtime = 'ollama' }
if ($Port -eq 0 -and $env:GARAGEAI_PORT) { $Port = [int]$env:GARAGEAI_PORT }
if (-not $Name) { $Name = $env:COMPUTERNAME.ToLower() }

if (@('ollama', 'lmstudio', 'llamacpp', 'other') -notcontains $Runtime) {
  Stop-With "Unknown runtime '$Runtime' (use ollama, lmstudio, llamacpp or other). vLLM and SGLang need Linux; on Windows they run in WSL2, which this script does not support yet."
}
if ($Port -eq 0) { $Port = Get-DefaultPort $Runtime }
if ($Port -eq 0) { Stop-With "Runtime '$Runtime' has no default port - pass -Port with the port its OpenAI-compatible server listens on." }

if ($Doctor) { Invoke-Doctor }

if (-not (Test-Admin)) {
  Stop-With 'Run this in a PowerShell window opened with "Run as administrator" (needed for NetBird, the firewall rule and the heartbeat task).'
}

if ($Uninstall) {
  Write-Head 'Remove GarageAI from this machine'
  Write-Info 'This removes the heartbeat task and the firewall rule and disconnects from the mesh.'
  Write-Info 'Your runtime and models are not touched. NetBird itself stays installed.'
  if (-not (Confirm-Step 'Continue?')) { Stop-With 'Aborted.' }
  Remove-Heartbeat; Write-Ok 'Heartbeat removed'
  Get-NetFirewallRule -DisplayName $FirewallRule -ErrorAction SilentlyContinue | Remove-NetFirewallRule
  Write-Ok 'Firewall rule removed'
  $nb = Get-NetbirdExe
  if ($nb) { & $nb down 2>$null | Out-Null; Write-Ok 'Disconnected from the mesh'; Write-Info 'To remove NetBird too: Settings -> Apps -> NetBird -> Uninstall' }
  Write-Info 'If you set OLLAMA_HOST for GarageAI, remove it with:  reg delete HKCU\Environment /v OLLAMA_HOST /f'
  Write-Info 'Finally, remove the garage under My garages in the portal so it is not offered again.'
  exit 0
}

Write-Head "GarageAI node connect - $Name"
Write-Host ''

# 1. NetBird client
Write-Head '1/6  NetBird client'
$nb = Get-NetbirdExe
if ($nb) {
  Write-Ok "netbird is installed ($((& $nb version 2>$null | Out-String).Trim()))"
} elseif ($SkipInstall) {
  Stop-With 'netbird is not installed and -SkipInstall was given.'
} else {
  Write-Info 'NetBird is not installed. It will be downloaded from https://pkgs.netbird.io/windows/msi/x64'
  if (-not (Confirm-Step 'Install NetBird now?')) { Stop-With 'Aborted. Install NetBird yourself (https://netbird.io) and run this again.' }
  $msi = Join-Path $env:TEMP 'netbird-installer.msi'
  Invoke-WebRequest -Uri 'https://pkgs.netbird.io/windows/msi/x64' -OutFile $msi -UseBasicParsing
  $p = Start-Process msiexec.exe -ArgumentList "/i `"$msi`" /quiet /norestart" -Wait -PassThru
  if ($p.ExitCode -ne 0) { Stop-With "The NetBird installer failed (exit code $($p.ExitCode))." }
  $nb = Get-NetbirdExe
  if (-not $nb) { Stop-With 'NetBird was installed but netbird.exe was not found. Open a new PowerShell window and run this again.' }
  Write-Ok 'netbird installed'
}
Write-Host ''

# 2. Join the mesh
Write-Head '2/6  Join the GarageAI mesh'
$mesh = Get-MeshStatus
if ($mesh -and $mesh.Ip -and -not $SetupKey) {
  Write-Ok 'Already on the mesh (no setup key given, keeping the current connection)'
} else {
  if (-not $SetupKey) { Stop-With 'No setup key. Pass -SetupKey (you get one from GarageAI).' }
  if (-not $ManagementUrl) { Stop-With 'No management URL. Pass -ManagementUrl (you get it from GarageAI).' }
  # The setup key goes in the environment, not on the command line. The peer is named after the garage.
  $env:NB_SETUP_KEY = $SetupKey
  & $nb up --management-url $ManagementUrl --hostname $Name | Out-Null
  Remove-Item Env:\NB_SETUP_KEY -ErrorAction SilentlyContinue
  for ($i = 0; $i -lt 30; $i++) { $mesh = Get-MeshStatus; if ($mesh -and $mesh.Ip) { break }; Start-Sleep -Seconds 1 }
}
if (-not ($mesh -and $mesh.Ip)) { Stop-With 'Could not join the mesh. Check the setup key (single-use, 3 days) and try again.' }
$MeshIp = $mesh.Ip
Write-Ok "Mesh IP: $MeshIp"
Write-Host ''

# 3. Local runtime
Write-Head "3/6  Inference runtime ($Runtime, port $Port)"
$all = Get-RuntimeModels '127.0.0.1'
if ($null -eq $all) { $all = Get-RuntimeModels $MeshIp }
if ($null -eq $all) {
  Write-Warn2 "No OpenAI-compatible API answers on port $Port."
  Show-RuntimeHint
  Stop-With "Start $Runtime and run this again."
}
Write-Ok "OpenAI-compatible API answers on port $Port"
Write-Host ''

# 4. Reachable over the mesh
Write-Head '4/6  Reachable over the mesh'
$listen = Get-ListenAddresses
if (-not (Test-ListensOnNetwork $listen $MeshIp)) {
  Write-Warn2 "The runtime only listens on $($listen -join ' '), so the gateway cannot reach it."
  if ($Runtime -eq 'ollama' -and (Confirm-Step 'Make Ollama listen on the network permanently (sets OLLAMA_HOST for your user)?')) {
    [Environment]::SetEnvironmentVariable('OLLAMA_HOST', '0.0.0.0', 'User')
    Write-Ok 'Done. Quit Ollama from the system tray (right-click -> Quit), start it again, then run this script again.'
    exit 0
  }
  Show-RuntimeHint
  Stop-With 'Restart the runtime listening on the network and run this script again.'
}
Write-Ok "Listening on $($listen -join ' ') (port $Port)"
Set-MeshFirewallRule
Write-Ok "Firewall: only the mesh ($MeshRange) may reach port $Port"

$offered = Select-OfferedModels $all
foreach ($m in $all) {
  if ($offered -contains $m) { Write-Info "model: $m" }
  elseif (-not $Models) { Write-Info "model: $m (skipped: embedding/reranker model)" }
  else { Write-Info "model: $m (not offered)" }
}
if ($offered.Count -eq 0) { Stop-With "No chat model to offer. Load one in $Runtime (or check -Models) and run this again." }
Write-Host ''

# 5. Register
Write-Head '5/6  Register with GarageAI'
$payload = @{ name = $Name; mesh_ip = $MeshIp; port = $Port; runtime = $Runtime; models = @($offered) }
if ($RuntimeApiKey) { $payload['runtime_api_key'] = $RuntimeApiKey }
if (-not $RegisterUrl) {
  Write-Info 'No -RegisterUrl given. Send these details to GarageAI to activate the node:'
  $shown = $payload.Clone(); $shown.Remove('runtime_api_key')
  Write-Host (ConvertTo-Json $shown -Depth 5)
  exit 0
}
if (-not $RegisterToken) { Stop-With '-RegisterUrl given without a register token (GARAGEAI_REGISTER_TOKEN).' }
Write-Info 'Registering and running the acceptance test (a real request through the gateway)...'
try {
  $res = Invoke-RestMethod -Method Post -Uri $RegisterUrl -Headers @{ Authorization = "Bearer $RegisterToken" } `
    -ContentType 'application/json' -Body (ConvertTo-Json $payload -Depth 5 -Compress) -TimeoutSec 180
} catch {
  Stop-With "Registration request to $RegisterUrl failed: $($_.Exception.Message)"
}
$passed = 0
foreach ($a in @($res.acceptance)) {
  if ($a.passed) { $passed++; Write-Ok "$($a.model): passed ($($a.tokens_per_second) tok/s, first token after $($a.ttft_ms) ms)" }
  else { Write-Warn2 "$($a.model): failed ($($a.error))" }
}
if ($passed -gt 0) { Write-Ok 'Node registered - your garage is live on GarageAI.' }
else {
  Write-Warn2 'Registered, but no model passed the acceptance test, so nothing is for sale yet.'
  Write-Warn2 'Run this script with -Doctor to see what to fix.'
}
Write-Host ''

# 6. Heartbeat
if (-not $NoHeartbeat) {
  Write-Head '6/6  Heartbeat'
  Install-Heartbeat $MeshIp
  Write-Ok 'Installed: a scheduled task reports your models every 5 minutes.'
  Write-Info 'Remove everything with:  .\garageai-connect.ps1 -Uninstall'
}
