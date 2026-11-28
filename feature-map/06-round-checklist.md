# Feature Map: Checklist & Kontrol Ronde (Operator & Admin)

Dokumen ini mendefinisikan peta interaksi untuk halaman **Checklist & Kontrol Ronde** (`RoundChecklistPage`, `tab === "checklist"`), yang memusatkan kontrol Start/Stop ronde, daftar ronde hari ini, monitoring posisi & waktu rotasi juri, serta checklist verifikasi burung–gantangan–WhatsApp.

Jadwal ronde 10 menit dan sinkronisasi waktu Operator–Juri yang ditargetkan dijelaskan di [Aturan Ronde Operator dan Juri](./07-competition-round-flow.md).

---

## Fitur 6.1: Kontrol Siklus Ronde (Start, Stop, Reset, Pilih Sesi Cloud)

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Kontrol Putaran Ronde & Multi-Shift Timer
- **Fungsi:** Memilih sesi lomba dari Cloud, memulai (*Start*) hitung mundur ronde (4 shift juri), menghentikan (*Stop*) ronde, dan membuka dialog penyimpanan rekaman (*Quick Save Modal*) serta *Time-Up Dialog*.

### 2. Cara User Mengakses Fitur
- Masuk ke tab navigasi **`Checklist & Kontrol Ronde`** (`tab === "checklist"`, tersedia untuk role `operator` dan `admin`).
- Terletak pada panel header atas (`OperatorHeader`).

---

## Fitur 6.2: Checklist Ronde Hari Ini & Monitoring Posisi Rotasi Juri

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Today's Round Picker & Live Judge Position Matrix
- **Fungsi:**
  - Menampilkan daftar ronde dari Cloud Arena aktif lengkap dengan nama event, jenis burung/spesies, jadwal jam (`start_time - end_time`), dan status ronde.
  - Menampilkan posisi **4 Juri (`J1–J4`)** secara real-time: zona aktif saat ini (`Zona A/B/C/D`) dan jadwal rentang waktu untuk setiap shift (`S1..S4`).

---

## Fitur 6.3: Checklist Kesesuaian Burung, Gantangan & WhatsApp

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Gantangan Verification Checklist (Check / Uncheck)
- **Fungsi:** Memudahkan operator memeriksa kesesuaian burung pada setiap tiang gantangan (1–24), melihat nama pemilik, serta nomor **WhatsApp** tiap gantangan.

### 2. Interaksi yang Harus Dilakukan
1. Operator memilih ronde aktif.
2. Operator memeriksa baris gantangan `#01` s/d `#24` (nomor gantangan, zona, nama & kode burung, pemilik, nomor WhatsApp).
3. Klik baris atau tombol checkbox untuk melakukan **check / uncheck**, atau gunakan tombol **Check Semua** / **Uncheck Semua** dan kolom pencarian cepat.

### 3. Hasil yang Diharapkan
- Status checklist (`Sesuai` / `Belum Dicek`) tersimpan persisten per `roundId` di `localStorage` dan tersinkronisasi melalui `BroadcastChannel`.
