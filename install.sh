#!/bin/bash
set -e

INSTALL_DIR="/opt/coding_agent_web"
if [ ! -d "$INSTALL_DIR" ]; then
    if [ -d "/root/coding_agent_web" ]; then
        INSTALL_DIR="/root/coding_agent_web"
    fi
fi

echo "======================================================"
echo "  AI Kurikulum Koding & AI - Direct Binary Installer"
echo "======================================================"

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

sudo mkdir -p "$INSTALL_DIR"
if [ -n "$USER" ] && [ "$USER" != "root" ]; then
    sudo chown -R $USER:$USER "$INSTALL_DIR" || true
fi

cd "$INSTALL_DIR"

RELEASE_URL="https://github.com/sodikinnaa/coding-agent-web/releases/latest/download/${BINARY_NAME}"
echo "📥 Mengunduh biner aplikasi langsung dari GitHub Release..."
echo "🔗 $RELEASE_URL"

if curl -fsSL "$RELEASE_URL" -o coding_agent_app.tmp; then
    mv coding_agent_app.tmp coding_agent_app
    chmod +x coding_agent_app
    echo "✅ Biner berhasil diunduh!"
else
    echo "❌ Gagal mengunduh biner rilis ($RELEASE_URL)."
    echo "Harap pastikan GitHub Release terbaru sudah dipublikasikan."
    exit 1
fi

# Download config.example.json as config.json if config.json does not exist
if [ ! -f "config.json" ]; then
    echo "⚙️ Menyiapkan config.json default..."
    curl -fsSL "https://raw.githubusercontent.com/sodikinnaa/coding-agent-web/main/config.example.json" -o config.json || true
fi

echo "⚙️ Memasang / Merestart service systemd (coding-knowledge.service)..."
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
echo "  Proses Berhasil! (Successful 🎉)"
echo "  Live Website: https://aikurikulum.siapdigital.my.id/"
echo "  Local Server: http://localhost:8097"
echo "======================================================"
