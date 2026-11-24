# Kicau Mania Desktop - AI Agent & Developer Guidelines

Dokumen ini adalah aturan utama (*Project Rules*) yang **wajib dipatuhi secara otomatis** oleh AI Agent dan developer saat bekerja di repositori desktop ini.

---

## 1. Aturan Arsitektur Go Backend (Vertical Slice Domain-First Clean Architecture)
- **Pola Modul Domain:** Wajib menggunakan pola *Vertical Slice (Domain-First Modular Monolith)* berbasis Clean Architecture:
  ```
  internal/<domain>/{model, usecase, repository, gateway, delivery}
  ```
- **Larangan Keras Horizontal Layering:**
  - Dilarang keras mengelompokkan kode berdasarkan layer teknis horizontal seperti `internal/usecase/`, `internal/domain/`, `internal/model/`, `internal/bindings/`, `internal/settings/`, atau `internal/localapp/` yang memecah satu domain ke banyak folder.
  - Setiap domain (misal: `auth`, `config`, `engine`, `player`, `recording`) wajib mandiri dan memiliki sub-paket sesuai kebutuhannya:
    - `model`: Struktur data, entitas, dan konstanta domain.
    - `usecase`: Business logic domain (tanpa ketergantungan pada Wails/HTTP/UI).
    - `repository` atau `gateway`: Driver penyimpanan lokal (registry/JSON) atau adapter client external (Cloud App / Edge Local-App).
    - `delivery`: Handler Wails RPC untuk domain tersebut.
- **Composition Root & Wiring:** Hanya `internal/container` (DI wiring) dan `internal/app` (lifecycle Wails) yang diizinkan mengimpor concrete implementations dari berbagai domain.

---

## 2. Aturan Arsitektur Frontend & State Management (Svelte 5)
- **Svelte 5 Runes:** Seluruh state reaktif wajib menggunakan Runes murni Svelte 5 (`$state`, `$derived`, `$effect`, `$props`). Dilarang mencampuradukkan sintaks Svelte 3/4 seperti `let x = ...` untuk reaktivitas atau `$: ...`.
- **Pola Modul Domain:** Kelompokkan komponen berbasis domain fitur di dalam `frontend/src/lib/components/{auth, cameras, juri, operator, recordings, sessions}`.
- **Pemisahan Layer:**
  - `lib/models/`: Struktur data dan tipe TypeScript murni terpusat.
  - `lib/services/`: Pemanggilan API HTTP / Wails Go bindings. Dilarang meletakkan logika render UI di layer service.
  - `lib/state/`: State global tersinkronisasi (`round.svelte.ts`, `auth.svelte.ts`, `engine.svelte.ts`).
  - `lib/components/`: Presentational & interactive UI (decomposed menjadi sub-komponen single-responsibility, hindari fat component >250 baris).
- **Sinkronisasi Multi-Window:** Seluruh perubahan timer ronde dan rotasi juri wajib disinkronkan melalui `BroadcastChannel("kicau_round_sync_channel")`.

---

## 3. Standar Zero-Comment & Executable Constraints (`no-code-comments`)
Dilarang mengandalkan komentar agar developer atau AI mengingat suatu aturan:
1. **Jika ada batasan (constraint):** Masukkan ke dalam API atau Type System TypeScript (misal: strict enum/union types, interface kontrak, Zod validation jika diperlukan).
2. **Jika ada konvensi tim:** Tegakkan melalui linter ESLint/Svelte check (`npm run check`).
3. **Jika ada aturan bisnis:** Buktikan melalui automated unit tests (`npm test` atau Go unit tests). Test adalah dokumentasi bisnis yang tidak bisa berbohong.
4. **Pengecualian:** Hanya izinkan dokumentasi JSDoc pada fungsi library publik yang diekspor.

---

## 4. Protokol Eliminasi Workaround (`eliminate-workarounds`)
- **Dilarang keras:**
  - Menulis komentar `/* ... */` untuk mengarsipkan kode lama. Hapus kodenya langsung; Git history adalah tempat arsip.
  - Membiarkan komponen usang hidup berdampingan dengan komponen baru.
  - Membuat *wrapper* baru di atas workaround lama.
  - Mempertahankan kode hack dengan komentar pembelaan.
- **Boy Scout Rule:** Setiap kali membaca kode dan menemukan workaround atau duplikasi, bersihkan dan selesaikan langsung pada akarnya.

---

## 5. Protokol Operasional & Verifikasi Agen (`agent-ctl`)
Binary pengontrol aplikasi terletak di repositori backend: `../kicau-mania/bin/agent-ctl.exe` (atau `D:\Workdir\Go\kicau-mania\bin\agent-ctl.exe`).
AI Agent wajib mematuhi alur kerja berikut:
1. **Cek Status:** Jalankan `agent-ctl status` sebelum dan sesudah perubahan.
2. **Eksekusi Pengujian:** Jalankan `agent-ctl verify desktop --no-smoke` setiap kali selesai mengubah kode.
3. **Bukti Kerja (Evidence):** Dilarang melaporkan pekerjaan selesai tanpa melampirkan ringkasan bukti kerja dari `agent-ctl evidence desktop`.

---

## 6. Rujukan Peta Fitur (Feature Map)
Sebelum mengubah atau menambahkan interaksi UI, selalu konsultasikan dokumen spesifikasi pada folder `feature-map/`:
- `feature-map/01-auth-and-session.md`
- `feature-map/02-operator-console.md`
- `feature-map/03-juri-hud.md`
- `feature-map/04-recordings-and-whatsapp.md`
- `feature-map/05-desktop-settings.md`
