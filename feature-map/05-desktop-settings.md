# Feature Map: Pengaturan Desktop & Tema Antarmuka

Dokumen ini mendefinisikan peta interaksi untuk konfigurasi preferensi tampilan (*Dark/Light Mode*), integrasi edge engine lokal, serta kalibrasi audio dan subnet LAN (`SettingsPage`).

## Pengaturan Arena

Operator dan admin membuka **Pengaturan Arena** di desktop PC server untuk memilih IP LAN operator, mengisi hingga empat IP PC juri, dan mengatur IP NVR. IP router menampilkan gateway yang sedang dipakai Windows untuk IP server tersimpan.

Daftar CCTV menyediakan 24 slot gantangan dan satu **Kamera Umum Arena**. **Tes NVR** memeriksa port RTSP 554. **Tes Video** memeriksa sumber RTSP tiap kamera. Kamera umum dapat dibuka dari layar Livestream operator.

Simpan pengaturan untuk memperbarui `.env` Local App. Mulai ulang Local App saat ronde dan rekaman sudah berhenti. Di PC juri, atur alamat server melalui **Server → Ubah** di layar login.

---

## Fitur 5.1: Penggantian Tema Antarmuka (Dark / Light Theme)

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Theme Preference Switcher (Gelap / Terang)
- **Fungsi:** Mengubah skema visual antarmuka antara tema gelap (*default high-contrast dark theme* yang nyaman di arena minim cahaya) dan tema terang (*light theme*).

### 2. Cara User Mengakses Fitur
- Ikon toggle tema di bagian bawah sidebar navigasi desktop, atau melalui halaman **`Pengaturan desktop`**.

### 3. Keyboard Shortcut
- `Ctrl+Shift+T` / `Cmd+Shift+T` : Beralih tema secara instan (jika diaktifkan di global listener).

### 4. DOM Element / Selector yang Digunakan
| Elemen | Selector / ID | Tipe |
| :--- | :--- | :--- |
| Tombol Toggle Tema Sidebar | `button[aria-label="Toggle Theme"]`, `button:has(.lucide-sun)`, `button:has(.lucide-moon)` | Action Button |
| Pilihan Radio Tema Gelap | `input[type="radio"][value="dark"]`, `label:has-text("Mode Gelap")` | Radio Button |
| Pilihan Radio Tema Terang | `input[type="radio"][value="light"]`, `label:has-text("Mode Terang")` | Radio Button |

### 5. Interaksi yang Harus Dilakukan
1. Klik tombol ikon matahari/bulan pada bagian pojok bawah sidebar navigasi.
2. Atau masuk ke menu **Pengaturan desktop** dan pilih radio button **Mode Gelap**.

### 6. Hasil yang Diharapkan
- Kelas CSS `.dark` diterapkan pada elemen `html` / `body`.
- Seluruh palet warna latar belakang berubah menjadi gelap pekat (`bg-background: #09090b`), kontras warna teks dan timer meningkat tajam, serta preferensi disimpan ke `localStorage`.

---

## Fitur 5.2: Konfigurasi Media Engine & Kalibrasi Audio

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Local Media Engine & Audio Calibration
- **Fungsi:** Mengatur URL backend edge (`http://127.0.0.1:8090`), folder penyimpanan lokal, ambang batas desibel mikrofon gantangan (*Noise Gate / Active dBFS*), dan subnet penonton LAN.

### 2. Cara User Mengakses Fitur
- Klik tab menu samping: **`Pengaturan desktop`** (Khusus role `admin`).

### 3. DOM Element / Selector yang Digunakan
| Elemen | Selector / ID | Tipe |
| :--- | :--- | :--- |
| Input URL Media Engine | `input#engineUrl`, `input[name="engine_url"]` | Text Input |
| Input Direktori Rekaman | `input#recordingsDir` | Text Input |
| Input Ambang Audio dBFS | `input#audioThreshold`, `input[type="number"]` (Default: `-35`) | Number Input |
| Input Subnet LAN CIDR | `input#readCidrs` (Default: `192.168.1.0/24`) | Text Input |
| Tombol Simpan Konfigurasi | `button:has-text("Simpan Pengaturan")` | Primary Action |
| Tombol Uji Koneksi Engine | `button:has-text("Uji Koneksi Engine")` | Secondary Action |

### 4. Interaksi yang Harus Dilakukan
1. Ubah direktori simpan rekaman jika drive default penuh (misal ganti dari `D:\KicauRecordings` ke `E:\Recordings`).
2. Sesuaikan *Ambang Audio* jika suara bising penonton di arena terlalu kencang (ubah dari `-35 dBFS` ke `-30 dBFS`).
3. Klik tombol `button:has-text("Simpan Pengaturan")`.

### 5. Hasil yang Diharapkan
- Konfigurasi dikirim ke file konfigurasi lokal desktop dan disinkronisasikan ke edge engine.
- Status koneksi menampilkan indikator hijau: *"Engine Terhubung & Siap"*.
