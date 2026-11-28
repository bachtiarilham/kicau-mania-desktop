# Feature Map - Rekaman & Arsip

Dokumen ini mencatat alur bisnis target untuk rekaman lomba. Di desktop, operator dan admin membuka **Rekaman & Arsip** untuk melihat, memutar, dan memantau status upload rekaman lokal.

## Alur setelah ronde

1. Setelah ronde selesai dan video siap, Local App otomatis mengunggah rekaman ke Cloud App.
2. Saat operator menekan **START** untuk ronde berikutnya, Local App menjeda upload yang sedang berjalan dan menahan upload baru. Antrean dilanjutkan otomatis setelah ronde berhenti. Upload hanya berjalan saat tidak ada ronde yang sedang berlangsung.
3. File asli disimpan di PC operator setelah upload berhasil. Status upload dapat dipantau di daftar **Rekaman Lokal**.
4. Cloud App memicu pengiriman tautan WhatsApp. Peserta membuka tautan di browser untuk menonton dan mengunduh video melalui Cloud App.
5. Tautan peserta berlaku selama lima hari sejak video berhasil diunggah. Setelah lima hari, Cloud App menghentikan akses dan menghapus video dari penyimpanan Cloud. Rekaman asli di PC operator tetap tersedia.

Dalam distribusi rekaman, Local App menangani upload dan desktop menampilkan statusnya. Fitur unduh peserta dan pemicu pengiriman WhatsApp tersedia hanya di Cloud App.

## Panduan desktop

| Fitur | Aksi |
|---|---|
| Daftar rekaman | Buka **Rekaman & Arsip** untuk melihat video lokal dan status upload. |
| Pemutaran lokal | Tekan **Putar** pada baris rekaman yang ingin ditonton. |
| Folder lokal | Buka direktori rekaman dari **Buka di Explorer**. |

Skenario uji lintas aplikasi mencakup memulai ronde kedua saat video ronde pertama masih diunggah, lalu memastikan upload terjeda dan lanjut setelah ronde berhenti. Setelah upload berhasil, periksa bahwa file asli tetap ada di PC operator.
