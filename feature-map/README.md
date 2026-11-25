# Feature Map - Kicau Mania Desktop (Wails + Svelte)

Direktori ini berisi katalog interaktif dan peta fitur (*Feature Map*) lengkap untuk aplikasi **Kicau Mania Desktop**.
Peta ini dirancang untuk developer, QA tester, dan AI Agent (khususnya *Browser / UI Automation Agent*) agar dapat menavigasi, berinteraksi, dan menguji seluruh antarmuka secara deterministik.

---

## Ringkasan Peran & Akses Antarmuka (Role-Based Access)

| Role | Tab Utama yang Dapat Diakses | Deskripsi Peran |
| :--- | :--- | :--- |
| **`juri`** | Monitor Juri (`cameras`) | Layar khusus dewan juri dengan grid kamera 3x2, penugasan zona A-D otomatis, dan timer shift. |
| **`operator`** | Checklist & Kontrol Ronde (`checklist`), Livestream (`operator`), Monitor Juri (`cameras`), Rekaman & Arsip (`recordings`) | Operator teknis arena yang memverifikasi checklist gantangan–burung–WhatsApp, mengontrol start/stop ronde, memantau rotasi juri, serta melihat livestream full video. |
| **`admin`** | Seluruh Tab + Pengaturan Desktop (`settings`) | Akses penuh termasuk konfigurasi NVR, kalibrasi audio, dan endpoint cloud. |

---

## Indeks Dokumen Fitur

| File | Nama Fitur Utama | Deskripsi Singkat |
| :--- | :--- | :--- |
| [01-auth-and-session.md](./01-auth-and-session.md) | Autentikasi & Pemilihan Sesi | Layar login akun Cloud (2-step scoped login) dan picker sesi lomba. |
| [02-operator-console.md](./02-operator-console.md) | Livestream (Full Video 24) | Tampilan full video 24 gantangan dengan filter zona (A–D) dan mode Spotlight. |
| [03-juri-hud.md](./03-juri-hud.md) | Monitor & HUD Dewan Juri | Grid kamera 3x2, checklist zona & jadwal waktu per juri, rotasi shift, dan alert pergantian shift. |
| [04-recordings-and-whatsapp.md](./04-recordings-and-whatsapp.md) | Rekaman & Arsip WhatsApp | Manajemen arsip video per gantangan dan modal pengiriman link video via WhatsApp. |
| [05-desktop-settings.md](./05-desktop-settings.md) | Pengaturan Desktop & Tema | Konfigurasi tema gelap/terang, kalibrasi ambang batas audio dBFS, dan binding kamera. |
| [06-round-checklist.md](./06-round-checklist.md) | Checklist & Kontrol Ronde | Checklist ronde hari ini, kontrol Start/Stop ronde, monitoring posisi juri, serta verifikasi gantangan–burung–WhatsApp (check/uncheck). |
