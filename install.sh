#!/bin/bash
set -e

INSTALL_DIR="/opt/coding_agent_web"
if [ ! -d "$INSTALL_DIR" ]; then
    if [ -d "/root/coding_agent_web" ]; then
        INSTALL_DIR="/root/coding_agent_web"
    fi
fi

# Set fast Go proxy for ultra-fast dependency downloads
export GOPROXY="https://proxy.golang.org,direct"

echo "======================================================"
echo "  AI Kurikulum Koding & AI - Installer & Updater"
echo "======================================================"

# Check if application is already installed on the server
if [ -d "$INSTALL_DIR/.git" ] || [ -f "$INSTALL_DIR/coding_agent_app" ]; then
    echo "🔍 Terdeteksi: Aplikasi sudah terinstall di server ($INSTALL_DIR)."
    echo "🔄 Menjalankan Mode Pembaruan Cepat (FAST UPDATE MODE)..."
    echo "------------------------------------------------------"

    cd "$INSTALL_DIR"

    echo "[1/4] Mengambil pembaruan kode terbaru dari GitHub (git pull)..."
    git fetch origin main
    git reset --hard origin/main

    echo "[2/4] Memeriksa compiler Go..."
    if ! command -v go &> /dev/null; then
        echo "❌ Compiler Go tidak ditemukan! Harap install Go 1.21+ terlebih dahulu."
        exit 1
    fi

    echo "[3/4] Mengompilasi ulang biner aplikasi Go..."
    go build -ldflags="-s -w" -o coding_agent_app main.go

    echo "[4/4] Merestart service systemd (coding-knowledge.service)..."
    if command -v systemctl &> /dev/null; then
        sudo systemctl restart coding-knowledge.service || true
    fi

    echo ""
    echo "======================================================"
    echo "  Pembaruan Berhasil! (Update Successful 🎉)"
    echo "  Live Website: https://aikurikulum.siapdigital.my.id/"
    echo "  Local Server: http://localhost:8097"
    echo "======================================================"
    exit 0
fi

echo "✨ Terdeteksi: Server baru belum terpasang aplikasi."
echo "🚀 Menjalankan Mode Instalasi Baru (FRESH INSTALL MODE)..."
echo "------------------------------------------------------"

INSTALL_DIR="/opt/coding_agent_web"
echo "[1/4] Menyiapkan direktori instalasi di $INSTALL_DIR..."
sudo mkdir -p "$INSTALL_DIR"
sudo chown -R $USER:$USER "$INSTALL_DIR"

echo "[2/4] Mengkloning repository resmi dari GitHub..."
git clone --depth 1 https://github.com/sodikinnaa/coding-agent-web.git "$INSTALL_DIR"

cd "$INSTALL_DIR"

echo "[3/4] Mengompilasi biner aplikasi Go..."
if ! command -v go &> /dev/null; then
    echo "❌ Compiler Go tidak ditemukan! Harap install Go 1.21+ terlebih dahulu."
    exit 1
fi

go build -ldflags="-s -w" -o coding_agent_app main.go

echo "[4/4] Memasang service systemd..."
sudo tee /etc/systemd/system/coding-knowledge.service > /dev/null <<EOF
[Unit]
Description=Koding and AI Knowledge Web Application
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/coding_agent_app
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

if command -v systemctl &> /dev/null; then
    sudo systemctl daemon-reload
    sudo systemctl enable coding-knowledge.service
    sudo systemctl restart coding-knowledge.service
fi

echo ""
echo "======================================================"
echo "  Instalasi Baru Berhasil! (Fresh Install Successful 🎉)"
echo "  Live Website: https://aikurikulum.siapdigital.my.id/"
echo "  Local Server: http://localhost:8097"
echo "======================================================"
