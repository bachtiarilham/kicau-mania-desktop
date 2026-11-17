# Verifikasi desktop dan engine H.265/MKV

16 November 2026: native libVLC 3.0.24, NVR HEVC 2304×1296 20 FPS, audio pcm_alaw asli, MKV, reconnect relay pengujian dan pembacaan rekaman setelah restart engine pengujian lolos. Codec/resolusi/FPS/audio diperiksa dengan ffprobe. Media tidak di-encode ulang.

Tes Go local-app/desktop serta pemeriksaan frontend dijalankan terpisah. Unit test logika memakai fixture metadata/HTTP, tanpa menghasilkan stream. Tidak ada mode mock runtime.

```powershell
./scripts/test-native.ps1 -NVRRTSPURL 'rtsp://127.0.0.1:8554/gantangan01' -FFmpeg 'D:/ilham/app/ffmpeg/bin/ffmpeg.exe' -MediaMTX 'D:/ilham/workdir/go/kicau-mania/local-app/bin/mediamtx.exe'
```

Sumber harus NVR nyata HEVC; environment KICAU_TEST_NVR_RTSP_URL dapat dipakai untuk menjaga credential dari riwayat command. Test tidak mengubah pengaturan NVR atau menghentikan engine operasional. Relay/recording pengujian memakai port/direktori terpisah, tanpa upload cloud.

Latency kamera-ke-layar, 24 kamera, UI multi-DPI dan provider S3/MinIO nyata belum diukur dalam verifikasi ini. Buffer live default 50 ms bukan jaminan latency nol. Tes PostgreSQL/Redis schema 011 berasal dari verifikasi sebelumnya dan tidak diulang untuk pembersihan local-app ini.

Data testing lama dipertahankan atas instruksi pengguna; tidak ada cleanup yang dijalankan. Engine operasional dihentikan setelah sesi terminal dan belum dijalankan ulang. Versi baru menolak manifest non-MKV; gunakan DATA_DIR baru bila ingin mempertahankan seluruh data lama tanpa migrasi.

Player live memakai sinkronisasi clock libVLC otomatis, tanpa memaksa clock-synchro=0. PLAYING mengikuti kemajuan frame, bukan state VLC saja. Jika frame berhenti 3 detik status STALLED; setelah 10 detik player channel disambung ulang. Startup tanpa frame dibatasi 20 detik. Panel tersembunyi/minimized dipantau melalui decode counter. Watchdog hanya untuk live; file recording tidak diulang otomatis setelah selesai. Tidak ada perubahan codec atau kontrol recording oleh pemulihan player.

## Grid lima kamera: 16 November 2026

Kelima relay operasional diperiksa read-only: HEVC dan pcm_alaw mengirim paket; sesi recording tetap RUNNING. Tes native paralel pada gantangan01–05 selama 15 detik per player lolos dengan kemajuan decode dan display pada seluruh player. Jalur live sekarang memakai output Direct3D9 dan decoder DXVA2, dengan dua thread untuk decoder software bila hardware tidak tersedia. Ini mengganti jalur decode/render lokal; tidak melakukan encode ulang atau mengubah media recording/upload.

Pada sampel sebelum perubahan, pertambahan displayed pictures adalah 190–217, lost 14–16; setelah perubahan displayed 210–295, lost 2–7. Sampel tidak berlangsung serentak sehingga angka tersebut bukan benchmark terkontrol atau bukti akar masalah driver. Tes memakai HWND tersembunyi; freeze visual pada grid WebView2, pergantian fokus, scroll dan dialog masih memerlukan verifikasi di aplikasi setelah restart desktop. UI kini menampilkan FPS berdasarkan pertambahan counter displayed libVLC, bukan estimasi dari metadata kamera; counter ini tidak mengukur latency kamera-ke-layar.

Sync panel tidak lagi memanggil resize/region Windows setiap polling bila geometry dan visibility tidak berubah. Penghentian/pemulihan player tidak menghentikan recording local-app.

```powershell
$env:KICAU_TEST_VLC = "$PWD/bin/vlc/libvlc.dll"
$env:KICAU_TEST_NVR_RTSP_URLS = 'rtsp://127.0.0.1:8554/gantangan01,rtsp://127.0.0.1:8554/gantangan02,rtsp://127.0.0.1:8554/gantangan03,rtsp://127.0.0.1:8554/gantangan04,rtsp://127.0.0.1:8554/gantangan05'
go test ./internal/player/vlc -run '^TestNativeConcurrentRTSP$' -v -count=1 -timeout 90s
```
