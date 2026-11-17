$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
& (Join-Path $PSScriptRoot 'build.ps1')
$file=Join-Path $root 'bin\kicau-mania-desktop-windows-amd64.zip'
Compress-Archive -LiteralPath (Join-Path $root 'bin\kicau-mania-desktop.exe'),(Join-Path $root 'bin\vlc'),(Join-Path $root 'bin\README.md') -DestinationPath $file -Force
Write-Output $file
