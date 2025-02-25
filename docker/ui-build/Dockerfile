#
# Frontend
#
FROM node:20-alpine AS ui

COPY ./docker/ui-build/entrypoint.sh /entrypoint.sh
RUN ["chmod", "+x", "/entrypoint.sh"]

WORKDIR /app
COPY package*.json vite.config.ts svelte.config.js tsconfig.json ./
COPY ./src ./src
COPY ./static ./static

VOLUME /ui

ENTRYPOINT ["/entrypoint.sh"]
