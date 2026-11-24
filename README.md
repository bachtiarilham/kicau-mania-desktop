# Kicau Mania Desktop

Aplikasi desktop Windows amd64 berbasis **Wails v3, Go, Svelte 5, dan Tailwind CSS**. Aplikasi ini menyediakan dua antarmuka utama:
1. **Operator Console**: Kontrol taktis pertandingan, matriks 24 gantangan (Live/Rec/Standby), audio monitoring, sinkronisasi cloud session, dan quick save rekaman MKV/MP4.
2. **Monitor Dewan Juri**: Layar penilaian juri dengan grid 6-kamera (3x2), zoom tracking per kamera, rotasi otomatis 4-shift area juri, dan alert rotasi tersinkronisasi.

---

## 1. Arsitektur Sistem

```text
NVR HEVC/H.265 ── RTSP/TCP ── local-app / MediaMTX (Port 8090) ── RTSP/TCP ── desktop / libVLC
                                      │
                                      ├── Recording MKV (Stream Copy)
                                      └── Upload Multipart S3 ke Cloud
Desktop (Wails v3) ── Go Client ── HTTP/WS Local-App (8090) + HTTPS Cloud-App (8080)
```

- **Video H.265/HEVC**: Video dan audio sumber diteruskan tanpa encode ulang (*stream copy*).
- **Dekoder Native Win32 + LibVLC**: Live video dirender langsung ke child HWND native Windows melalui LibVLC C-binding (Direct3D9/DXVA2 hardware decoding), tanpa overhead WebRTC browser.
- **Isolasi Proses**: MediaMTX, FFmpeg, dan engine recording dimiliki oleh `local-app`. Menutup aplikasi desktop tidak mematikan recording atau background upload engine.

---

## 2. Struktur Kode Modular (Clean Architecture)

Mengikuti panduan utama [GEMINI.md](GEMINI.md), repositori ini menggunakan pola **Vertical Slice (Domain-First Modular Monolith)**:

```text
internal/
├── auth/                         # Domain Autentikasi Cloud
│   ├── contract.go               # Interface AuthUsecase, CloudGateway
│   ├── gateway/                  # Client HTTP ke Cloud App
│   └── usecase/                  # Logika login, token, & validasi hak akses
├── config/                       # Domain Pengaturan Desktop
│   ├── contract.go               # Interface ConfigUsecase, ConfigStore
│   ├── model/                    # Struct Config
│   ├── repository/               # Penyimpanan terenkripsi Windows DPAPI & JSON
│   └── usecase/                  # Preferensi & direktori rekaman
├── engine/                       # Domain Edge / Local-App
│   ├── contract.go               # Interface EngineUsecase, LocalEngineGateway
│   ├── gateway/                  # Client HTTP & WebSocket ke Local-App
│   └── usecase/                  # Logika pairing, session state, & metrics
├── player/                       # Domain Video & Dekoder Native
│   ├── contract.go               # Interface PlayerUsecase, PlayerDriver
│   ├── model/                    # Rect, PlayerStatus
│   ├── driver/                   # Windows HWND LibVLC manager (windows/, vlc/)
│   └── usecase/                  # Geometry player, zoom, & playback
├── recording/                    # Domain Rekaman & Ekspor
│   ├── contract.go               # Interface RecordingUsecase
│   └── usecase/                  # Ekspor video dan auto-save sesi rekaman
├── bindings/                     # Wails v3 RPC Facade (delegasi modular per domain)
├── container/                    # Dependency Injection Container
└── app/                          # Entrypoint & Wails Lifecycle

frontend/
├── src/lib/
│   ├── models/                   # Type definitions terpusat (round, session, camera, config)
│   ├── components/               # Komponen Svelte 5 Runes ($state, $derived, $effect)
│   │   ├── operator/             # Dekomposisi Operator Console
│   │   │   ├── operator-dashboard.svelte # Orchestrator (<260 baris)
│   │   │   ├── operator-header.svelte    # Tactical match control, timer, START/STOP
│   │   │   ├── grid/                     # gantangan-grid-24.svelte, gantangan-card.svelte
│   │   │   ├── spotlight/                # spotlight-view.svelte
│   │   │   └── modals/                   # quick-save-modal.svelte, round-config-modal.svelte
│   │   ├── cameras/              # native-player.svelte, zone-grid-3x2.svelte, zoom-slider.svelte
│   │   ├── juri/                 # juri-round-hud.svelte
│   │   ├── recordings/           # recordings-page.svelte
│   │   └── sessions/             # cloud-session-selector.svelte, round-timeup-dialog.svelte
│   ├── services/                 # api.ts (Wails RPC adapter), livestream.mock.ts
│   └── state/                    # round.svelte.ts, auth.svelte.ts, engine.svelte.ts
```

---

## 3. Prasyarat & Lingkungan

| Komponen | Versi / Kebutuhan |
| :--- | :--- |
| **OS** | Windows 10/11 amd64 (menggunakan Win32 HWND dan Windows DPAPI) |
| **Go** | 1.24+ atau lebih baru |
| **Node.js** | 20+ atau 24+ |
| **Package Manager** | `npm` atau `pnpm` |
| **WebView** | Microsoft Edge WebView2 Runtime |
| **VLC Runtime** | libVLC 3.0.24 amd64 (diunduh via script `scripts/fetch-deps.ps1`) |
| **Local-App** | Local Edge Media Engine berjalan pada port `8090` |

---

## 4. Alur Menjalankan Aplikasi

### A. Setup Pertama
Jalankan PowerShell dari root folder `kicau-mania-desktop`:
```powershell
# 1. Unduh dependency Go dan install dependency frontend
go mod download
cd frontend; npm install; cd ..

# 2. Unduh runtime LibVLC yang terverifikasi SHA-256
.\scripts\fetch-deps.ps1

# 3. Build pertama kali
.\scripts\build.ps1
```

### B. Menjalankan dalam Mode Development
```powershell
go run github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0 dev -config config.yml -port 9245
```

---

## 5. Pemeriksaan & Pengujian Otomatis

Verifikasi sistem dapat dijalankan melalui CLI orchestrator atau perintah standar:

```powershell
# Menggunakan agent-ctl dari repo backend (Direkomendasikan)
..\kicau-mania\bin\agent-ctl.exe verify desktop --no-smoke

# Pengujian manual Go dan Svelte
go test ./...
cd frontend
npm run check
npm test
```

---

## 6. Rujukan Peta Fitur (Feature Maps)

Sebelum menambah interaksi UI atau logika bisnis baru, selalu rujuk dokumen spesifikasi pada folder `feature-map/`:
- **`01-auth-and-session.md`**: Autentikasi multi-peran dan pairing sesi arena.
- **`02-operator-console.md`**: Konsol 24 gantangan, rotasi juri, dan simpan cepat sesi.
- **`03-juri-hud.md`**: Layar juri kuadran 3x2 dan HUD rotasi shift.
- **`04-recordings-and-whatsapp.md`**: Manajemen arsip video dan sharing WhatsApp.
- **`05-desktop-settings.md`**: Preferensi penyimpanan lokal dan integrasi local-app.
