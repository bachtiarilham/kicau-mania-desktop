param([switch]$SkipFrontend,[switch]$Dev)
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
Push-Location $root
try {
 if(!$SkipFrontend){pnpm --dir frontend install --frozen-lockfile;if($LASTEXITCODE -ne 0){throw 'Install gagal'};pnpm --dir frontend check;if($LASTEXITCODE -ne 0){throw 'Pemeriksaan frontend gagal'};pnpm --dir frontend build;if($LASTEXITCODE -ne 0){throw 'Build frontend gagal'}}
 New-Item -ItemType Directory -Force -Path bin | Out-Null
 go run build/windows/icon.go
 if($LASTEXITCODE -ne 0){throw 'Pembuatan icon gagal'}
 go run github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0 generate syso -arch amd64 -icon build/windows/icon.ico -manifest build/windows/app.manifest -out wails_windows_amd64.syso
 if($LASTEXITCODE -ne 0){throw 'Windows resource gagal'}
 if($Dev){go build -tags dev -o bin/kicau-mania-desktop.exe .}else{go build -tags production -trimpath -ldflags '-H windowsgui' -o bin/kicau-mania-desktop.exe .}
 if($LASTEXITCODE -ne 0){throw 'Build Go gagal'}
 $vlc=Join-Path $root 'third-party\vlc\runtime\vlc-3.0.24'
 if(!(Test-Path -LiteralPath (Join-Path $vlc 'libvlc.dll'))){& (Join-Path $PSScriptRoot 'fetch-deps.ps1')}
 New-Item -ItemType Directory -Force -Path bin/vlc | Out-Null
 # Loaded VLC plugins may be locked by an existing desktop. Do not rewrite
 # identical dependencies on every frontend build.
 foreach($source in @(Get-ChildItem -LiteralPath $vlc -Recurse -File)){
  $relative=$source.FullName.Substring($vlc.Length).TrimStart('\', '/')
  $destination=Join-Path (Join-Path $root 'bin/vlc') $relative
  New-Item -ItemType Directory -Force -Path ([IO.Path]::GetDirectoryName($destination)) | Out-Null
  if((Test-Path -LiteralPath $destination) -and (Get-FileHash -LiteralPath $source.FullName -Algorithm SHA256).Hash -eq (Get-FileHash -LiteralPath $destination -Algorithm SHA256).Hash){continue}
  Copy-Item -LiteralPath $source.FullName -Destination $destination -Force
 }
 Copy-Item -LiteralPath README.md -Destination bin/README.md
 Copy-Item -LiteralPath third-party/vlc/manifest.json -Destination bin/vlc/manifest.json
 Write-Output "Build tersedia: $(Join-Path $root 'bin\kicau-mania-desktop.exe')"
} finally {Pop-Location}
