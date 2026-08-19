# 🤖 Bot WhatsApp ASCII Informatika (Go + Whatsmeow)

Bot WhatsApp resmi Laboratorium Informatika ASCII (Universitas Mulawarman) yang dibangun menggunakan Go dan library [whatsmeow](https://github.com/tulir/whatsmeow), siap terintegrasi penuh dengan portal web **`ascii-if`**.

---

## 🌟 Fitur

### 1. Interaksi Chat WhatsApp (Mahasiswa & Aslab)
- `!help` / `!menu` : Menampilkan panduan dan daftar perintah bot.
- `!jadwal` : Mengambil jadwal praktikum aktif secara langsung dari web `ascii-if`.
- `!kelas [kode]` : Mencari info kelas, ruangan, dosen, dan aslab pendamping.
- `!modul [matkul]` : Mencari modul praktikum dan link unduhan resmi.
- `!pengumuman` : Menampilkan pengumuman lab terkini.
- `!berita` : Menampilkan update berita praktikum terbaru.
- `!kontak` / `!aslab` : Daftar nomor kontak asisten & koordinator praktikum.
- `!aturan` : Tata tertib dan aturan praktikum lab.
- `!faq` : Pertanyaan seputar praktikum yang sering ditanyakan.
- `!ping` : Cek latency dan status koneksi bot ke web portal.

### 2. Integrasi Dua Arah dengan Web `ascii-if`
- **Dari Bot ke Web**: Bot mengambil data via REST/OpenAPI di `/api/*` pada web `ascii-if` menggunakan otentikasi header `x-api-key`.
- **Dari Web ke WhatsApp (Webhook Server)**: Bot membuka REST API lokal (default port `8080`) sehingga web `ascii-if` bisa memicu pengiriman pesan / broadcast otomatis:
  - `POST /api/v1/send-message` : Kirim pesan WA ke nomor tertentu.
  - `POST /api/v1/broadcast` : Broadcast notifikasi ke banyak nomor / grup kelas.
  - `GET /api/v1/health` : Health check status bot & web.

---

## 📁 Struktur Direktori

```
botwa_go_ascii/
├── cmd/
│   └── bot/
│       └── main.go           # Entry point utama
├── config/
│   └── config.go            # Loader konfigurasi (.env)
├── pkg/
│   ├── asciiapi/            # Client HTTP ke API web ascii-if
│   │   ├── client.go        # HTTP methods & endpoints
│   │   └── models.go        # Data structures
│   ├── bot/                 # Engine WhatsApp
│   │   ├── bot.go           # Inisialisasi whatsmeow & SQLite
│   │   ├── formatter.go     # WhatsApp markdown formatting
│   │   ├── qr.go            # Terminal QR renderer
│   │   └── router.go        # Command dispatching
│   └── server/              # Webhook REST Server
│       └── server.go        # REST endpoint /send-message & /broadcast
├── .env.example
├── .gitignore
└── README.md
```

---

## 🚀 Cara Menjalankan

### 1. Konfigurasi `.env`
Salin `.env.example` ke `.env` dan sesuaikan nilainya:
```env
# URL API Web ascii-if (lokal atau domain)
ASCII_API_URL=http://localhost:3000/api
ASCII_API_KEY=your-secret-key-di-ascii-if
ASCII_WEB_URL=https://ascii.web.id

# Server Webhook Bot
ENABLE_SERVER=true
SERVER_PORT=8080
BOT_SECRET=ascii-secret-bot-key
```

### 2. Jalankan Bot
```bash
# Menjalankan langsung
go run ./cmd/bot

# Atau build binary
go build -o bot_ascii.exe ./cmd/bot
./bot_ascii.exe
```

### 3. Autentikasi WhatsApp (Scan QR)
1. Saat bot pertama kali dijalankan, QR Code akan muncul di terminal.
2. Buka aplikasi WhatsApp di HP Anda.
3. Masuk ke **Menu (titik tiga) / Pengaturan > Perangkat Tertaut > Tautkan Perangkat**.
4. Scan QR code yang tampil di terminal.
5. Sesi akan otomatis tersimpan di file `whatsapp.db`. Anda tidak perlu scan ulang setiap kali restart.

---

## 📡 Contoh Memanggil Webhook dari Web `ascii-if`

Web `ascii-if` dapat mengirim pesan WA melalui endpoint bot:

```bash
# Kirim Pesan Tunggal
curl -X POST http://localhost:8080/api/v1/send-message \
  -H "Content-Type: application/json" \
  -H "X-Bot-Secret: ascii-secret-bot-key" \
  -d '{
    "to": "628123456789",
    "message": "Halo! Praktikum Pemrograman Web Kelas A dimulai 15 menit lagi di Lab 1."
  }'
```

```bash
# Broadcast ke Banyak Kontak / Grup
curl -X POST http://localhost:8080/api/v1/broadcast \
  -H "Content-Type: application/json" \
  -H "X-Bot-Secret: ascii-secret-bot-key" \
  -d '{
    "recipients": ["628123456789", "628987654321"],
    "message": "📢 *Pengumuman:* Modul 3 telah diunggah di portal ascii.web.id/modul"
  }'
```
