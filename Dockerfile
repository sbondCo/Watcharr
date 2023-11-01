# Backend
FROM golang:1.21 AS server

WORKDIR /server
COPY server/*.go server/go.* ./

RUN go mod download && GOOS=linux go build -o ./watcharr

# Frontend
FROM node:20 AS ui

WORKDIR /app
COPY package*.json vite.config.ts svelte.config.js tsconfig.json ./
COPY ./src ./src
COPY ./static ./static

RUN npm install && npm run build

# Production
FROM node:20 AS runner

COPY --from=server /server/watcharr /
COPY --from=ui /app/build /ui
COPY --from=ui /app/package.json /app/package-lock.json /ui
RUN cd /ui && npm ci

EXPOSE 3080

CMD ["/watcharr"]
