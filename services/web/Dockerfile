# syntax=docker/dockerfile:1

# ── Stage 1: install dependencies ──────────────────────────────────────────
FROM node:22-alpine AS deps
# WORKDIR must mirror the real monorepo depth (services/web is 2 levels below
# the repo root), NOT a flat /app. @inori/ui is a file: link dependency
# (file:../../packages/ui), recorded in package-lock.json under the literal
# key "../../packages/ui". npm 10.9.8's Arborist resolves that key at runtime
# via path.resolve(WORKDIR, "../../packages/ui") + path.relative(...) and
# compares the result against the lockfile's literal string (#resolveLinks in
# @npmcli/arborist/lib/arborist/load-virtual.js). With a shallow WORKDIR like
# /app (1 level deep), path.resolve clamps at filesystem root, so the
# recomputed key comes out as "../packages/ui" — one level short — the lookup
# misses, and Arborist crashes with "Cannot read properties of undefined
# (reading 'extraneous')" in #loadLink, immediately, before touching disk.
# Merely copying packages/ui into the stage (tried previously) does not fix
# this: it's a path-string mismatch in Arborist's in-memory link resolution,
# not a missing-file problem. Verified locally: path.resolve('/app',
# '../../packages/ui') -> '/packages/ui' (wrong), but path.resolve(
# '/repo/services/web', '../../packages/ui') -> '/repo/packages/ui' (correct,
# matches the lockfile key when made relative again).
WORKDIR /repo/services/web
ARG CACHE_BUST=1
RUN echo "cache-bust=${CACHE_BUST}" > /dev/null
COPY services/web/package.json services/web/package-lock.json ./
COPY packages/ui ../../packages/ui
RUN ls -la package.json package-lock.json && \
    node -e "const p=require('./package-lock.json'); console.log('lockfileVersion='+p.lockfileVersion, 'name='+p.name)" && \
    npm ci

# ── Stage 2: build ─────────────────────────────────────────────────────────
FROM node:22-alpine AS builder
# Same WORKDIR depth as the deps stage: npm ci above created a *relative*
# symlink at node_modules/@inori/ui (pointing back up into packages/ui). That
# symlink's target string travels verbatim with the COPY --from=deps below,
# so this stage must sit at the same depth for it to keep resolving.
WORKDIR /repo/services/web
COPY --from=deps /repo/services/web/node_modules ./node_modules
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
# Independent of the deps/builder path depth above: Next.js's output:
# "standalone" tracing root is inferred from the nearest package-lock.json
# walking *up* from the app dir (see next/dist/lib/find-root.js), and
# packages/ui, packages/api-contract are siblings, not ancestors, of
# services/web — so they're never considered, and the app's own lockfile at
# services/web/package-lock.json makes it its own tracing root regardless of
# the absolute path depth chosen above. Flat /app here is fine and unchanged.
FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup -S inori && adduser -S -G inori inori

COPY --from=builder /repo/services/web/public ./public
COPY --from=builder --chown=inori:inori /repo/services/web/.next/standalone ./
COPY --from=builder --chown=inori:inori /repo/services/web/.next/static ./.next/static

USER inori
EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -qO- http://127.0.0.1:3000/ >/dev/null || exit 1

CMD ["node", "server.js"]
