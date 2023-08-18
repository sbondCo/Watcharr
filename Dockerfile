# Backend
FROM golang:1.21 AS server

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
FROM debian:12.0 AS runner

RUN apt-get update && apt-get install ca-certificates -y

ENV NODE_VERSION=18.13.0
RUN apt install -y curl
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.3/install.sh | bash
ENV NVM_DIR=/root/.nvm
RUN . "$NVM_DIR/nvm.sh" && nvm install ${NODE_VERSION}
RUN . "$NVM_DIR/nvm.sh" && nvm use v${NODE_VERSION}
RUN . "$NVM_DIR/nvm.sh" && nvm alias default v${NODE_VERSION}
ENV PATH="/root/.nvm/versions/node/v${NODE_VERSION}/bin/:${PATH}"
RUN node --version

COPY --from=server /server/watcharr /
COPY --from=ui /app/build /ui
COPY --from=ui /app/package.json /app/package-lock.json /ui
RUN cd /ui && npm ci

EXPOSE 3080

CMD ["/watcharr"]
