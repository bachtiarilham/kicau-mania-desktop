# Kontrak saving dan upload schema 011

Prefix cloud `/api/v1`, envelope `{success,message,data,errors}`. Semua named SQL berada di `cloud-app/db/queries/desktop_media.sql`; migrasi 011 additif. PostgreSQL/Redis autentikasi pengguna tetap digunakan. Redis bukan penyimpanan antrean file.

| Endpoint | Auth dan fungsi |
|---|---|
| POST `/media/engines/{engine_id}` | ADMIN arena, body `{arena_id}`; UUID engine dibuat desktop. Menghasilkan token opaque sekali; pairing ulang ID/arena sama merotasi token. |
| DELETE `/media/engines/{engine_id}` | ADMIN arena; mencabut credential tanpa menghapus arsip. |
| POST `/media/ingestions` | ADMIN/OPERATOR arena; `{engine_id,session_id,round_id,mappings:[{channel_id,registration_id,gantangan_id}]}`. Tiket harus confirmed/verified dan scope ronde/arena/gantangan cocok. Mapping identik idempotent; perubahan mapping sesi yang sudah terikat 409. |
| POST `/engine/media/uploads` | Credential engine; `{container:"mkv",session_id,channel_id,recording_id,size_bytes,sha256,started_at,ended_at}`. SHA-256 file hex; membuat atau memulihkan upload idempotent. |
| GET `/engine/media/uploads/{id}` | Credential engine pemilik; status dan parts yang dikonfirmasi storage, termasuk recovery upload kedaluwarsa. |
| POST `/engine/media/uploads/{id}/parts` | `{number,sha256}`; checksum part dalam Base64. Response `{url,headers}`, berlaku 15 menit. |
| POST `/engine/media/uploads/{id}/complete` | `{}`; mengambil parts dari storage, menyelesaikan multipart, memeriksa ukuran/content-type/metadata checksum lalu commit UPLOADED. Retry setelah objek selesai tetapi SQL gagal tidak mengunggah file kembali. |
| GET `/media/recordings` | ADMIN/OPERATOR arena atau akun peserta untuk miliknya. `arena_id` wajib; filter round_id/session_id/channel_id/cursor/limit. |
| GET `/media/recordings/{id}` | Detail terotorisasi tanpa multipart ID atau path lokal. |
| GET `/media/recordings/{id}/download` | Akses objek terotorisasi, URL berlaku 60 detik; tidak menyimpan signed URL permanen dalam DB. |

Multipart berukuran 16 MiB per part; maksimal 10.000 part. Object key dibuat cloud, bukan input path dari desktop. Storage memverifikasi SHA-256 setiap part yang ditandatangani. Hash file penuh dihitung engine dan dipersistenkan; HEAD memverifikasi ukuran serta metadata hash yang dikaitkan saat inisiasi, bukan menghitung ulang seluruh file di VPS.

Identitas unik `(engine_id,local_recording_id)` menjaga idempotensi. Tabel `media.recording_assets` mempertahankan setiap sesi/channel dan mengacu mapping; tabel lama `media.round_recordings` tidak ditimpa. Penyelesaian upload tidak mengubah pembayaran, status peserta, atau status ronde. Ronde CLOSED tetap menerima penyelesaian arsip.

API local-app tambahan khusus mode LAN, loopback dan nonce:

| Endpoint | Fungsi |
|---|---|
| GET/POST `/api/v1/desktop/cloud` | Status pairing tanpa token / simpan pairing terenkripsi. |
| POST `/api/v1/desktop/mappings` | Simpan hasil mapping cloud untuk sesi lokal; immutable setelah ada. |
| GET `/api/v1/recordings` | Cursor/limit; filter session_id/channel_id/state. |
| GET `/api/v1/recordings/{id}/upload` | Status transfer/sinkronisasi dan bytes part terkonfirmasi. |
| POST `/api/v1/recordings/{id}/retry` | Reset waktu retry rekaman READY tanpa mengubah identitas. |
| GET `/api/v1/recordings/{id}/file` | MKV READY/UPLOADED dengan Range support, tanpa mengekspos path lokal. |

State upload UI: WAITING_MAPPING, HASHING, QUEUED, UPLOADING, SYNCING, RETRYING, ACCESS_DENIED, UPLOADED. State saving engine tetap terpisah. File final tidak dihapus oleh upload, termasuk setelah sukses.

`container` accepts `mkv` for this engine. The immutable cloud object key retains the suffix; S3 content type and download filename follow it. No new database column is required. MKV engines reject an upload response with a non-MKV object key, retaining the local file until the cloud is upgraded.

Engine hanya menerima NVR HEVC/H.265 dan file MKV. Acceptance: scripts/test-native.ps1 -NVRRTSPURL <NVR_RTSP_URL> beserta binary; sumber NVR nyata wajib tersedia. Unit test logika tetap ada.
