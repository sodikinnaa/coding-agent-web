#!/bin/bash
set -e

INSTALL_DIR="/opt/coding_agent_web"
if [ ! -d "$INSTALL_DIR" ]; then
    if [ -d "/root/coding_agent_web" ]; then
        INSTALL_DIR="/root/coding_agent_web"
    fi
fi

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
BINARY_NAME=""

if [ "$OS" = "linux" ]; then
    if [ "$ARCH" = "x86_64" ]; then
        BINARY_NAME="coding_agent_app_linux_amd64"
    elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
        BINARY_NAME="coding_agent_app_linux_arm64"
    fi
elif [ "$OS" = "darwin" ]; then
    if [ "$ARCH" = "x86_64" ]; then
        BINARY_NAME="coding_agent_app_darwin_amd64"
    elif [ "$ARCH" = "arm64" ]; then
        BINARY_NAME="coding_agent_app_darwin_arm64"
    fi
fi

if [ -z "$BINARY_NAME" ]; then
    echo "❌ Arsitektur sistem ($OS/$ARCH) tidak didukung."
    exit 1
fi

echo "======================================================"
echo "  AI Kurikulum Koding & AI - Direct Binary Installer"
echo "======================================================"

IS_UPDATE=0
if [ -f "$INSTALL_DIR/coding_agent_app" ] || [ -f "$INSTALL_DIR/config.json" ]; then
    IS_UPDATE=1
fi

sudo mkdir -p "$INSTALL_DIR"
if [ -n "$USER" ] && [ "$USER" != "root" ]; then
    sudo chown -R $USER:$USER "$INSTALL_DIR" || true
fi

cd "$INSTALL_DIR"

if [ $IS_UPDATE -eq 1 ]; then
    echo "🔍 Terdeteksi: Aplikasi sudah terinstall di server ($INSTALL_DIR)."
    echo "🔄 Menjalankan Mode Pembaruan Cepat (FAST UPDATE MODE)..."
    echo "------------------------------------------------------"
else
    echo "✨ Terdeteksi: Server baru belum terpasang aplikasi."
    echo "🚀 Menjalankan Mode Instalasi Baru (FRESH INSTALL MODE)..."
    echo "------------------------------------------------------"
fi

RELEASE_URL="https://github.com/sodikinnaa/coding-agent-web/releases/latest/download/${BINARY_NAME}"
echo "📥 Mengunduh biner aplikasi dari GitHub Release..."

if curl -fsSL "$RELEASE_URL" -o coding_agent_app.tmp; then
    mv coding_agent_app.tmp coding_agent_app
    chmod +x coding_agent_app
    echo "✅ Biner aplikasi berhasil diperbarui!"
else
    echo "❌ Gagal mengunduh biner rilis ($RELEASE_URL)."
    exit 1
fi

# Manage Config & Admin Credentials
ADMIN_USER="admin"
ADMIN_PASS=""

if [ -f "config.json" ]; then
    ADMIN_PASS=$(grep -o '"admin_password": *"[^"]*"' config.json | head -n1 | cut -d'"' -f4 || true)
fi

if [ -z "$ADMIN_PASS" ]; then
    # Generate random 12-character password for fresh install
    ADMIN_PASS=$(tr -dc 'a-zA-Z0-9' < /dev/urandom 2>/dev/null | head -c 12 || echo "Admin$(date +%s)")

    if [ ! -f "config.json" ]; then
        if [ -f "config.example.json" ]; then
            cp config.example.json config.json
        else
            curl -fsSL "https://raw.githubusercontent.com/sodikinnaa/coding-agent-web/main/config.example.json" -o config.json 2>/dev/null || true
        fi
    fi

    if [ -f "config.json" ]; then
        sed -i "s/\"admin_password\": *\"[^\"]*\"/\"admin_password\": \"$ADMIN_PASS\"/g" config.json
    fi
fi

# Systemd Service Installation / Restart
if command -v systemctl &> /dev/null; then
    if [ ! -f "/etc/systemd/system/coding-knowledge.service" ]; then
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
        sudo systemctl daemon-reload
        sudo systemctl enable coding-knowledge.service
    fi

    sudo systemctl restart coding-knowledge.service || true
fi

echo ""
echo "======================================================"
if [ $IS_UPDATE -eq 1 ]; then
    echo " 🎉 Pembaruan Berhasil! (Update Successful)"
    echo " ------------------------------------------------------"
    echo " 🔑 Admin Credentials:"
    echo "    Username : $ADMIN_USER"
    echo "    Password : $ADMIN_PASS (Tetap / Tidak Berubah)"
else
    echo " 🎉 Instalasi Baru Berhasil! (Fresh Install Successful)"
    echo " ------------------------------------------------------"
    echo " 🔑 Admin Credentials (Dibuat Acak Otomatis):"
    echo "    Username : $ADMIN_USER"
    echo "    Password : $ADMIN_PASS"
fi
echo "======================================================"
