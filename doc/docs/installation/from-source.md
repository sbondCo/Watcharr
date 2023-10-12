---
sidebar_position: 10
---

# From Source

## Installing

1. Get the code

   ```bash
   git clone https://github.com/sbondCo/Watcharr.git && \
   cd Watcharr
   ```

   If you want a specific version, checkout a version tag (replace `<tag>` with the tag):

   ```bash
   git checkout <tag>
   ```

2. Build the frontend

   The server will use the built files, so we move them to the `ui` folder, besides where it's binary will be built.

   ```bash
   npm i && \
   npm run build && \
   mv ./build ./server/ui
   ```

3. Build the server

   ```bash
   cd server && \
   go mod download && \
   GOOS=linux go build -o ./watcharr
   ```

4. Run the server

   :::danger first account

   When **first** running Watcharr, make sure only you have access. The first user created will become admin.

   :::

   ```bash
   ./watcharr
   ```

5. Visit [http://localhost:3080/](http://localhost:3080/) and setup Watcharr.

## Updating

:::danger Take care

We try taking care as to not release breaking changes, however it is still recommended that
you lookover changelogs before updating!

Breaking changes are marked at the top of releases: https://github.com/sbondCo/Watcharr/releases

:::

Updating is the same as installing except:

1. Cleanup old build. Delete the `ui` folder and the `watcharr` binary.
2. Replace step one with getting the new code:

   ```bash
   git pull
   ```

   If you checked out a tag, update to the new tag as well.
