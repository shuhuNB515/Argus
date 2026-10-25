param([Parameter(Mandatory=$true)][string]$Installer)
$ErrorActionPreference = 'Stop'
$env:ARGUS_POSTGRES_CACHE = $null
$root = Join-Path $env:RUNNER_TEMP ('Argus 中文安装 ' + [guid]::NewGuid().ToString())
# Keep data under the Windows user profile, as on customer machines; the app
# itself remains installed into a Unicode directory on the runner's work drive.
$env:LOCALAPPDATA = Join-Path ([System.IO.Path]::GetTempPath()) ('Argus 用户数据 ' + [guid]::NewGuid().ToString())
$install = Join-Path $root 'Application'
$env:HTTP_PROXY = 'http://127.0.0.1:1'
$env:HTTPS_PROXY = 'http://127.0.0.1:1'
$env:NO_PROXY = '127.0.0.1,localhost'
$setup = Start-Process -FilePath $Installer -ArgumentList @('/S', "/D=$install") -Wait -PassThru
if ($setup.ExitCode -ne 0) { throw "Installer failed: $($setup.ExitCode)" }
$exe = Join-Path $install 'argus.exe'
if (-not (Test-Path $exe)) { throw 'Installed executable missing' }
if (-not (Test-Path (Join-Path $install 'database-cache/embedded-postgres-binaries-windows-amd64-18.3.0.txz'))) {
  throw 'Installed offline database runtime missing'
}
$process = Start-Process $exe -ArgumentList @('-open-browser=false', '-proxy=') -PassThru
try {
  $ready = $false
  for ($i = 0; $i -lt 120; $i++) {
    $process.Refresh()
    if ($process.HasExited) { throw "Argus exited: $($process.ExitCode)" }
    try {
      $health = Invoke-WebRequest 'http://127.0.0.1:8787/api/health' -TimeoutSec 2
      if ($health.StatusCode -eq 200) { $ready = $true; break }
    } catch { }
    Start-Sleep -Seconds 1
  }
  if (-not $ready) { throw 'Installed Argus did not become ready' }
  $page = Invoke-WebRequest 'http://127.0.0.1:8787/' -TimeoutSec 5
  if ($page.StatusCode -ne 200 -or $page.Content -notmatch '<html') { throw 'Workbench did not load' }
  if (-not (Test-Path (Join-Path $env:LOCALAPPDATA 'Argus/startup.log'))) { throw 'Startup log missing from user data directory' }
  Write-Host 'PASS: installed Windows application serves health and workbench without runtime downloads'
} finally {
  Get-Content (Join-Path $env:LOCALAPPDATA 'Argus/startup.log') -ErrorAction SilentlyContinue
  & taskkill.exe /PID $process.Id /T /F 2>$null | Out-Null
}
