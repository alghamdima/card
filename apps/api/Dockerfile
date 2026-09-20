# Multi-stage Go build
FROM golang:1.23-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/cards-api ./cmd/server/main.go

# Minimal runtime image
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata sqlite

WORKDIR /app

COPY --from=builder /build/cards-api /app/cards-api
COPY migrations /app/migrations

RUN mkdir -p /data

ENV PORT=8080 \
    DATABASE_PATH=/data/cards.db \
    MIGRATIONS_PATH=/app/migrations \
    APP_ENV=production

EXPOSE 8080

VOLUME ["/data"]

CMD ["/app/cards-api"]
