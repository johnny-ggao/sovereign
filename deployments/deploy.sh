#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"

cd "$PROJECT_DIR"

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
