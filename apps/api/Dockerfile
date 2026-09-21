# Multi-stage Go build
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Dependencies first, so this layer is cached until go.mod/go.sum change.
# go.sum is committed: the build fails loudly if it is out of sync instead of silently re-resolving versions.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /build/cards-api ./cmd/server

# Minimal runtime image
FROM alpine:3.20

# sqlite: used by scripts/backup-sqlite.sh (online .backup). su-exec: drop root after fixing volume ownership.
RUN apk add --no-cache ca-certificates tzdata sqlite su-exec \
    && addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=builder /build/cards-api /app/cards-api
COPY migrations /app/migrations
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

RUN chmod +x /usr/local/bin/docker-entrypoint.sh && mkdir -p /data && chown app:app /data

ENV PORT=8080 \
    DATABASE_PATH=/data/cards.db \
    MIGRATIONS_PATH=/app/migrations \
    APP_ENV=production

EXPOSE 8080

VOLUME ["/data"]

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["/app/cards-api"]
