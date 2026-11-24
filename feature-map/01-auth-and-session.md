# Feature Map: Autentikasi & Pemilihan Sesi Lomba

Dokumen ini mendefinisikan peta interaksi untuk login pengguna, konfigurasi endpoint cloud arena, dan pemilihan sesi lomba aktif (*Session Picker*).

---

## Fitur 1.1: Login Akun Operator / Juri / Admin

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Fullscreen Arena Login & Role Authentication
- **Fungsi:** Mengotentikasi pengguna ke server cloud atau mode arena lokal offline, menentukan role pengguna (`admin`, `operator`, `juri`), serta memuat profil arena aktif.

### 2. Cara User Mengakses Fitur
- Muncul secara otomatis saat aplikasi desktop pertama kali dibuka ketika pengguna belum login (`!auth.isLoggedIn`).
- Dapat diakses kembali dari sidebar/header dengan mengklik tombol **Logout**.

### 3. Keyboard Shortcut
- `Enter` : Melakukan submit form login saat fokus berada di input email atau password.
- `Tab` / `Shift+Tab` : Berpindah fokus antar field input.

### 4. DOM Element / Selector yang Digunakan
| Elemen | Selector / ID | Tipe |
| :--- | :--- | :--- |
| Input Email | `input#email` atau `input[type="email"]` | Text Input |
| Input Password | `input#password` atau `input[type="password"]` | Password Input |
| Tombol Submit Login | `button[type="submit"]` (Text: *"Masuk ke Sistem"*) | Action Button |
| Tombol Quick Role Admin | `button:has-text("Admin Arena")` | Quick Preset |
| Tombol Quick Role Operator | `button:has-text("Operator Lomba")` | Quick Preset |
| Tombol Quick Role Juri | `button:has-text("Dewan Juri")` | Quick Preset |
| Toggle Server Cloud URL | `button:has-text("Pengaturan Alamat Server")` | Toggle Button |
| Input Server Cloud URL | `input[placeholder*="https://"]` | Text Input |
| Tombol Simpan URL Cloud | `button:has-text("Simpan Alamat Server")` | Action Button |
| Notifikasi Error | `.text-destructive`, `p:has-text("Gagal")` | Status Text |

### 5. Interaksi yang Harus Dilakukan
1. Ketik email pengguna pada selector `input#email` (Contoh: `operator@kicaumania.com`).
2. Ketik password pada selector `input#password` (Contoh: `operator123`).
3. Atau klik tombol Quick Preset (`button:has-text("Operator Lomba")`) untuk autofill instan.
4. Klik tombol `button[type="submit"]` atau tekan tombol `Enter`.

### 6. Hasil yang Diharapkan
- **Jika Sukses:**
  - Token autentikasi dan data role tersimpan di `auth.svelte`.
  - Halaman login tertutup secara mulus.
  - Jika login sebagai **`juri`**, layar langsung diarahkan ke *Layar Penilaian Dewan Juri* (`Monitor Juri (Grid 3x2)`).
  - Jika login sebagai **`operator` / `admin`**, layar diarahkan ke *Operator Arena (Live 24)*.
- **Jika Gagal:**
  - Teks error muncul di bawah form: *"Silakan periksa kembali email dan kata sandi Anda."*

---

## Fitur 1.2: Pemilihan Sesi & Ronde Lomba (Session Picker)

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Cloud Session & Round Selector Bar
- **Fungsi:** Mengambil daftar event dan sesi/ronde yang aktif dari Cloud Database dan menetapkannya sebagai sesi aktif sebelum perekaman video dimulai.

### 2. Cara User Mengakses Fitur
- Terletak pada bagian atas bilah navigasi operator: komponen `SessionBar` (`cloud-session-selector.svelte`).
- Dialog pemilih sesi muncul otomatis saat operator hendak memulai perekaman jika sesi belum ditentukan.

### 3. Keyboard Shortcut
- `Esc` : Menutup dialog modal pemilihan sesi tanpa menyimpan perubahan.

### 4. DOM Element / Selector yang Digunakan
| Elemen | Selector / ID | Tipe |
| :--- | :--- | :--- |
| Bilah Sesi Aktif | `header .session-bar`, `button[data-testid="session-selector-trigger"]` | Bar Trigger |
| Dropdown Pilihan Event | `select#event-select`, `div[role="combobox"]` | Dropdown |
| Dropdown Pilihan Ronde | `select#round-select` | Dropdown |
| Tombol Refresh Data Cloud | `button[aria-label="Refresh Sesi"]`, `button:has(.lucide-refresh-cw)` | Action Button |
| Tombol Terapkan Sesi | `button:has-text("Gunakan Sesi Ini")` | Primary Action |
| Label Sesi Terpilih | `span.session-active-title` | Display Label |

### 5. Interaksi yang Harus Dilakukan
1. Klik bilah sesi pada header atau tombol trigger `button[data-testid="session-selector-trigger"]`.
2. Pilih nama event dari daftar (misal *"Piala Umi Kasum"*).
3. Pilih ronde aktif (misal *"Sesi 1 - Murai Batu Ring A"*).
4. Klik tombol `button:has-text("Gunakan Sesi Ini")`.

### 6. Hasil yang Diharapkan
- Data ronde dan pemetaan gantangan peserta (`ParticipantMapping`) dimuat ke dalam `round.svelte.ts`.
- Broadcast channel `kicau_round_sync_channel` menyiarkan sesi terpilih ke seluruh tab juri dan streaming overlay.
- Label sesi pada HUD Juri dan Konsol Operator terbarui dengan nama sesi yang dipilih.
