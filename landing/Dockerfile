# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

ENV NEXT_TELEMETRY_DISABLED=1

COPY package*.json ./

RUN --mount=type=cache,target=/root/.npm npm install

COPY . .

RUN --mount=type=cache,target=/app/.next/cache \
    npm run build

# Production stage
FROM node:20-alpine AS production

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000
ENV NEXT_TELEMETRY_DISABLED=1

COPY --from=builder --chown=node:node /app/.next/standalone ./
COPY --from=builder --chown=node:node /app/.next/static ./.next/static
COPY --from=builder --chown=node:node /app/public ./public

EXPOSE 3000

USER node

CMD ["node", "server.js"]
