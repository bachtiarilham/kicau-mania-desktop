# Feature Map: Rekaman, Arsip & Distribusi WhatsApp

Dokumen ini mendefinisikan peta interaksi untuk penelusuran arsip rekaman video lomba (`RecordingsPage`), pemutaran ulang lokal (*Native Player*), sinkronisasi ke cloud, dan pengiriman tautan video rekaman ke peserta via WhatsApp.

---

## Fitur 4.1: Penelusuran & Pemutaran Video Arsip (Local & Cloud)

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** Recording Archive Explorer & Native Video Player
- **Fungsi:** Menampilkan daftar file video hasil rekaman sesi berdasarkan event, sesi, dan gantangan, serta memutar ulang rekaman secara instan di desktop.

### 2. Cara User Mengakses Fitur
- Klik tab menu samping: **`Rekaman & Arsip`** (`activeTab === "local"` atau `"cloud"`).

### 3. DOM Element / Selector yang Digunakan
| Elemen | Selector / ID | Tipe |
| :--- | :--- | :--- |
| Tab Arsip Lokal | `button:has-text("Arsip Lokal")` | Tab Filter |
| Tab Antrean Upload Cloud | `button:has-text("Antrean Unggah")` | Tab Filter |
| Tab Arsip Cloud | `button:has-text("Cloud Publik")` | Tab Filter |
| Dropdown Filter Ronde | `select[name="round"]`, `select#round-filter` | Select Input |
| Tabel Baris Rekaman | `table tr.recording-row`, `div.recording-card` | Data Row |
| Tombol Putar Video | `button[aria-label="Putar"]`, `button:has(.lucide-play)` | Action Button |
| Tombol Buka Folder File | `button:has-text("Buka Folder")`, `.lucide-folder-open` | External Action |
| Container Pemutar Native | `div.native-player-container`, `#video-player-modal` | Modal / Embed |

### 4. Interaksi yang Harus Dilakukan
1. Pilih tab `button:has-text("Arsip Lokal")`.
2. Filter berdasarkan nomor ronde jika arsip sangat banyak.
3. Klik tombol `Putar` pada baris gantangan yang ingin ditonton.
4. Klik tombol `Buka Folder` jika ingin melihat file MP4 langsung di Windows Explorer.

### 5. Hasil yang Diharapkan
- Video rekaman gantangan terkait diputar dengan kendala audio normal dan *seek bar*.
- Folder file `D:\KicauRecordings\<sesi>` terbuka di File Explorer saat tombol folder diklik.

---

## Fitur 4.2: Distribusi Video Peserta via WhatsApp

### 1. Nama & Fungsi Fitur
- **Nama Fitur:** WhatsApp Video Delivery Action
- **Fungsi:** Mengirimkan tautan tonton video publik (`https://.../v/:id`) langsung ke nomor WhatsApp pemilik burung berdasarkan pemetaan gantangan ronde tersebut.

### 2. Cara User Mengakses Fitur
- Pada baris rekaman di tab **`Rekaman & Arsip`**, atau tombol aksi batch pada header arsip.

### 3. DOM Element / Selector yang Digunakan
| Elemen | Selector / ID | Tipe |
| :--- | :--- | :--- |
| Tombol Kirim WhatsApp | `button:has-text("Kirim WhatsApp")`, `button:has(.lucide-message-circle)` | Action Button |
| Modal Konfirmasi Pengiriman | `div[role="dialog"]:has-text("Kirim Rekaman via WhatsApp")` | Modal Dialog |
| Field Nomor WhatsApp Peserta | `input#participant-phone` | Text Input |
| Template Pesan WhatsApp | `textarea#wa-message-preview` | Text Area |
| Tombol Eksekusi Pengiriman | `button:has-text("Kirim Sekarang")` | Primary Action |
| Indikator Status Pengiriman | `span.wa-delivery-status`, `span:has-text("Terkirim")` | Status Badge |

### 4. Interaksi yang Harus Dilakukan
1. Klik tombol `Kirim WhatsApp` pada kartu gantangan peserta.
2. Modal konfirmasi terbuka menampilkan nomor WhatsApp pemilik burung dan pratinjau pesan:
   ```text
   Halo [Nama Pemilik], berikut adalah link video penampilan burung [Nama Burung] pada [Event] [Sesi]:
   https://dev.megaallstararena.com/v/rec-12345
   Terima kasih atas partisipasi Anda!
   ```
3. Klik tombol `button:has-text("Kirim Sekarang")`.

### 5. Hasil yang Diharapkan
- Desktop memanggil endpoint backend `POST /api/v1/arenas/:arena/rounds/:id/deliveries/whatsapp`.
- Backend memproses pengiriman melalui WhatsApp Gateway dengan ketahanan Circuit Breaker.
- Badge pada baris gantangan berubah menjadi hijau: `Terkirim (WIB)`.
