.PHONY: dev dev-server dev-front dev-worker build test lint docker-up docker-down docker-build clean

# Development - run all (Ctrl+C stops everything)
dev:
	@trap 'kill 0' INT TERM; \
	echo "Starting server and frontend..."; \
	(cd server && go run ./cmd/server) & \
	(cd front && pnpm dev) & \
	wait

dev-server:
	cd server && go run ./cmd/server

dev-worker:
	cd server && go run ./cmd/worker

dev-front:
	cd front && pnpm dev

# Build
build:
	cd server && $(MAKE) build
	cd front && pnpm build

# Test
test:
	cd server && $(MAKE) test
	cd front && pnpm lint

# Lint
lint:
	cd server && $(MAKE) lint
	cd front && pnpm lint

# Database
migrate-up:
	cd server && $(MAKE) migrate-up

migrate-down:
	cd server && $(MAKE) migrate-down

migrate-create:
	cd server && $(MAKE) migrate-create name=$(name)

seed-premium:
	cd server && go run scripts/seed_premium.go

seed-wallets:
	cd server && go run scripts/seed_wallets.go

seed-trades:
	cd server && go run scripts/seed_trades.go

seed: seed-premium seed-wallets seed-trades

# Docker deployment
deploy:
	./deployments/deploy.sh up

deploy-down:
	./deployments/deploy.sh down

deploy-rebuild:
	./deployments/deploy.sh rebuild

deploy-logs:
	./deployments/deploy.sh logs

deploy-status:
	./deployments/deploy.sh status

deploy-migrate:
	./deployments/deploy.sh migrate

deploy-clean:
	./deployments/deploy.sh clean

# Clean
clean:
	cd server && $(MAKE) clean
	rm -rf front/.next front/node_modules
