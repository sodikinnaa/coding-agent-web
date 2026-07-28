#!/bin/bash
set -e

echo "======================================================"
echo "  AI Kurikulum Koding & AI - One-Line Installer v1.0.0"
echo "======================================================"

INSTALL_DIR="/opt/coding_agent_web"

echo "[1/4] Preparing installation directory at $INSTALL_DIR..."
sudo mkdir -p "$INSTALL_DIR"
sudo chown -R $USER:$USER "$INSTALL_DIR"

echo "[2/4] Cloning latest codebase from GitHub..."
git clone --depth 1 https://github.com/sodikinnaa/coding-agent-web.git "$INSTALL_DIR" || (cd "$INSTALL_DIR" && git pull origin main)

cd "$INSTALL_DIR"

echo "[3/4] Building Go application binary..."
if ! command -v go &> /dev/null; then
    echo "Go compiler not found! Please install Go 1.21+ first."
    exit 1
fi

go build -o coding_agent_app main.go

echo "[4/4] Setting up systemd service..."
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
sudo systemctl restart coding-knowledge.service

echo ""
echo "======================================================"
echo "  Installation Successful! 🎉"
echo "  Live Website: https://aikurikulum.siapdigital.my.id/"
echo "  Local Server: http://localhost:8097"
echo "======================================================"
