$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$dir=Join-Path $root 'third-party\vlc'
$manifest=Get-Content -LiteralPath (Join-Path $dir 'manifest.json') -Raw | ConvertFrom-Json
$archive=Join-Path $dir 'vlc.zip'
if(!(Test-Path -LiteralPath $archive)){Invoke-WebRequest -Uri $manifest.url -OutFile $archive}
if((Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant() -ne $manifest.sha256){throw 'Checksum arsip VLC tidak cocok. Arsip tidak diekstrak.'}
$destination=Join-Path $dir 'runtime'
Expand-Archive -LiteralPath $archive -DestinationPath $destination -Force
Write-Output "VLC terverifikasi: $destination"
