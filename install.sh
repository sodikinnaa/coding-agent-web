#!/bin/bash
set -e

INSTALL_DIR="/opt/coding_agent_web"
if [ ! -d "$INSTALL_DIR" ]; then
    if [ -d "/root/coding_agent_web" ]; then
        INSTALL_DIR="/root/coding_agent_web"
    fi
fi

# Set fast Go proxy for ultra-fast dependency downloads (fallback mode)
export GOPROXY="https://proxy.golang.org,direct"

echo "======================================================"
echo "  AI Kurikulum Koding & AI - Installer & Updater"
echo "======================================================"

get_binary() {
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

    if [ -n "$BINARY_NAME" ]; then
        RELEASE_URL="https://github.com/sodikinnaa/coding-agent-web/releases/latest/download/${BINARY_NAME}"
        echo "📥 Mengunduh biner rilis pre-compiled CI/CD (${BINARY_NAME})..."
        if curl -fsSL "$RELEASE_URL" -o coding_agent_app 2>/dev/null; then
            chmod +x coding_agent_app
            echo "✅ Biner pre-compiled CI/CD berhasil diunduh!"
            return 0
        else
            echo "⚠️  Biner pre-compiled belum tersedia di Release terbaru. Mencoba opsi kompilasi lokal..."
        fi
    fi

    if command -v go &> /dev/null; then
        echo "⚙️  Mengompilasi ulang biner aplikasi Go dari sumber..."
        go build -ldflags="-s -w" -o coding_agent_app main.go
        chmod +x coding_agent_app
        return 0
    else
        echo "❌ Biner rilis CI/CD tidak dapat diunduh dan compiler Go tidak ditemukan di server!"
        exit 1
    fi
}

# Check if application is already installed on the server
if [ -d "$INSTALL_DIR/.git" ] || [ -f "$INSTALL_DIR/coding_agent_app" ]; then
    echo "🔍 Terdeteksi: Aplikasi sudah terinstall di server ($INSTALL_DIR)."
    echo "🔄 Menjalankan Mode Pembaruan Cepat (FAST UPDATE MODE)..."
    echo "------------------------------------------------------"

    cd "$INSTALL_DIR"

    echo "[1/3] Mengambil pembaruan kode terbaru dari GitHub (git pull)..."
    git fetch origin main || true
    git reset --hard origin/main || true

    echo "[2/3] Memperbarui biner aplikasi..."
    get_binary

    echo "[3/3] Merestart service systemd (coding-knowledge.service)..."
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

echo "[3/4] Menyiapkan biner aplikasi Go..."
get_binary

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
