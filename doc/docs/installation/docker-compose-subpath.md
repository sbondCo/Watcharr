---
sidebar_position: 2
description: Install and setup with Docker Compose for access through a subpath.
---

# Docker Compose (subpath)

Install and setup with Docker Compose for access through a subpath.

:::info Not a great experience at the moment!

Currently, hosting Watcharr via a subpath on your server is not very easy (unlike hosting under a subdomain). Hopefully this will change in the future, but as a temporary measure to at least allow hosting under a subpath, this method has been provided.

The following issue will continue to track this: https://github.com/sbondCo/Watcharr/issues/312

:::

## Installing

### Build UI

First we have to build the frontend with our subpath provided as an environment variable.

If you don't want to use `/watcharr` as your subpath, replace the value of `WATCHARR_BASE` before running the command.

```bash
docker run -e WATCHARR_BASE=/watcharr -v watcharr-ui:/ui --rm ghcr.io/sbondco/watcharr-ui-build:latest
```

Once the container has finished running the build script, it will exit and remove itself. The `watcharr-ui` volume will contain the built files.

### Install Watcharr

Now we can install Watcharr. You can copy the example below to get started:

```yaml title="compose.yml"
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
      # Use our volume containing built ui files
      # instead of default ui included in image.
      - type: volume
        source: watcharr-ui
        target: /ui
        volume:
          nocopy: true
          subpath: build
    restart: unless-stopped

volumes:
  watcharr-ui:
    external: true
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

1. Update your built ui files by following the [Build UI](#build-ui) step again.

2. Update the `image` version in your `compose.yml` file.
   Skip this step if you are using the `latest` tag.

   ```yaml
   # eg. update v1.19.0 to v1.20.0 (or whatever version you are updating to)
   image: ghcr.io/sbondco/watcharr:v1.19.0
   ```

3. Pull the new changes and re-create your container:

   ```bash
   docker compose pull && docker compose down && docker compose up -d
   ```

And that is it!
