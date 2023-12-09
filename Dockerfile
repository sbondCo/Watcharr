# Backend
FROM golang:1.21-alpine AS server

WORKDIR /server
COPY server/*.go server/go.* ./
COPY server/arr/*.go ./arr/

# Required so we can build with cgo
RUN apk update && apk add --no-cache musl-dev gcc build-base

# CGO_CFLAGS: https://github.com/mattn/go-sqlite3/issues/1164#issuecomment-1635253695
RUN go mod download && GOOS=linux CGO_ENABLED=1 CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -o ./watcharr

# Frontend
FROM node:20-alpine AS ui

WORKDIR /app
COPY package*.json vite.config.ts svelte.config.js tsconfig.json ./
COPY ./src ./src
COPY ./static ./static

RUN npm install && npm run build

# Production
FROM node:20-alpine AS runner

COPY --from=server /server/watcharr /
COPY --from=ui /app/build /ui
COPY --from=ui /app/package.json /app/package-lock.json /ui
RUN cd /ui && npm ci --omit=dev

EXPOSE 3080

CMD ["/watcharr"]
