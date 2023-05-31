# Backend
FROM golang:1.20 AS server

WORKDIR /server
COPY server/*.go server/go.* ./

RUN go mod download && GOOS=linux go build -o ./watcharr

# Frontend
FROM node:19 AS ui

WORKDIR /app
COPY package*.json vite.config.ts svelte.config.js tsconfig.json ./
COPY ./src ./src
COPY ./static ./static

RUN npm install && npm run build

# Production
FROM debian:11.6 AS runner

RUN apt-get update && apt-get install ca-certificates -y

COPY --from=server /server/watcharr /
COPY --from=ui /app/build /ui

EXPOSE 3080

CMD ["/watcharr"]
