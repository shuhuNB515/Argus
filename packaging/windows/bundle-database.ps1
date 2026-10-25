param([Parameter(Mandatory=$true)][string]$PackageRoot)
$ErrorActionPreference = 'Stop'
# Keep the version aligned with embedded-postgres V18 in go.mod's dependency.
$version = '18.3.0'
$artifact = "embedded-postgres-binaries-windows-amd64-$version"
$url = "https://repo.maven.apache.org/maven2/io/zonky/test/postgres/embedded-postgres-binaries-windows-amd64/$version/$artifact.jar"
$cache = Join-Path $PackageRoot 'database-cache'
New-Item -ItemType Directory -Force $cache | Out-Null
$temp = Join-Path ([System.IO.Path]::GetTempPath()) ([guid]::NewGuid().ToString())
New-Item -ItemType Directory $temp | Out-Null
try {
  $jar = Join-Path $temp 'postgres.zip'
  Invoke-WebRequest $url -OutFile $jar
  $checksum = Join-Path $temp 'postgres.sha256'
  Invoke-WebRequest "$url.sha256" -OutFile $checksum
  $expected = (Get-Content $checksum -Raw).Trim().ToLowerInvariant()
  $actual = (Get-FileHash $jar -Algorithm SHA256).Hash.ToLowerInvariant()
  if ($actual -ne $expected) { throw 'Database archive SHA256 mismatch' }
  Expand-Archive $jar -DestinationPath (Join-Path $temp 'unpacked')
  $archive = @(Get-ChildItem (Join-Path $temp 'unpacked') -Recurse -Filter *.txz)
  if ($archive.Count -ne 1) { throw 'Expected one database runtime archive' }
  # PostgreSQL depends on Microsoft's CRT. Include the redistributable app-local
  # DLLs from the Windows build toolchain so a clean PC needs no extra installer.
  $vswhere = Join-Path ${env:ProgramFiles(x86)} 'Microsoft Visual Studio/Installer/vswhere.exe'
  $vs = & $vswhere -latest -products '*' -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath
  if (-not $vs) { throw 'Visual C++ redistributable toolchain missing' }
  $crt = Get-ChildItem (Join-Path $vs 'VC/Redist/MSVC/*/x64/Microsoft.VC*.CRT') -Directory |
    Sort-Object FullName -Descending | Select-Object -First 1
  if (-not $crt) { throw 'Visual C++ x64 app-local runtime missing' }
  $runtime = Join-Path $temp 'runtime'
  New-Item -ItemType Directory $runtime | Out-Null
  & tar.exe -xf $archive[0].FullName -C $runtime
  if ($LASTEXITCODE -ne 0) { throw 'Database runtime extraction failed' }
  Copy-Item (Join-Path $crt.FullName '*.dll') (Join-Path $runtime 'bin')
  if (-not (Test-Path (Join-Path $runtime 'bin/vcruntime140.dll'))) { throw 'App-local VCRUNTIME140.dll missing' }
  & tar.exe -cJf (Join-Path $cache "$artifact.txz") -C $runtime .
  if ($LASTEXITCODE -ne 0) { throw 'Offline database archive creation failed' }
} finally {
  Remove-Item $temp -Recurse -Force
}
