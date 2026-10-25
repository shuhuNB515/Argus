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
  Copy-Item $archive[0].FullName (Join-Path $cache "$artifact.txz")
} finally {
  Remove-Item $temp -Recurse -Force
}
