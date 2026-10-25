$ErrorActionPreference = 'Stop'
# Unicode shell properties preserve installation paths in every language.
Add-Type -Path (Join-Path $PSScriptRoot 'ArgusShortcuts.cs')
$desktop = [Environment]::GetFolderPath('DesktopDirectory')
$programs = Join-Path ([Environment]::GetFolderPath('Programs')) 'Argus'
New-Item -ItemType Directory -Force $programs | Out-Null
foreach ($path in @((Join-Path $desktop 'Argus.lnk'), (Join-Path $programs 'Argus.lnk'))) {
  $arguments = '/d /s /c ""' + (Join-Path $PSScriptRoot 'start.bat') + '" -addr 127.0.0.1:8787"'
  [ArgusShortcuts]::Create($path, $env:ComSpec, $arguments, $PSScriptRoot, (Join-Path $PSScriptRoot 'argus.exe'))
  $saved = [ArgusShortcuts]::Read($path)
  if ($saved[1] -cne $arguments) { throw 'Unicode shortcut arguments were not preserved' }
}
