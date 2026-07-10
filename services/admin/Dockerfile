# syntax=docker/dockerfile:1
FROM node:22-alpine AS deps
WORKDIR /app
ARG CACHE_BUST=1
RUN echo "cache-bust=${CACHE_BUST}" > /dev/null
COPY services/admin/package.json ./package.json
COPY services/admin/package-lock.json ./package-lock.json
RUN npm ci

FROM node:22-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY services/admin ./
COPY packages/api-contract ../../packages/api-contract
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
RUN npm run build

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
RUN addgroup -S inori && adduser -S -G inori inori
COPY --from=builder /app/public ./public
COPY --from=builder --chown=inori:inori /app/.next/standalone ./
COPY --from=builder --chown=inori:inori /app/.next/static ./.next/static
USER inori
EXPOSE 3001
ENV PORT=3001
ENV HOSTNAME=0.0.0.0
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -qO- http://127.0.0.1:3001/admin/login >/dev/null || exit 1
CMD ["node", "server.js"]
