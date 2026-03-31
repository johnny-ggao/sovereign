FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app
COPY --from=builder /bin/server .
COPY server/config/config.yaml config/
COPY server/migrations/ migrations/

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["./server"]
