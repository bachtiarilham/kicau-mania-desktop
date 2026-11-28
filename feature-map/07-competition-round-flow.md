# Aturan Ronde Operator dan Juri

Dokumen ini menjadi acuan alur bisnis perlombaan dan pengujian ronde Operator–Juri.

## Jadwal rotasi

Ronde berlangsung 10 menit sejak operator menekan **START**. Delapan menit pertama dibagi menjadi 4 shift masing-masing 2 menit. Dua menit terakhir dibagi menjadi 4 giliran masing-masing 30 detik. Setiap juri menempati zona yang berbeda pada setiap interval.

| Waktu berlalu | Tahap | Juri 1 | Juri 2 | Juri 3 | Juri 4 |
|---|---|---|---|---|---|
| 00:00–02:00 | Shift 1 | A | B | C | D |
| 02:00–04:00 | Shift 2 | B | C | D | A |
| 04:00–06:00 | Shift 3 | C | D | A | B |
| 06:00–08:00 | Shift 4 | D | A | B | C |
| 08:00–08:30 | Putaran akhir | D | A | B | C |
| 08:30–09:00 | Putaran akhir | A | B | C | D |
| 09:00–09:30 | Putaran akhir | B | C | D | A |
| 09:30–10:00 | Putaran akhir | C | D | A | B |

- Pada **08:00**, setiap juri memulai putaran akhir di zona Shift 4. Perpindahan zona pertama pada putaran akhir terjadi pada **08:30**.
- Grid kamera berpindah otomatis sesuai jadwal. Selama ronde, layar juri tidak menyediakan pemilih zona manual.
- Checklist zona setiap juri otomatis tercentang saat shift 2 menit di zona tersebut selesai. Pada **08:00**, keempat zona sudah tercentang untuk setiap juri. Checklist zona tetap sama sepanjang putaran akhir. Checklist kesiapan gantangan tetap dipegang operator.
- Pada **10:00**, timer dan rotasi berakhir. Operator tetap memegang kontrol **STOP** ronde.

## Sinkronisasi dan skenario uji

- Server arena sekaligus PC operator `192.168.1.6` menjadi sumber status ronde dan waktu. PC juri mengikuti ronde aktif, hitung mundur, zona, dan checklist dari server lewat LAN.
- Di layar login desktop PC juri, buka **Server → Ubah**. Isi API arena dengan `http://192.168.1.6:8090` dan RTSP MediaMTX dengan `rtsp://192.168.1.6:8554`. Simpan, buka ulang desktop, lalu login dengan akun juri yang ditugaskan pada ronde cloud.
- Di `.env` Local App operator, gunakan `RTSP_ADDR=0.0.0.0:8554`. Isi `JUDGE_READ_CIDRS` dengan alamat setiap PC juri dalam format `192.168.1.58/32`, dipisahkan koma. Daftar ini berlaku untuk status ronde, daftar kamera, metrik audio, dan RTSP. Buka ulang Local App setelah mengedit daftar.
- Uji awal memakai PC operator `192.168.1.6` dan PC juri `192.168.1.58`. Jalankan satu ronde penuh. Bandingkan timer di kedua layar, cocokkan grid juri dengan matriks saat mencapai setiap batas waktu, dan cek centang otomatis pada **02:00**, **04:00**, **06:00**, dan **08:00**.
- Putus dan sambungkan ulang PC juri ke LAN di tengah ronde. Setelah tersambung, layar juri mengikuti waktu dan zona aktif dari server.
- Validasi matriks agar keempat juri selalu menempati empat zona berbeda. Skenario operasional akhir memakai **1 PC operator dan 4 PC juri**.
