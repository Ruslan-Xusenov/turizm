#!/bin/bash
# =============================================================
# LOKAL DEPLOY SCRIPTI
# Ishlatish: ./deploy.sh "commit xabari"
# Misol:     ./deploy.sh "Yangi funksiya qo'shildi"
# =============================================================

set -e  # Xato bo'lsa to'xtasin

# --- SOZLAMALAR ---
SERVER_IP="46.224.133.140"
SERVER_USER="root"        # Serveringizda kim bilan kirasiz
COMMIT_MSG="${1:-Auto deploy: $(date '+%Y-%m-%d %H:%M')}"

echo "============================================="
echo "  TURIZM LOYIHASINI DEPLOY QILISH"
echo "============================================="

# 1. GitHub ga push qilish
echo ""
echo "[1/3] GitHub ga yuklanmoqda..."
cd "$(dirname "$0")"

git add -A
git commit -m "$COMMIT_MSG" || echo "  -> Yangi o'zgarish yo'q, commit o'tkazib yuborildi"
git push origin main
echo "  -> GitHub ga muvaffaqiyatli yuklandi!"

# 2. Serverda deploy qilish
echo ""
echo "[2/3] Server ($SERVER_IP) yangilanmoqda..."
ssh -o StrictHostKeyChecking=no $SERVER_USER@$SERVER_IP "bash /opt/turizm/update.sh"
echo "  -> Server yangilandi!"

# 3. Servis holatini tekshirish
echo ""
echo "[3/3] Servis holati tekshirilmoqda..."
sleep 2
ssh $SERVER_USER@$SERVER_IP "systemctl is-active turizm && echo '  -> Dastur ISHLAYAPTI ✓' || echo '  -> XATO: Dastur ishlamayapti!'"

echo ""
echo "============================================="
echo "  DEPLOY MUVAFFAQIYATLI YAKUNLANDI!"
echo "  Sayt: http://$SERVER_IP"
echo "============================================="