# Feature Map - Kicau Mania Desktop (Wails + Svelte)

Direktori ini berisi katalog interaktif dan peta fitur (*Feature Map*) lengkap untuk aplikasi **Kicau Mania Desktop**.
Peta ini dirancang untuk developer, QA tester, dan AI Agent (khususnya *Browser / UI Automation Agent*) agar dapat menavigasi, berinteraksi, dan menguji seluruh antarmuka secara deterministik.

---

## Ringkasan Peran & Akses Antarmuka (Role-Based Access)

| Role | Tab Utama yang Dapat Diakses | Deskripsi Peran |
| :--- | :--- | :--- |
| **`juri`** | Monitor Juri (`cameras`) | Layar khusus dewan juri dengan grid kamera 3x2, penugasan zona A-D otomatis, dan timer shift. |
| **`operator`** | Operator Arena (`operator`), Monitor Juri (`cameras`), Rekaman & Arsip (`recordings`) | Operator teknis arena yang mengontrol ronde, start/stop recording, alert rotasi juri, dan distribusi WA. |
| **`admin`** | Seluruh Tab + Pengaturan Desktop (`settings`) | Akses penuh termasuk konfigurasi NVR, kalibrasi audio, dan endpoint cloud. |

---

## Indeks Dokumen Fitur

| File | Nama Fitur Utama | Deskripsi Singkat |
| :--- | :--- | :--- |
| [01-auth-and-session.md](./01-auth-and-session.md) | Autentikasi & Pemilihan Sesi | Layar login role (Juri/Operator/Admin) dan picker sesi lomba sebelum recording. |
| [02-operator-console.md](./02-operator-console.md) | Konsol Operator Arena (Live 24) | Kontrol ronde, multi-shift countdown, dialog waktu habis/overtime, dan timer sync. |
| [03-juri-hud.md](./03-juri-hud.md) | Monitor & HUD Dewan Juri | Grid kamera 3x2, rotasi juri (Shift 1–4 -> Zona A–D), dan popup alert pergantian shift. |
| [04-recordings-and-whatsapp.md](./04-recordings-and-whatsapp.md) | Rekaman & Arsip WhatsApp | Manajemen arsip video per gantangan dan modal pengiriman link video via WhatsApp. |
| [05-desktop-settings.md](./05-desktop-settings.md) | Pengaturan Desktop & Tema | Konfigurasi tema gelap/terang, kalibrasi ambang batas audio dBFS, dan binding kamera. |
