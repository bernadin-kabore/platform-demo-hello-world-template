# syntax=docker/dockerfile:1
# Node 20 reached end of life in April 2026: no more security releases, so its
# CVEs are permanent and the image scan would refuse the push. The runtime base
# determines what Trivy reports against the published image -- the deps stage
# only contributes node_modules -- so both stages move together, and the
# workflows' node-version moves with them so CI tests on what actually ships.
#
# debian13, not debian12. The debian12 distroless nonroot tag ships
# libssl3 3.0.18-1~deb12u2, which carries a CRITICAL (CVE-2026-31789) and
# several HIGH advisories fixed in 3.0.19/3.0.20 -- fixes that exist but that
# the published tag had not been rebuilt with. Unlike the Go service, which
# runs on distroless/static and therefore carries no OpenSSL at all, a Node
# runtime needs it, so the base image's rebuild cadence is this service's
# vulnerability exposure.
FROM node:22.23.2-alpine3.24 AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci --omit=dev

FROM gcr.io/distroless/nodejs22-debian13:nonroot AS runtime
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY package.json ./
COPY src ./src

# distroless nonroot images already run as uid 65532; no USER directive
# needed, and there's no shell to drop into even if a Kyverno policy or the
# PSS restricted profile were somehow bypassed.
EXPOSE 8080
ENV NODE_ENV=production
ENTRYPOINT ["/nodejs/bin/node", "--require", "/app/src/tracing.js", "/app/src/index.js"]
