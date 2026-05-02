#!/bin/bash
# =============================================================
# SERVER O'RNATISH SCRIPTI
# Serverda FAQAT BIR MARTA ishlatiladi:
#   bash server-setup.sh
#
# Kerakli ma'lumotlar:
#   - Ubuntu 20.04+ server
#   - root yoki sudo huquqi
# =============================================================

set -e

REPO_URL="https://github.com/Ruslan-Xusenov/turizm"
APP_DIR="/opt/turizm"
SERVICE_NAME="turizm"
APP_PORT="8080"

echo "============================================="
echo "  SERVER O'RNATISH BOSHLANDI"
echo "============================================="

# 1. Tizim paketlarini yangilash
echo ""
echo "[1/7] Tizim yangilanmoqda..."
apt-get update -qq
apt-get install -y git gcc curl sqlite3

# 2. Go o'rnatish
echo ""
echo "[2/7] Go (Golang) o'rnatilmoqda..."
if ! command -v go &> /dev/null; then
    GO_VERSION="1.22.3"
    curl -sL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    export PATH=$PATH:/usr/local/go/bin
    echo "  -> Go $GO_VERSION o'rnatildi!"
else
    echo "  -> Go allaqachon o'rnatilgan: $(go version)"
fi
export PATH=$PATH:/usr/local/go/bin

# 3. Loyihani GitHub dan yuklab olish
echo ""
echo "[3/7] Loyiha GitHub dan yuklanmoqda..."
if [ -d "$APP_DIR" ]; then
    rm -rf "$APP_DIR"
fi
git clone "$REPO_URL" "$APP_DIR"
echo "  -> Loyiha $APP_DIR ga yuklandi!"

# 4. .env fayl yaratish
echo ""
echo "[4/7] .env fayl sozlanmoqda..."
if [ ! -f "$APP_DIR/.env" ]; then
    cat > "$APP_DIR/.env" << 'ENVFILE'
# === ASOSIY SOZLAMALAR ===
PORT=8080
SESSION_SECRET=your_very_secret_session_key_change_this

# === GOOGLE OAUTH (ixtiyoriy) ===
GOOGLE_KEY=
GOOGLE_SECRET=
CALLBACK_URL=https://imanturizm.uz/auth/google/callback

# === JWT SIRLI KALIT ===
JWT_SECRET=your_jwt_secret_key_change_this

# === EMAIL (ixtiyoriy) ===
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=

# === MA'LUMOTLAR BAZASI ===
DB_PATH=/opt/turizm/turizm.db
ENVFILE
    echo "  -> .env fayl yaratildi. /opt/turizm/.env ni tahrirlang!"
else
    echo "  -> .env fayl allaqachon mavjud"
fi

# 5. Dasturni build qilish
echo ""
echo "[5/7] Dastur build qilinmoqda..."
cd "$APP_DIR"
go mod download
go build -ldflags="-s -w" -o turizm_server .
chmod +x turizm_server
echo "  -> Build muvaffaqiyatli!"

# 6. update.sh scripti yaratish (keyingi deploylar uchun)
echo ""
echo "[6/7] Update scripti yaratilmoqda..."
cat > "$APP_DIR/update.sh" << 'UPDATESCRIPT'
#!/bin/bash
set -e
APP_DIR="/opt/turizm"
export PATH=$PATH:/usr/local/go/bin

echo "-> Kod yangilanmoqda..."
cd "$APP_DIR"
git pull origin main

echo "-> Qayta build qilinmoqda..."
go build -ldflags="-s -w" -o turizm_server .

echo "-> Servis qayta ishga tushirilmoqda..."
systemctl restart turizm
systemctl status turizm --no-pager

echo "-> Yangilash yakunlandi!"
UPDATESCRIPT
chmod +x "$APP_DIR/update.sh"

# 7. Systemd servis yaratish
echo ""
echo "[7/7] Systemd servis yaratilmoqda..."
cat > "/etc/systemd/system/${SERVICE_NAME}.service" << SERVICEFILE
[Unit]
Description=Turizm Go Web Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${APP_DIR}
ExecStart=${APP_DIR}/turizm_server
Restart=always
RestartSec=5
EnvironmentFile=${APP_DIR}/.env
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
SERVICEFILE

systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl start "$SERVICE_NAME"

# Holat tekshirish
sleep 2
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo ""
    echo "============================================="
    echo "  O'RNATISH MUVAFFAQIYATLI YAKUNLANDI! ✓"
    echo ""
    echo "  Sayt:    https://imanturizm.uz"
    echo "  Log:     journalctl -u $SERVICE_NAME -f"
    echo "  To'xtat: systemctl stop $SERVICE_NAME"
    echo "  Yangilash: bash $APP_DIR/update.sh"
    echo "============================================="
else
    echo ""
    echo "  XATO: Servis ishlamadi!"
    echo "  Logni ko'rish: journalctl -u $SERVICE_NAME -n 50"
    exit 1
fi
