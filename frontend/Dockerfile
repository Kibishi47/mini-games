# Stage 1: Build
FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm install

COPY . .

RUN npm run build

# Stage 2: Runtime Nitro Server
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000
ENV HOST=0.0.0.0

RUN apk --no-cache add wget \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/.output /app/.output

USER appuser

EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
