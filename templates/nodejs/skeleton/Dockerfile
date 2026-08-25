# syntax=docker/dockerfile:1
FROM node:20.15.1-alpine3.20 AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci --omit=dev

FROM gcr.io/distroless/nodejs20-debian12:nonroot AS runtime
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
