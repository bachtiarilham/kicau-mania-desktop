# Feature Map: Monitor & HUD Dewan Juri (Grid 3x2)

Dokumen ini mendefinisikan peta interaksi untuk layar monitor dewan juri (`ZoneGrid3x2`), pemetaan rotasi shift ke zona penilaian (A, B, C, D), dan popup notifikasi pergantian shift.

---

## Fitur 3.1: Matriks Kamera Juri (Grid 3x2)

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** 6-Camera Grid Matrix (Zone View)
- **Fungsi:** Menampilkan 6 kamera gantangan sekaligus secara real-time yang sesuai dengan Zona tugas juri saat ini (misal Zona A: Gantangan 1–6).

### 2. Cara User Mengakses Fitur
- Pengguna dengan role **`juri`** otomatis masuk ke layar ini saat login.
- Pengguna dengan role **`operator` / `admin`** dapat mengklik tab **`Monitor Juri (Grid 3x2)`** pada bilah samping.

### 3. Keyboard Shortcut
- `1` s/d `6` : Membuka mode full-focus pada kamera gantangan ke-1 hingga ke-6 di grid aktif.
- `Esc` : Keluar dari mode full-focus kembali ke grid 3x2.

### 4. DOM Element / Selector yang Digunakan
| Elemen | Selector / ID | Tipe |
| :--- | :--- | :--- |
| Grid Container | `.grid.grid-cols-3.grid-rows-2`, `div[data-testid="juri-grid"]` | Grid Container |
| Slot Kamera Juri (1-6) | `div.camera-slot[data-slot-index]` | Camera Card |
| Badge Nomor Gantangan | `span.gantangan-pill`, `span:has-text("Gantangan")` | Visual Badge |
| Indikator Zona Aktif Juri | `header .judge-zone-indicator`, `span.font-bold:has-text("ZONA")` | Status Badge |
| Indikator Shift Saat Ini | `header .judge-shift-indicator`, `span:has-text("Shift")` | Status Badge |
| Countdown Timer Shift Juri | `header .judge-shift-countdown` | Display Timer |

### 5. Interaksi yang Harus Dilakukan
1. Dewan juri mengamati 6 burung di zonanya selama shift berjalan.
2. Juri dapat mengklik salah satu kotak gantangan untuk memperbesar fokus (*Single Bird Focus*).
3. Audio burung pada gantangan yang dipilih akan otomatis terdengar di headset juri.

### 6. Hasil yang Diharapkan
- Stream kamera RTSP/WebRTC ter-render dengan latensi sangat rendah (< 300ms).
- Nama burung, nomor gantangan, dan nama pemilik (jika terdaftar di cloud) muncul di header masing-masing slot video.

---

## Fitur 3.2: Rotasi Juri Otomatis & Popup Notifikasi Shift

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Judge Shift Rotation Matrix & Audio-Visual Alert
- **Fungsi:** Menghitung rumus rotasi juri `(((judgeIndex % 4) + (currentShift - 1)) % 4)` dan menampilkan banner alert mencolok ketika waktu shift juri habis agar dewan juri berpindah blok pengamatan.

### 2. Cara User Mengakses Fitur
- Berjalan otomatis di latar belakang melalui sinkronisasi `BroadcastChannel` saat operator memulai ronde.

### 3. Matriks Perputaran Rotasi Juri:
- **Shift 1:** Juri 1 -> Zona A | Juri 2 -> Zona B | Juri 3 -> Zona C | Juri 4 -> Zona D
- **Shift 2:** Juri 1 -> Zona B | Juri 2 -> Zona C | Juri 3 -> Zona D | Juri 4 -> Zona A
- **Shift 3:** Juri 1 -> Zona C | Juri 2 -> Zona D | Juri 3 -> Zona A | Juri 4 -> Zona B
- **Shift 4:** Juri 1 -> Zona D | Juri 2 -> Zona A | Juri 3 -> Zona B | Juri 4 -> Zona C

### 4. DOM Element / Selector yang Digunakan
| Elemen | Selector / ID | Tipe |
| :--- | :--- | :--- |
| Overlay Peringatan Rotasi | `div.rotation-alert-overlay`, `.animate-pulse` | Fullscreen Banner |
| Teks Perintah Rotasi | `h1:has-text("WAKTUNYA ROTASI JURI!")` | Alert Header |
| Keterangan Zona Baru | `p:has-text("Silakan berpindah ke ZONA")` | Instruction Text |
| Tombol Tutup / Konfirmasi Alert | `button:has-text("Saya Mengerti")` | Action Button |

### 5. Interaksi yang Harus Dilakukan
1. Saat timer shift tersisa 10 detik, hitung mundur berwarna merah berkedip di header.
2. Saat detik 0 tercapai, banner rotasi muncul dan suara buzzer berbunyi.
3. Juri mengklik tombol `button:has-text("Saya Mengerti")` atau menunggu banner otomatis hilang dalam 5 detik.
4. Juri melihat indikator zona baru di header (misal dari "ZONA A" berubah menjadi "ZONA B").

### 6. Hasil yang Diharapkan
- Grid 3x2 kamera otomatis memuat 6 kamera gantangan dari blok zona yang baru tanpa perlu refresh manual.

---

## Fitur 3.3: Checklist Penilaian per Zona & Jadwal Waktu Otomatis

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Judge Zone Checklist & Auto-Scheduled Shift Windows
- **Fungsi:** Menampilkan daftar 4 zona penilaian (`A`, `B`, `C`, `D`) yang urutannya sudah ditentukan otomatis oleh sistem sesuai indeks juri (`J1–J4`), lengkap dengan rentang jam mulai–selesai tiap shift berdasarkan `startTime` ronde dan durasi shift.

### 2. Interaksi yang Harus Dilakukan
1. Juri melihat blok **Checklist Zona Anda** pada panel HUD di sisi kiri.
2. Sistem otomatis menyorot zona yang sedang aktif beserta rentang waktunya (misal `13:00–13:02:30`).
3. Juri dapat melakukan **check / uncheck** pada setiap zona yang telah selesai dinilai.
