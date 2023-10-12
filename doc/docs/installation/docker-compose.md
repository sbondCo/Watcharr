---
sidebar_position: 1
---

# Docker Compose

## Installing

Installing Watcharr with a docker compose file is easy. You can copy the example below to get started:

```yaml title="docker-compose.yml"
version: "3"

services:
  watcharr:
    # The :latest tag is used for simplicity, it is recommended
    # to use an actual version, then when updating check the releases for changelogs.
    image: ghcr.io/sbondco/watcharr:latest
    container_name: watcharr
    ports:
      - 3080:3080
    volumes:
      # Contains all of watcharr data (database & cache)
      - ./data:/data
```

:::danger first account

When **first** running Watcharr, make sure only you have access. The first user created will become admin.

:::

You can now start `Watcharr` like so:

```bash
docker compose up -d
```

If you didn't change the ports in the example, the server will be available at [http://localhost:3080/](http://localhost:3080/).

## Updating

:::danger Take care

We try taking care as to not release breaking changes, however it is still recommended that
you lookover changelogs before updating!

Breaking changes are marked at the top of releases: https://github.com/sbondCo/Watcharr/releases

:::

Updating your server can be done in two steps:

1. Update the `image` version in your `docker-compose.yml` file.
   Skip this step if you are using the `latest` tag.

   ```yaml
   # eg. update v1.19.0 to v1.20.0 (or whatever version you are updating to)
   image: ghcr.io/sbondco/watcharr:v1.19.0
   ```

2. Pull the new changes and re-create your container:

   ```bash
   docker compose pull && docker compose down && docker compose up -d
   ```

And that is it!
