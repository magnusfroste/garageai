# Installs GarageAI Bridge on Windows: the `garageai` command, and (with -Connect) connects this
# computer. Run in PowerShell as Administrator:
#
#   irm https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.ps1 | iex
#
# or, to install and connect in one go:
#
#   & ([scriptblock]::Create((irm https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.ps1))) -Connect
#
# One binary, nothing else. Downloaded from the release and checked against its SHA256SUMS. The
# command goes to %LOCALAPPDATA%\Programs\GarageAI (added to your PATH); `garageai connect` copies
# it to %ProgramData%\GarageAI for the services, which run as SYSTEM.
#   $env:GARAGEAI_BRIDGE_VERSION = "bridge-v0.7.0"   a specific release instead of the newest
param([switch]$Connect)
$ErrorActionPreference = "Stop"
$repo = if ($env:GARAGEAI_BRIDGE_REPO) { $env:GARAGEAI_BRIDGE_REPO } else { "magnusfroste/garageai" }
$version = if ($env:GARAGEAI_BRIDGE_VERSION) { $env:GARAGEAI_BRIDGE_VERSION } else { "latest" }
$arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq "Arm64") { "arm64" } else { "amd64" }

if ($version -eq "latest" -and -not $env:GARAGEAI_BRIDGE_DOWNLOAD_BASE) {
  $releases = Invoke-RestMethod "https://api.github.com/repos/$repo/releases?per_page=50"
  $tags = $releases | ForEach-Object { $_.tag_name } | Where-Object { $_ -like "bridge-v*" }
  $stable = $tags | Where-Object { $_ -notmatch "-" } | Sort-Object { [version]($_ -replace "^bridge-v", "") } -Descending
  $version = if ($stable) { $stable[0] } else { $tags[0] }
  if (-not $version) { throw "garageai: could not find a Bridge release in $repo (is github.com reachable?)" }
}
$base = if ($env:GARAGEAI_BRIDGE_DOWNLOAD_BASE) { $env:GARAGEAI_BRIDGE_DOWNLOAD_BASE } else { "https://github.com/$repo/releases/download/$version" }
$asset = "garageai-windows-$arch.exe"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("garageai-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
  Write-Host "Downloading GarageAI Bridge $version for windows/$arch..."
  Invoke-WebRequest "$base/$asset" -OutFile (Join-Path $tmp $asset)
  Invoke-WebRequest "$base/SHA256SUMS" -OutFile (Join-Path $tmp "SHA256SUMS")
  $want = (Get-Content (Join-Path $tmp "SHA256SUMS") | Where-Object { $_ -match "\s\*?$([regex]::Escape($asset))$" }) -replace "\s.*$", ""
  $have = (Get-FileHash (Join-Path $tmp $asset) -Algorithm SHA256).Hash.ToLower()
  if (-not $want -or $want.ToLower() -ne $have) { throw "garageai: checksum mismatch for $asset, not installing" }

  $dir = Join-Path $env:LOCALAPPDATA "Programs\GarageAI"
  New-Item -ItemType Directory -Force -Path $dir | Out-Null
  $dest = Join-Path $dir "garageai.exe"
  Move-Item -Force (Join-Path $tmp $asset) $dest
  Unblock-File $dest -ErrorAction SilentlyContinue  # drop the download mark, so a later double-click does not hit SmartScreen
  $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
  if (-not ($userPath -split ";" | Where-Object { $_ -eq $dir })) {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$dir", "User")
    $env:Path = "$env:Path;$dir"
    Write-Host "Added $dir to your PATH (new terminals see it)."
  }
  # Windows Defender sometimes quarantines an unsigned Go program as a false positive, right after
  # it is written. Say so plainly, with the fix, instead of "garageai is not recognized".
  $ver = $null
  try { if (Test-Path $dest) { $ver = & $dest version 2>$null } } catch { $ver = $null }
  if (-not $ver) {
    Write-Host ""
    Write-Host "garageai.exe was removed or cannot start. This is usually Windows Defender (a false positive on an unsigned program)."
    Write-Host "Check: Windows Security -> Virus & threat protection -> Protection history."
    Write-Host "Fix:   Windows Security -> Virus & threat protection -> Manage settings -> Exclusions -> Add folder:"
    Write-Host "         $dir"
    Write-Host "         $env:ProgramData\GarageAI"
    Write-Host "       then run this command again. Report it so it stops: https://www.microsoft.com/wdsi/filesubmission"
    throw "garageai: the installed program does not run (see above)"
  }
  Write-Host "Installed $ver -> $dest (checksum verified)"

  # A garage already connected: update the services' copy too.
  $svc = Join-Path $env:ProgramData "GarageAI\garageai.exe"
  if (Test-Path $svc) {
    Write-Host "Updating the services (needs Administrator)..."
    & $dest install-service
  }
  if ($Connect) {
    & $dest connect
  } else {
    Write-Host ""
    if (Test-Path $svc) {
      Write-Host "This computer is a garage already: nothing more to do. Check it with: garageai doctor"
    } else {
      Write-Host "Not connected yet? Run: garageai connect   (Windows asks for Administrator; answer Yes)"
      Write-Host "What runs here and what to fix: garageai doctor"
    }
  }
} finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
