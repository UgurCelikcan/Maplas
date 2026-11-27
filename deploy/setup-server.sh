#!/usr/bin/env bash
# Maplas - Ubuntu sunucu kurulum betiği (22.04 / 24.04)
#
# Kullanım (sunucuda root olarak):
#   curl -fsSL https://raw.githubusercontent.com/UgurCelikcan/Maplas/main/deploy/setup-server.sh -o setup-server.sh
#   bash setup-server.sh
#
# Başka bir branch için:  BRANCH=fix/require-secrets bash setup-server.sh
#
# Betik tekrar çalıştırılabilir: var olan .env dosyasına ve verilere dokunmaz.

set -euo pipefail

REPO_URL="${REPO_URL:-https://github.com/UgurCelikcan/Maplas.git}"
BRANCH="${BRANCH:-main}"
APP_DIR="${APP_DIR:-/opt/maplas}"
DOMAIN="${DOMAIN:-maplas.net}"

if [ "$(id -u)" -ne 0 ]; then
  echo "Bu betiği root olarak çalıştırın:  sudo bash setup-server.sh"
  exit 1
fi

echo "==> 1/6 Sistem güncelleniyor"
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get upgrade -y
apt-get install -y ca-certificates curl git ufw fail2ban unattended-upgrades openssl dnsutils

echo "==> 2/6 Otomatik güvenlik güncellemeleri açılıyor"
dpkg-reconfigure -f noninteractive unattended-upgrades

echo "==> 3/6 Güvenlik duvarı (sadece SSH, HTTP, HTTPS)"
ufw default deny incoming
ufw default allow outgoing
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 443/udp
ufw --force enable
systemctl enable --now fail2ban

echo "==> 4/6 Docker kuruluyor"
if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sh
fi
systemctl enable --now docker

echo "==> 5/6 Proje indiriliyor: $APP_DIR ($BRANCH)"
if [ -d "$APP_DIR/.git" ]; then
  git -C "$APP_DIR" fetch origin
  git -C "$APP_DIR" checkout "$BRANCH"
  git -C "$APP_DIR" pull --ff-only origin "$BRANCH"
else
  git clone --branch "$BRANCH" "$REPO_URL" "$APP_DIR"
fi
cd "$APP_DIR"

if [ ! -f .env ]; then
  echo "    .env oluşturuluyor (rastgele şifrelerle)"
  ADMIN_CODE="$(openssl rand -hex 12)"
  cat > .env <<EOF
DB_HOST=db
DB_PORT=5432
DB_USER=maplas_user
DB_PASSWORD=$(openssl rand -hex 24)
DB_NAME=maplas_db
JWT_SECRET=$(openssl rand -hex 32)
ADMIN_SECRET=${ADMIN_CODE}
SERVER_PORT=8080
VITE_API_URL=https://api.${DOMAIN}
EOF
  chmod 600 .env
  echo
  echo "    ************************************************************"
  echo "    ADMIN DAVET KODUNUZ: ${ADMIN_CODE}"
  echo "    Bunu güvenli bir yere kaydedin. Admin hesabı açarken gerekir."
  echo "    (Daha sonra görmek için: grep ADMIN_SECRET $APP_DIR/.env)"
  echo "    ************************************************************"
  echo
else
  echo "    .env zaten var, dokunulmadı."
fi

echo "==> 6/6 Uygulama başlatılıyor (ilk derleme birkaç dakika sürebilir)"
docker compose up -d --build

# Günlük veritabanı yedeği (her gece 03:30)
chmod +x deploy/backup.sh
cat > /etc/cron.d/maplas-backup <<EOF
30 3 * * * root $APP_DIR/deploy/backup.sh >> /var/log/maplas-backup.log 2>&1
EOF

echo
echo "==> DNS kontrolü"
SERVER_IP="$(curl -fsS https://api.ipify.org || true)"
for host in "$DOMAIN" "api.$DOMAIN" "www.$DOMAIN"; do
  resolved="$(dig +short A "$host" | tail -n1)"
  if [ -n "$SERVER_IP" ] && [ "$resolved" = "$SERVER_IP" ]; then
    echo "    OK   $host -> $resolved"
  else
    echo "    EKSİK $host -> '${resolved:-yok}' (olması gereken: ${SERVER_IP:-sunucu IP'si})"
  fi
done

echo
echo "Kurulum bitti."
echo "DNS kayıtları doğruysa birkaç dakika içinde https://$DOMAIN açılır."
echo "Durum:  cd $APP_DIR && docker compose ps"
echo "Loglar: cd $APP_DIR && docker compose logs -f backend"
