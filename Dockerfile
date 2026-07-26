#
# Backend
#
FROM golang:1.25-alpine AS server

WORKDIR /server

# We can copy the whole server folder, since our .dockerignore file
# will filter out stuff we don't want in the final image.
COPY server/ ./

# Required so we can build with cgo
RUN apk update && apk add --no-cache musl-dev gcc build-base

# CGO_CFLAGS: https://github.com/mattn/go-sqlite3/issues/1164#issuecomment-1635253695
RUN go mod download && GOOS=linux CGO_ENABLED=1 CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -o ./watcharr

#
# Frontend
#
FROM node:24-alpine AS ui

WORKDIR /app
COPY package*.json vite.config.ts svelte.config.js tsconfig.json ./
COPY ./src ./src
COPY ./static ./static

RUN npm install && npm run build

#
# Production
#
FROM node:24-alpine AS runner

COPY --from=server /server/watcharr /
COPY --from=ui /app/build /ui
COPY --from=ui /app/package.json /app/package-lock.json /ui/

# Install just the prod dependencies for final step.
# We --ignore-scripts, to stop the `prepare` script from auto-
# running, which is only meant for development (will error).
RUN cd /ui && npm ci --omit=dev --ignore-scripts=true

EXPOSE 3080

CMD ["/watcharr"]
