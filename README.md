# AI Kurikulum Koding & AI Web Application (v1.0.1)

[![Live Website](https://img.shields.io/badge/Website-aikurikulum.siapdigital.my.id-blue?style=for-the-badge&logo=googlechrome)](https://aikurikulum.siapdigital.my.id/)
[![Release](https://img.shields.io/badge/Release-v1.0.1-emerald?style=for-the-badge&logo=github)](https://github.com/sodikinnaa/coding-agent-web/releases/tag/v1.0.1)
[![Payment Gateway](https://img.shields.io/badge/Payment-Mayar.id-purple?style=for-the-badge)](https://web.mayar.id/)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)

Aplikasi web modern berbasis AI RAG (Retrieval-Augmented Generation) & Direct PDF Vision Engine untuk menjawab pertanyaan, evaluasi, serta kuis interaktif Kurikulum Koding & Artificial Intelligence (AI) jenjang SD, SMP, hingga SMA.

🌐 **Live Website**: [https://aikurikulum.siapdigital.my.id/](https://aikurikulum.siapdigital.my.id/)  
📦 **GitHub Repository**: [https://github.com/sodikinnaa/coding-agent-web](https://github.com/sodikinnaa/coding-agent-web)  
💳 **Payment Gateway**: [Mayar.id Dynamic QRIS Payment](https://web.mayar.id/)

---

## ⚡ Install Cepat via `curl` (One-Line Installer & Auto-Updater)

Jalankan perintah satu baris berikut di server Linux Anda untuk mengunduh, memperbarui, mengompilasi, dan mengaktifkan service secara otomatis:

```bash
curl -fsSL https://raw.githubusercontent.com/sodikinnaa/coding-agent-web/main/install.sh | bash
```

---

## ✨ Fitur Utama

- 🧠 **Direct PDF Vision RAG Engine**: Analisis berkas PDF buku panduan koding & AI secara presisi lengkap dengan sitasi nomor halaman buku.
- 🎯 **Kuis Interaktif & Papan Peringkat (Leaderboard)**:
  - 8 Kategori Kuis berbasis modul PDF kurikulum (SD, SMP, SMA, Kombinatorik, & Coding Dasar).
  - Verifikasi jawaban dilakukan secara **Server-Side Only** (aman dari inspeksi DevTools).
  - Papan Peringkat publik dengan enkripsi sensor privasi email (`so***aa@gmail.com`).
- 💳 **Sistem Billing QRIS Mayar.id & Kredit Chat**:
  - Integrasi Payment Gateway resmi via [Mayar.id](https://web.mayar.id/) untuk penerimaan pembayaran QRIS Instant.
  - Batas kredit harian gratis dengan dukungan Upgrade Instant via QRIS Mayar.id.
  - Multi-fallback Webhook handler untuk pencocokan transaksi otomatis.
- 🔐 **Keamanan Enkripsi AES-256-GCM & Captcha**:
  - Token API Key & Mayar Key tersimpan aman dengan enkripsi militer **AES-256-GCM**.
  - Modal pendaftaran akun dilengkapi soal matematika acak interaktif untuk mencegah spam/bot.
- 📱 **Desain UI Responsif (Mobile & Desktop)**:
  - Tampilan Obrolan Desktop terpusat (*Centered Reading Stream*) & Kartu Saran Pertanyaan (*Quick Suggestion Cards*).
  - Drawer Sidebar Mobile modern dengan auto-close dan topbar responsif.
  - Custom Alert & Notification Modal System replacing raw browser popups.

---

## 🛠️ Panduan Instalasi Manual

### Persyaratan Sistem
- Linux (Ubuntu/Debian/CentOS/OpenCloudOS)
- Go 1.21+ compiler
- SQLite3

### Langkah Instalasi
```bash
# 1. Clone repository
git clone https://github.com/sodikinnaa/coding-agent-web.git
cd coding_agent_web

# 2. Salin file konfigurasi contoh
cp config.example.json config.json

# 3. Jalankan pengujian unit (Unit Test)
go test -v ./...

# 4. Kompilasi aplikasi
go build -o coding_agent_app main.go

# 5. Jalankan server
./coding_agent_app
```

Server akan berjalan secara lokal di `http://localhost:8097`.

---

## 🧪 Eksekusi Unit Test

Aplikasi ini dilengkapi dengan paket pengujian unit internal (`internal/auth`, `internal/db`, `internal/service`, `internal/config`) dengan rating **100% PASS**:

```bash
go test -v ./...
```

---

## 📄 Lisensi

Proyek ini dirilis di bawah lisensi [MIT License](LICENSE).
