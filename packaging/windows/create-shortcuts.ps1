$ErrorActionPreference = 'Stop'
# Unicode shell properties preserve installation paths in every language.
$shell = New-Object -ComObject WScript.Shell
$desktop = [Environment]::GetFolderPath('DesktopDirectory')
$programs = Join-Path ([Environment]::GetFolderPath('Programs')) 'Argus'
New-Item -ItemType Directory -Force $programs | Out-Null
foreach ($path in @((Join-Path $desktop 'Argus.lnk'), (Join-Path $programs 'Argus.lnk'))) {
  $link = $shell.CreateShortcut($path)
  $link.TargetPath = $env:ComSpec
  $link.Arguments = '/d /s /c ""' + (Join-Path $PSScriptRoot 'start.bat') + '" -addr 127.0.0.1:8787"'
  $link.WorkingDirectory = $PSScriptRoot
  $link.IconLocation = (Join-Path $PSScriptRoot 'argus.exe') + ',0'
  $link.Description = 'Argus'
  $link.Save()
}
