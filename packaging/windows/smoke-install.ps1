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
$desktop = [Environment]::GetFolderPath('DesktopDirectory')
$shortcutPath = Join-Path $desktop 'Argus.lnk'
if (-not (Test-Path $shortcutPath)) { throw 'Desktop shortcut missing' }
Add-Type -Path (Join-Path $install 'ArgusShortcuts.cs')
$shortcut = [ArgusShortcuts]::Read($shortcutPath)
Write-Host "Desktop shortcut target: $($shortcut[0]); arguments: $($shortcut[1])"
if (-not (Test-Path -LiteralPath $shortcut[0])) { throw 'Desktop shortcut target does not exist' }
if (-not $shortcut[1].Contains((Join-Path $install 'start.bat'))) { throw 'Shortcut arguments lost the Unicode install path' }
$launcher = Start-Process -FilePath $shortcutPath -PassThru
$process = $null
try {
  $ready = $false
  $deadline = [DateTime]::UtcNow.AddSeconds(90)
  while ([DateTime]::UtcNow -lt $deadline) {
    $backend = Get-CimInstance Win32_Process -Filter "Name='argus.exe'" |
      Where-Object { $_.ExecutablePath -eq $exe } | Select-Object -First 1
    if ($backend) { $process = Get-Process -Id $backend.ProcessId -ErrorAction SilentlyContinue }
    try {
      $health = Invoke-WebRequest 'http://127.0.0.1:8787/api/health' -TimeoutSec 2
      $listener = Get-NetTCPConnection -LocalPort 8787 -State Listen -ErrorAction SilentlyContinue
      if ($process -and $health.StatusCode -eq 200 -and $listener.OwningProcess -contains $process.Id) {
        $ready = $true; break
      }
    } catch { }
    Start-Sleep -Seconds 1
  }
  if (-not $ready) { throw 'Installed Argus did not become ready' }
  $page = Invoke-WebRequest 'http://127.0.0.1:8787/' -TimeoutSec 5
  if ($page.StatusCode -ne 200 -or $page.Content -notmatch '<html') { throw 'Workbench did not load' }
  if (-not (Test-Path (Join-Path $env:LOCALAPPDATA 'Argus/startup.log'))) { throw 'Startup log missing from user data directory' }
  if (-not (Test-Path (Join-Path $env:LOCALAPPDATA 'Argus/data/database/binaries/bin/vcruntime140.dll'))) {
    throw 'Database app-local Visual C++ runtime missing'
  }
  Write-Host 'PASS: installed desktop shortcut launches Argus, owns the listening port, and serves the workbench without runtime downloads'
} finally {
  Get-Content (Join-Path $env:LOCALAPPDATA 'Argus/startup.log') -ErrorAction SilentlyContinue
  if ($process) { & taskkill.exe /PID $process.Id /T /F 2>$null | Out-Null }
  if ($launcher -and -not $launcher.HasExited) { & taskkill.exe /PID $launcher.Id /T /F 2>$null | Out-Null }
}
