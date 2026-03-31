#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"

cd "$PROJECT_DIR"

# 检查并安装 Docker
install_docker() {
  echo "==> Docker not found, installing..."
  if [ -f /etc/os-release ]; then
    . /etc/os-release
    case "$ID" in
      ubuntu|debian)
        sudo apt-get update -qq
        sudo apt-get install -y -qq ca-certificates curl gnupg
        sudo install -m 0755 -d /etc/apt/keyrings
        curl -fsSL "https://download.docker.com/linux/$ID/gpg" | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
        sudo chmod a+r /etc/apt/keyrings/docker.gpg
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$ID $(lsb_release -cs) stable" | \
          sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        sudo apt-get update -qq
        sudo apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-compose-plugin
        ;;
      centos|rhel|fedora|amzn)
        sudo yum install -y yum-utils
        sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
        sudo yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
        ;;
      *)
        echo "Error: unsupported OS ($ID). Please install Docker manually:"
        echo "  https://docs.docker.com/engine/install/"
        exit 1
        ;;
    esac
  elif [ "$(uname)" = "Darwin" ]; then
    echo "Error: please install Docker Desktop for Mac:"
    echo "  https://docs.docker.com/desktop/install/mac-install/"
    exit 1
  else
    echo "Error: unsupported OS. Please install Docker manually:"
    echo "  https://docs.docker.com/engine/install/"
    exit 1
  fi

  # 启动 Docker 并设置开机自启
  sudo systemctl start docker 2>/dev/null || true
  sudo systemctl enable docker 2>/dev/null || true

  # 将当前用户加入 docker 组（免 sudo）
  if ! groups | grep -q docker; then
    sudo usermod -aG docker "$USER"
    echo "==> Added $USER to docker group. You may need to re-login for this to take effect."
  fi

  echo "==> Docker installed successfully: $(docker --version)"
}

# 检查 Docker
if ! command -v docker &>/dev/null; then
  install_docker
fi

# 检查 Docker 是否在运行
if ! docker info &>/dev/null; then
  echo "==> Starting Docker..."
  sudo systemctl start docker 2>/dev/null || true
  sleep 2
  if ! docker info &>/dev/null; then
    echo "Error: Docker is not running. Please start Docker and try again."
    exit 1
  fi
fi

# 检查 docker compose
if ! docker compose version &>/dev/null; then
  echo "Error: docker compose plugin not found. Please install:"
  echo "  sudo apt-get install docker-compose-plugin"
  exit 1
fi

# 检查 .env
if [ ! -f .env ]; then
  echo "Error: .env file not found. Copy from .env.example and fill in values:"
  echo "  cp .env.example .env"
  exit 1
fi

# 加载 .env
set -a; source .env 2>/dev/null; set +a

# 确保数据目录存在
DATA_DIR="${DATA_DIR:-./deployments/data}"
mkdir -p "$DATA_DIR/postgres" "$DATA_DIR/redis"

ACTION="${1:-help}"

case "$ACTION" in
  up)
    echo "==> Building and starting all services..."
    docker compose -f "$COMPOSE_FILE" build
    echo "==> Running database migrations..."
    docker compose -f "$COMPOSE_FILE" run --rm migrate \
      "postgres://${DB_USER:-sovereign}:${DB_PASSWORD:-sovereign_dev}@postgres:5432/${DB_NAME:-sovereign}?sslmode=disable" up
    echo "==> Starting services..."
    docker compose -f "$COMPOSE_FILE" up -d server worker front nginx
    echo "==> Done! Services:"
    docker compose -f "$COMPOSE_FILE" ps
    echo ""
    echo "Access the app at: http://localhost:${NGINX_PORT:-80}"
    ;;

  down)
    echo "==> Stopping all services..."
    docker compose -f "$COMPOSE_FILE" down
    ;;

  restart)
    echo "==> Restarting services..."
    docker compose -f "$COMPOSE_FILE" restart server worker front nginx
    ;;

  rebuild)
    echo "==> Rebuilding and restarting..."
    docker compose -f "$COMPOSE_FILE" build server worker front
    docker compose -f "$COMPOSE_FILE" up -d server worker front nginx
    ;;

  logs)
    SERVICE="${2:-}"
    if [ -n "$SERVICE" ]; then
      docker compose -f "$COMPOSE_FILE" logs -f "$SERVICE"
    else
      docker compose -f "$COMPOSE_FILE" logs -f server worker front
    fi
    ;;

  migrate)
    echo "==> Running database migrations..."
    docker compose -f "$COMPOSE_FILE" run --rm migrate \
      "postgres://${DB_USER:-sovereign}:${DB_PASSWORD:-sovereign_dev}@postgres:5432/${DB_NAME:-sovereign}?sslmode=disable" up
    ;;

  migrate-down)
    echo "==> Rolling back last migration..."
    docker compose -f "$COMPOSE_FILE" run --rm migrate \
      "postgres://${DB_USER:-sovereign}:${DB_PASSWORD:-sovereign_dev}@postgres:5432/${DB_NAME:-sovereign}?sslmode=disable" down 1
    ;;

  status)
    docker compose -f "$COMPOSE_FILE" ps
    ;;

  clean)
    echo "==> Stopping and removing all containers, volumes..."
    docker compose -f "$COMPOSE_FILE" down -v
    echo "==> Cleaned."
    ;;

  help|*)
    echo "Sovereign Deployment Script"
    echo ""
    echo "Usage: $0 <command>"
    echo ""
    echo "Commands:"
    echo "  up           Build, migrate, and start all services"
    echo "  down         Stop all services"
    echo "  restart      Restart app services (server, worker, front, nginx)"
    echo "  rebuild      Rebuild images and restart"
    echo "  logs [svc]   Tail logs (optionally for a specific service)"
    echo "  migrate      Run database migrations"
    echo "  migrate-down Rollback last migration"
    echo "  status       Show service status"
    echo "  clean        Stop and remove everything including volumes"
    ;;
esac
