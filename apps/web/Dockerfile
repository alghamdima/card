# Build context is the repository root (see infra/docker-compose.yml), because the
# runtime image also needs the nginx configuration from infra/nginx.

# --- Build stage: compile SvelteKit to static files ---
FROM node:24-alpine AS builder

WORKDIR /app

# Reproducible install from the lockfile (npm install could drift).
COPY apps/web/package.json apps/web/package-lock.json ./
RUN npm ci

COPY apps/web/ .
RUN npm run build

# --- Runtime stage: nginx serves the SPA and proxies /api to the Go service ---
FROM nginx:1.27-alpine

COPY infra/nginx/nginx.conf /etc/nginx/nginx.conf
COPY infra/nginx/proxy-api.conf /etc/nginx/proxy-api.conf
COPY infra/nginx/security-headers.conf /etc/nginx/security-headers.conf
COPY --from=builder /app/build /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=15s --timeout=5s --retries=3 CMD wget -q --spider http://127.0.0.1/ || exit 1
