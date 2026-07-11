# syntax=docker/dockerfile:1

# ── Stage 1: install dependencies ──────────────────────────────────────────
FROM node:22-alpine AS deps
WORKDIR /app
ARG CACHE_BUST=1
RUN echo "cache-bust=${CACHE_BUST}" > /dev/null
COPY services/web/package.json services/web/package-lock.json ./
RUN ls -la package.json package-lock.json && \
    node -e "const p=require('./package-lock.json'); console.log('lockfileVersion='+p.lockfileVersion, 'name='+p.name)" && \
    npm ci

# ── Stage 2: build ─────────────────────────────────────────────────────────
FROM node:22-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY services/web ./
# Include the API contract package for codegen
COPY packages/api-contract ../../packages/api-contract
# @inori/ui is a file: dependency (package-lock symlinks into ../../packages/ui)
# and next.config.ts transpiles it — must be physically present at build time.
COPY packages/ui ../../packages/ui
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
RUN npm run build

# ── Stage 3: runtime ───────────────────────────────────────────────────────
FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup -S inori && adduser -S -G inori inori

COPY --from=builder /app/public ./public
COPY --from=builder --chown=inori:inori /app/.next/standalone ./
COPY --from=builder --chown=inori:inori /app/.next/static ./.next/static

USER inori
EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -qO- http://127.0.0.1:3000/ >/dev/null || exit 1

CMD ["node", "server.js"]
