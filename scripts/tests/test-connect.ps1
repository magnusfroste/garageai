# Tests for garageai-connect.ps1 that need no NetBird, no GPU and no network:
# the script parses, validates options and -Doctor works against a fake runtime.
# Run under both Windows PowerShell 5.1 and PowerShell 7.
$ErrorActionPreference = 'Stop'
$here   = Split-Path -Parent $MyInvocation.MyCommand.Path
$script = Join-Path (Split-Path -Parent $here) 'garageai-connect.ps1'
$shell  = (Get-Process -Id $PID).Path
$fails  = 0
Write-Host "PowerShell $($PSVersionTable.PSVersion)"

function Check($name, $expected, $actual) {
  if ($actual -like "*$expected*") { Write-Host "  ok    $name" }
  else { Write-Host "  FAIL  $name`n        expected: $expected`n        got:`n$actual"; $script:fails++ }
}
function Run($arguments) {
  return (& $shell -NoProfile -ExecutionPolicy Bypass -File $script @arguments 2>&1 | Out-String)
}

$errors = $null
[System.Management.Automation.Language.Parser]::ParseFile($script, [ref]$null, [ref]$errors) | Out-Null
if ($errors.Count -eq 0) { Write-Host '  ok    script parses' } else { Write-Host "  FAIL  script parses: $($errors | Out-String)"; $fails++ }

Check 'unknown runtime is rejected' 'Unknown runtime' (Run @('-Runtime', 'nope'))

$listener = New-Object System.Net.Sockets.TcpListener([System.Net.IPAddress]::Loopback, 0)
$listener.Start(); $port = $listener.LocalEndpoint.Port; $listener.Stop()

Check 'doctor: no runtime' "nothing answers on port $port" (Run @('-Doctor', '-Runtime', 'other', '-Port', "$port"))

function Start-Fake($bind) {
  $p = Start-Process python -ArgumentList (Join-Path $here 'fake_runtime.py'), $bind, $port -PassThru -WindowStyle Hidden
  for ($i = 0; $i -lt 40; $i++) {
    try { Invoke-RestMethod "http://127.0.0.1:$port/v1/models" -TimeoutSec 1 | Out-Null; return $p } catch { Start-Sleep -Milliseconds 500 }
  }
  Write-Host "  FAIL  fake runtime did not start on ${bind}:$port"; $script:fails++; return $p
}

$rt = Start-Fake '127.0.0.1'
$out = Run @('-Doctor', '-Runtime', 'ollama', '-Port', "$port")
Check 'doctor: lists models' "answers on port $port with 2 model(s)" $out
Check 'doctor: localhost-only is a problem' 'only listens on 127.0.0.1' $out
Check 'doctor: gives the Ollama hint' 'setx OLLAMA_HOST 0.0.0.0' $out
Stop-Process -Id $rt.Id -Force

$rt = Start-Fake '0.0.0.0'
$out = Run @('-Doctor', '-Runtime', 'ollama', '-Port', "$port")
Check 'doctor: network bind is fine' 'listens on the network' $out
Check 'doctor: missing firewall rule is reported' 'no rule lets the mesh reach the runtime' $out
Stop-Process -Id $rt.Id -Force

if ($fails -eq 0) { Write-Host 'all tests passed'; exit 0 } else { Write-Host "$fails test(s) failed"; exit 1 }
