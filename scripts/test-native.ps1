param(
 [Parameter(Mandatory)][string]$FFmpeg,
 [Parameter(Mandatory)][string]$MediaMTX,
 [string]$NVRRTSPURL=$env:KICAU_TEST_NVR_RTSP_URL,
 [string]$LocalAppRoot="D:\ilham\workdir\go\kicau-mania\local-app"
)
$ErrorActionPreference='Stop'
if([string]::IsNullOrWhiteSpace($NVRRTSPURL)){throw 'Set KICAU_TEST_NVR_RTSP_URL ke stream NVR nyata H.265. Tidak ada sumber pengganti.'}
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$vlc=Join-Path $root 'bin\vlc\libvlc.dll'
if(!(Test-Path -LiteralPath $vlc)){$vlc=Join-Path $root 'third-party\vlc\runtime\vlc-3.0.24\libvlc.dll'}
if(!(Test-Path -LiteralPath $vlc)){throw 'Jalankan scripts/fetch-deps.ps1 terlebih dahulu.'}
$probe=Join-Path ([IO.Path]::GetDirectoryName($FFmpeg)) 'ffprobe.exe'
$raw=& $probe -v error -rtsp_transport tcp -timeout 8000000 -read_intervals '%+1' -show_streams -of json $NVRRTSPURL 2>$null
if($LASTEXITCODE -ne 0){throw 'Probe NVR gagal; periksa koneksi dan credential.'}
$media=$raw | ConvertFrom-Json
$video=@($media.streams | Where-Object codec_type -eq 'video')
if($video.Count -ne 1 -or $video[0].codec_name -ne 'hevc'){throw 'Pengujian hanya menerima stream NVR HEVC/H.265.'}
$old=@{}
foreach($key in @('KICAU_TEST_VLC','KICAU_TEST_NVR_RTSP_URL','KICAU_TEST_FFMPEG','KICAU_TEST_FFPROBE','KICAU_TEST_MEDIAMTX')){$old[$key]=[Environment]::GetEnvironmentVariable($key,'Process')}
try {
 $env:KICAU_TEST_VLC=$vlc;$env:KICAU_TEST_NVR_RTSP_URL=$NVRRTSPURL
 $env:KICAU_TEST_FFMPEG=$FFmpeg;$env:KICAU_TEST_FFPROBE=$probe;$env:KICAU_TEST_MEDIAMTX=$MediaMTX
 Push-Location $root
 try{go test ./internal/player/vlc -run TestNativeRTSPEmbeddingAndCleanup -count=1 -timeout 60s -v;if($LASTEXITCODE -ne 0){throw 'Native NVR acceptance gagal'}}finally{Pop-Location}
 Push-Location $LocalAppRoot
 try{go test ./internal/app -run TestDesktopRealSaving -count=1 -timeout 90s -v;if($LASTEXITCODE -ne 0){throw 'Saving NVR acceptance gagal'}}finally{Pop-Location}
} finally {
 foreach($key in $old.Keys){[Environment]::SetEnvironmentVariable($key,$old[$key],'Process')}
}
