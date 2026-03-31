.PHONY: build run dev dev-worker test test-coverage test-integration lint migrate-up migrate-down migrate-create clean

# Build
build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server
	CGO_ENABLED=0 go build -o bin/worker ./cmd/worker

run: build
	./bin/server

dev:
	go run ./cmd/server

dev-worker:
	go run ./cmd/worker

# Test
test:
	go test ./... -short -count=1

test-coverage:
	go test ./... -short -coverprofile=coverage.out
	go tool cover -func=coverage.out

test-integration:
	go test ./... -tags=integration -count=1

# Lint
lint:
	golangci-lint run ./...

# Database
migrate-up:
	migrate -path migrations -database "$${DATABASE_URL}" up

migrate-down:
	migrate -path migrations -database "$${DATABASE_URL}" down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# Clean
clean:
	rm -rf bin/ coverage.out
