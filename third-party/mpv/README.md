# mpv Windows dependency

Build yang dikunci tercantum pada manifest.json, termasuk SHA-256 arsip. Sumber binary adalah first-party mpv CI, Windows amd64 full build; ini build development yang dipin, bukan klaim rilis stabil mpv.

`scripts/fetch-deps.ps1` memverifikasi checksum sebelum ekstraksi. Paket portable membawa mpv.exe dan LICENSE.GPL dari arsip yang sama. Source dan build recipes: https://github.com/mpv-player/mpv serta release URL pada manifest. Dependency berjalan sebagai proses terpisah dan tidak mendaftarkan file association Windows; script register/unregister dari arsip tidak dijalankan.
