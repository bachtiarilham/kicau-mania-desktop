# Feature Map: Konsol Operator Arena (Live 24)

Dokumen ini mendefinisikan peta interaksi untuk dasbor utama operator arena lomba (`OperatorDashboard`), pengelolaan waktu ronde/shift, dialog perpanjangan waktu (*Overtime Dialog*), dan penyimpanan berkas rekaman.

---

## Fitur 2.1: Manajemen Siklus Ronde (Start, Pause, Stop)

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Kontrol Putaran Ronde & Multi-Shift Timer
- **Fungsi:** Memulai hitung mundur ronde lomba (default 10 menit terbagi 4 shift juri masing-masing 2.5 menit), memicu rotasi juri antar blok, dan menghentikan sesi.

### 2. Cara User Mengakses Fitur
- Masuk ke tab navigasi: **`Operator Arena (Live 24)`** (`tab === "operator"`).
- Terletak pada panel kontrol atas (*Top Control Bar*).

### 3. Keyboard Shortcut
- `Space` : Memulai (*Start*) atau Menghentikan (*Pause*) ronde jika fokus pada tombol kontrol.
- `Esc` : Menutup dialog konfirmasi simpan atau dialog perpanjangan waktu.

### 4. DOM Element / Selector yang Digunakan
| Elemen | Selector / ID | Tipe |
| :--- | :--- | :--- |
| Tombol Mulai Ronde | `button:has-text("Mulai Ronde")`, `button:has(.lucide-play)` | Primary Button |
| Tombol Berhenti / Selesai | `button:has-text("Selesai")`, `button:has(.lucide-square)` | Destructive Button |
| Tombol Reset Timer | `button:has(.lucide-rotate-ccw)` | Secondary Button |
| Display Timer Total | `.round-timer-display`, `span:has-text("Waktu Tersisa")` | Display Counter |
| Display Shift Aktif | `span:has-text("Shift")`, `.shift-badge` | Badge Indicator |
| Display Timer Shift | `.shift-timer-countdown` | Display Counter |
| Banner Alert Rotasi | `.bg-amber-500\/20`, `div:has-text("Waktunya Rotasi Juri!")` | Alert Banner |

### 5. Interaksi yang Harus Dilakukan
1. Pastikan sesi lomba sudah terpilih pada *SessionBar*.
2. Klik tombol `button:has-text("Mulai Ronde")`.
3. Timer total (`roundRemainingSeconds`) dan timer shift (`shiftRemainingSeconds`) mulai menghitung mundur.
4. Ketika timer shift habis (mencapai 0 detik), alert rotasi muncul dan nomor shift berpindah dari Shift 1 ke Shift 2.
5. Untuk menyelesaikan ronde lebih awal, klik `button:has-text("Selesai")`.

### 6. Hasil yang Diharapkan
- State berubah dari `IDLE` -> `RUNNING`.
- Broadcast event dikirim via `BroadcastChannel("kicau_round_sync_channel")` sehingga seluruh layar HUD Juri dan overlay streaming mendapatkan update detik yang sinkron.
- NVR recording service memulai perekaman ke folder lokal.

---

## Fitur 2.2: Dialog Waktu Habis & Mode Overtime (Time-Up Dialog)

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Manual Overtime & Completion Dialog
- **Fungsi:** Menampilkan dialog interaktif saat waktu ronde reguler (600 detik) habis agar operator dapat memutuskan apakah ronde benar-benar disudahi atau dilanjutkan perpanjangan waktu (*overtime*) jika juri masih butuh penilaian tambahan.

### 2. Cara User Mengakses Fitur
- Muncul secara otomatis sebagai modal dialog saat `roundRemainingSeconds === 0` dan ronde berstatus `RUNNING`.

### 3. Keyboard Shortcut
- `Enter` : Memilih aksi "Lanjutkan Ronde" (Overtime).
- `Esc` : Memilih aksi "Sudahi Ronde Sekarang".

### 4. DOM Element / Selector yang Digunakan
| Elemen | Selector / ID | Tipe |
| :--- | :--- | :--- |
| Modal Container | `div[role="alertdialog"]`, `.round-timeup-dialog` | Modal Dialog |
| Teks Peringatan Waktu | `h2:has-text("Waktu Ronde Telah Habis")` | Modal Header |
| Indikator Overtime Counter | `span.overtime-timer` | Display Counter |
| Tombol Lanjutkan (Overtime) | `button:has-text("Lanjutkan Tambahan Waktu")` | Action Button |
| Tombol Sudahi Ronde | `button:has-text("Sudahi Ronde Sekarang")` | Destructive Action |

### 5. Interaksi yang Harus Dilakukan
- **Opsi A (Lanjutkan):** Klik `button:has-text("Lanjutkan Tambahan Waktu")`.
- **Opsi B (Sudahi):** Klik `button:has-text("Sudahi Ronde Sekarang")`.

### 6. Hasil yang Diharapkan
- Jika memilih **Lanjutkan**: Dialog tertutup, status ronde tetap `RUNNING`, dan waktu lembur dihitung maju (`overtimeSeconds++`) dengan visual kuning/oranye.
- Jika memilih **Sudahi**: Dialog tertutup, status ronde beralih ke `FINISHED`, timer berhenti, dan dialog ekspor file rekaman terbuka.

---

## Fitur 2.3: Dialog Ekspor & Simpan Rekaman Sesi

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Quick Save & Export Session Modal
- **Fungsi:** Menyimpan potongan video seluruh gantangan pada sesi tersebut dengan format nama file standar dan mengekspornya ke disk.

### 2. Cara User Mengakses Fitur
- Muncul otomatis saat tombol **Selesai** diklik atau saat ronde disudahi dari *Time-Up Dialog*.

### 3. DOM Element / Selector yang Digunakan
| Elemen | Selector / ID | Tipe |
| :--- | :--- | :--- |
| Modal Dialog | `div[role="dialog"]:has-text("Simpan Rekaman Sesi")` | Modal Dialog |
| Input Nama Sesi Berkas | `input#sessionName`, `input[placeholder*="Nama sesi"]` | Text Input |
| Dropdown Format Video | `select#selectedFormat` (Pilihan: `mp4`, `mkv`) | Select Input |
| Input Direktori Target | `input#storageDir` (Default: `D:\KicauRecordings`) | Text Input |
| Tombol Simpan & Ekspor | `button:has-text("Simpan Rekaman")` | Primary Action |
| Tombol Batal | `button:has-text("Batal")` | Cancel Button |

### 4. Interaksi yang Harus Dilakukan
1. Periksa nama sesi yang di-generate otomatis (format: `Event_Sesi_Tanggal`).
2. Pilih format video (`mp4` direkomendasikan untuk WhatsApp).
3. Klik tombol `button:has-text("Simpan Rekaman")`.

### 5. Hasil yang Diharapkan
- Background task media engine menggabungkan dan menyimpan file rekaman ke direktori tujuan.
- Muncul notifikasi banner hijau (*toast notification*): *"Berhasil menyimpan N file rekaman sesi ke D:\KicauRecordings"*.
- State ronde kembali ke `IDLE` dan siap untuk sesi berikutnya.

---

## Fitur 2.4: Filter Zona & Spotlight Monitor Kamera

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Live 24 Camera Grid Filter & Spotlight
- **Fungsi:** Memfilter tampilan 24 kamera gantangan berdasarkan blok Zona (A, B, C, D) serta memperbesar satu gantangan (*Spotlight Mode*) lengkap dengan *audio meter* dan *digital zoom*.

### 2. Cara User Mengakses Fitur
- Di dalam tab **`Operator Arena (Live 24)`** pada area *Camera Matrix*.

### 3. DOM Element / Selector yang Digunakan
| Elemen | Selector / ID | Tipe |
| :--- | :--- | :--- |
| Tombol Filter "Semua" | `button:has-text("Semua (24)")` | Filter Tab |
| Tombol Filter Zona A | `button:has-text("Blok A")` | Filter Tab |
| Tombol Filter Zona B | `button:has-text("Blok B")` | Filter Tab |
| Tombol Filter Zona C | `button:has-text("Blok C")` | Filter Tab |
| Tombol Filter Zona D | `button:has-text("Blok D")` | Filter Tab |
| Kartu Kamera Gantangan | `div.gantangan-card[data-gantangan-id]` | Camera Card |
| Tombol Spotlight (Maximize) | `button[aria-label="Perbesar"]`, `.lucide-maximize-2` | Card Action |
| Tombol Dengarkan Audio | `button[aria-label="Dengarkan Audio"]`, `.lucide-volume-2` | Card Action |
| Slider Digital Zoom | `input[type="range"].zoom-slider` | Slider Input |

### 4. Interaksi yang Harus Dilakukan
1. Klik tombol filter `button:has-text("Blok A")` untuk hanya melihat gantangan 1–6 (Zona A).
2. Klik tombol maximize pada salah satu gantangan untuk masuk ke mode *Spotlight*.
3. Geser *slider digital zoom* ke kanan untuk memperbesar visual burung hingga 4x.
4. Klik tombol speaker untuk mengaktifkan saluran mikrofon gantangan tersebut ke headphone operator.

### 5. Hasil yang Diharapkan
- Grid kamera terfilter secara instan tanpa buffering ulang video stream.
- Kamera yang di-spotlight memenuhi layar dengan visual resolusi penuh dan indikator dBFS audio live bergerak.
