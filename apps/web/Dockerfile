# Build stage: compile SvelteKit to static files
FROM node:24-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm install

COPY . .
RUN npm run build

# Static verification stage (outputs are in /app/build)
FROM alpine:3.20 AS final
WORKDIR /dist
COPY --from=builder /app/build /dist

CMD ["echo", "Web static build complete"]
