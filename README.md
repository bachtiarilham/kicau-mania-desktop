# Kicau Mania Desktop

Aplikasi CCTV Windows amd64 berbasis Wails, Go, Svelte dan shadcn-svelte. Fitur desktop mencakup setup NVR, grid kamera, metrik audio, kontrol recording, mapping ronde/peserta, rekaman lokal dan upload cloud. CRUD lomba dikelola melalui aplikasi web.

Dokumen ini ditujukan untuk developer yang baru mengambil source proyek.

## Arsitektur singkat

```text
NVR HEVC/H.265 ── RTSP/TCP ── local-app / MediaMTX ── RTSP/TCP ── desktop / libVLC
                                      │
                                      ├── Recording MKV (stream copy)
                                      └── Upload object storage melalui cloud-app
Desktop ── Go client ── HTTP/WebSocket local-app + HTTPS cloud-app
```

MediaMTX, FFmpeg, file recording dan worker upload dimiliki **local-app**, bukan desktop. Local-app harus berjalan di komputer Windows yang sama. Menutup desktop tidak menghentikan recording atau upload engine.

Video harus **HEVC/H.265**; video dan audio sumber diteruskan tanpa encode ulang. Recording baru menggunakan **MKV**. Decode libVLC untuk layar serta decode PCM untuk metrik audio tetap diperlukan. WebRTC/WHEP tidak dipakai. Live Windows memakai panel HWND, Direct3D9/DXVA2 dan buffer default 50 ms; nilai buffer bukan ukuran latency kamera-ke-layar.

## Prasyarat

| Komponen | Kebutuhan |
| --- | --- |
| OS | Windows amd64; implementasi HWND dan DPAPI khusus Windows |
| Go | 1.26 atau lebih baru, mengikuti `go.mod` |
| Node.js | 24 atau lebih baru |
| Package manager | pnpm tersedia di PATH |
| WebView | Microsoft Edge WebView2 Runtime |
| Wails | 3.1.0, dipanggil melalui `go run` oleh script; tidak wajib instal CLI global |
| VLC | libVLC 3.0.24 amd64, diunduh dan diverifikasi oleh script |
| Engine | Source/binary local-app yang kompatibel, FFmpeg, ffprobe dan MediaMTX |
| Sumber live | NVR/kamera nyata dengan profil HEVC/H.265 dan akses RTSP |

Versi dependency dikunci melalui `go.mod`, `go.sum`, `frontend/pnpm-lock.yaml` dan `third-party/vlc/manifest.json`. Commit perubahan lock file bersama perubahan dependency. Jangan mengedit runtime VLC hasil unduhan.

## Setup pertama

Jalankan PowerShell dari root checkout **kicau-mania-desktop**; semua perintah di bawah memakai direktori tersebut kecuali disebutkan lain.

```powershell
go version
node --version
pnpm --version

go mod download
pnpm --dir frontend install --frozen-lockfile
.\scripts\fetch-deps.ps1
.\scripts\build.ps1
.\bin\kicau-mania-desktop.exe
```

`fetch-deps.ps1` mengunduh VLC sesuai manifest dan memeriksa SHA-256 sebelum ekstraksi. `build.ps1` memasang dependency frontend, memeriksa tipe, membangun frontend, membuat resource Windows, membangun executable dan menyalin runtime VLC ke `bin/vlc`.

Build frontend perlu dilakukan sebelum menjalankan tes/build Go dari checkout baru karena `main.go` menyematkan `frontend/dist` melalui `go:embed`. Jangan commit folder `dist` untuk mengatasi error embed.

### Menjalankan local-app

Local-app berada di repository/backend Kicau Mania dan tidak disalin ke proyek desktop. Dari folder `local-app`:

```powershell
go run ./cmd/api -env .env.portable.example
```

Periksa dan sesuaikan file environment lokal sebelum menjalankannya: path binary FFmpeg/ffprobe/MediaMTX, `DATA_DIR`, `RUNTIME_DIR`, interface LAN dan sumber kamera. Gunakan `ENGINE_ACCESS_MODE=lan`. Jangan commit kredensial RTSP atau environment pribadi.

Default koneksi desktop:

| Pengaturan | Default |
| --- | --- |
| Local API | `http://127.0.0.1:8090` |
| Relay RTSP | `rtsp://127.0.0.1:8554/{channel_id}` |
| Cloud API | `https://dev.megaallstararena.com/api/v1` |
| Runtime VLC | `vlc/libvlc.dll` di sebelah executable |

Desktop tidak otomatis memulai engine. Port 8090 dan 8554 harus tersedia. Cloud diperlukan untuk login, pairing, mapping dan upload; setup/live/recording lokal dikelola oleh engine.

## Mode development

Setelah build pertama dan local-app tersedia:

```powershell
go run github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0 dev -config config.yml -port 9245
```

Konfigurasi dev menjalankan Vite pada loopback port **9245**, membangun Go dengan tag `dev`, dan menjalankan executable di `bin`. Frontend memakai Vite/Svelte, bukan SvelteKit. Jangan menjalankan Vite kedua pada port yang sama. `pnpm --dir frontend dev` saja tidak menyediakan binding Go, sehingga bukan pengganti menjalankan Wails.

Setelah mengubah method Go yang diekspos ke frontend:

```powershell
go run github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0 generate bindings -ts
```

Periksa perubahan `frontend/bindings` dan pemanggilnya pada `frontend/src/lib/services/api.ts`. Binding tidak diabaikan Git. Go Task opsional: `task dev`, `task build`, `task package`, `task bindings`, dan `task check` tersedia melalui `Taskfile.yml`.

## Struktur source

```text
main.go                      Entry point dan embedded frontend
internal/
  app/                       Lifecycle aplikasi Wails
  bindings/                  API Go yang dipanggil UI
  model/                     DTO dan konfigurasi desktop
  localapp/                  Client kontrol engine dan metrik
  cloud/                     Auth dan API cloud
  player/
    vlc/                     LibVLC, pemantauan frame dan playback
    windows/                 HWND, posisi, clipping dan DPI
  settings/                  Preferensi dan enkripsi DPAPI
frontend/
  bindings/                  Binding Wails hasil generate
  src/lib/
    components/              UI per domain dan komponen shadcn
    services/                Pemanggilan binding/API
    state/                   State engine, sesi dan metrik
build/windows/               Source resource dan manifest Windows
scripts/                     Build, dependency, package dan tes native
third-party/vlc/             Manifest, checksum dan dokumentasi dependency
docs/                        Arsitektur, kontrak cloud dan hasil verifikasi
```

Pisahkan UI per domain; kontrol engine dan token tetap melalui Go. Jangan meneruskan token/credential ke page state atau localStorage. Hormati status engine: penerimaan request HTTP 202 belum berarti recording atau finalisasi selesai.

## Konfigurasi developer

Gunakan menu **Pengaturan desktop**. Pengaturan berlaku setelah desktop dibuka ulang dan disimpan pada:

```text
%LOCALAPPDATA%\KicauManiaCCTV\desktop.json
```

`desktop.example.json` adalah referensi format, **tidak dimuat otomatis dari root checkout**. Jika menyalinnya ke lokasi di atas, sesuaikan `vlc` menjadi path absolut `bin/vlc/libvlc.dll` pada komputer sendiri. Tanpa file preferensi, aplikasi memakai default berdasarkan lokasi executable.

Sesi cloud dilindungi DPAPI akun Windows yang sama. Konfigurasi NVR, pairing dan credential engine berada di `DATA_DIR` milik local-app. Jangan menyalin secret DPAPI antar akun atau menghapus folder data untuk mengatasi kegagalan koneksi.

Untuk fitur cloud, gunakan ADMIN/OPERATOR arena yang berwenang. ADMIN melakukan pairing engine; mapping sesi/channel/ronde/tiket harus dikonfirmasi sebelum upload. Mapping tidak tersedia tetap memungkinkan saving lokal. Backend harus mendukung kontrak media MKV dan object storage privat. Detail endpoint, scope, idempotensi dan deployment tersedia dalam [kontrak cloud](docs/cloud-media-contract.md); desktop tidak menjalankan migrasi database.

## Pemeriksaan dan pengujian

Dari checkout baru, selesaikan build frontend terlebih dahulu. Pemeriksaan rutin:

```powershell
go test ./...
pnpm --dir frontend check
pnpm --dir frontend test
pnpm --dir frontend build
```

Tes native membutuhkan NVR HEVC nyata dan bersifat **opt-in**; tanpa environment terkait, tes tersebut dilewati. Tes unit logika/HTTP tidak membuktikan playback kamera nyata atau latency. Tidak ada generator stream sintetis atau mode mock runtime.

Contoh tes native dan saving, dengan path yang disesuaikan untuk komputer developer:

```powershell
$env:KICAU_TEST_NVR_RTSP_URL = 'rtsp://127.0.0.1:8554/gantangan01'
.\scripts\test-native.ps1 `
  -FFmpeg 'C:\tools\ffmpeg\bin\ffmpeg.exe' `
  -MediaMTX 'C:\src\kicau-mania\local-app\bin\mediamtx.exe' `
  -LocalAppRoot 'C:\src\kicau-mania\local-app'
```

Script memerlukan `ffprobe.exe` di folder yang sama dengan FFmpeg. Berikan `-LocalAppRoot` secara eksplisit karena default script masih menunjuk path mesin pengembang awal. Pilih sumber relay yang aktif atau URL NVR pengujian. Jangan membagikan URL berisi password dalam log atau commit.

Untuk tes beberapa player sekaligus, gunakan `KICAU_TEST_VLC` serta daftar URL dipisahkan koma di `KICAU_TEST_NVR_RTSP_URLS`, lalu jalankan `go test ./internal/player/vlc -run '^TestNativeConcurrentRTSP$' -v -count=1 -timeout 90s`. Detail dan batas hasil pengujian tersedia dalam [acceptance](docs/acceptance.md).

## Build dan distribusi

```powershell
# Build lengkap
.\scripts\build.ps1

# Hanya Go; frontend/dist harus sudah dibangun dari source terbaru
.\scripts\build.ps1 -SkipFrontend

# Build lengkap dan ZIP portable
.\scripts\package.ps1
```

Hasil:

- `bin/kicau-mania-desktop.exe`
- `bin/vlc/` — runtime dan plugin VLC, wajib berada di sebelah executable.
- `bin/kicau-mania-desktop-windows-amd64.zip` — paket portable.

Penerima paket tetap memerlukan WebView2 dan local-app terkonfigurasi. Ikuti informasi distribusi/lisensi di `third-party/vlc/README.md`. Build tidak menyalin engine, mengubah VPS, mereset database, atau menghapus rekaman.

## Troubleshooting

| Gejala | Pemeriksaan |
| --- | --- |
| Go gagal menemukan `frontend/dist` | Jalankan install dan build frontend sebelum tes/build Go |
| `libvlc.dll` tidak ditemukan | Jalankan fetch/build, periksa `bin/vlc` dan path absolut pada preferensi |
| API lokal menolak koneksi | Pastikan local-app berjalan dan local API sesuai port engine |
| `NETWORK_UNAVAILABLE` | Pilih interface LAN dengan IP komputer yang sekarang, lalu terapkan konfigurasi |
| RTSP timeout | Periksa daya/PoE, kabel, IP, port dan path kamera; uji koneksi dari komputer engine |
| Kredensial RTSP ditolak | Isi akun RTSP pada mapping kamera; akun ONVIF dapat berbeda |
| Disk di bawah minimum | Periksa ruang pada disk `DATA_DIR`; default engine mensyaratkan 5 GiB bebas |
| Live tersendat | Periksa FPS/status frame, jaringan dan GOP kamera; coba buffer 100 ms lalu restart desktop |
| Upload belum berhasil | Bedakan saving lokal, transfer objek dan sinkronisasi metadata; periksa pairing, mapping dan kontrak MKV cloud |

Stop sesi dan tunggu finalisasi sebelum mengelola rekaman. Jangan menghapus konfigurasi, credential, queue atau journal untuk memperbaiki retry. Perubahan buffer/player hanya memerlukan restart desktop; penggantian binary/config engine mengikuti lifecycle local-app.

## Sebelum commit

- Sertakan source, lock file, manifest/checksum dependency, binding yang berubah dan dokumentasi terkait.
- Jangan sertakan `bin`, `node_modules`, `dist`, runtime VLC unduhan, `.env` pribadi, token atau rekaman.
- Jalankan pemeriksaan yang relevan; laporkan tes native/cloud yang belum dijalankan.
- Periksa `git status` dan diff sebelum commit. `.gitignore` tidak menghapus file yang telanjur tracked.

Bacaan lanjutan: [arsitektur](docs/architecture.md), [kontrak cloud](docs/cloud-media-contract.md), [acceptance](docs/acceptance.md).
