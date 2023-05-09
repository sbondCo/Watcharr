<center>
<h1>Watcharr</h1>
<img src="./static/logo-col.png" alt="logo" />
</center>

I'm the place you store your watched content for later rememberance, sadly I am still in my infancy. I aspire to be easily self hosted.

I am built with Go and Svelte(Kit).

Feel free to abuse this demo instance (nicely), which runs on the latest `dev` build: [https://watcharr.lab.sbond.co/](https://watcharr.lab.sbond.co/)

# Set Up

Currently we have docker images for each version that are fairly simple to setup, but it is a lot more complex that I would like it to be. Hopefully this improves in the future.

For now though, we have two container images ([UI](https://github.com/sbondCo/Watcharr/pkgs/container/watcharr-ui) & [Server](https://github.com/sbondCo/Watcharr/pkgs/container/watcharr-ui)) that work together through the use of a proxy. We have an example of this with Caddy in the repo.

Here's a simple setup to get you started:

**docker-compose.yml**

```
version: "3"

services:
  # Watcharr API
  watcharr:
    image: ghcr.io/sbondco/watcharr:latest
    container_name: watcharr
    ports:
      - 3080:3080
    volumes:
      # .env file to configure watcharr
      - ./.env:/.env
      # Contains all of watcharr data (database & cache)
      - ./data:/data

  # Watcharr Frontend
  watcharr-ui:
    image: ghcr.io/sbondco/watcharr-ui:latest
    container_name: watcharr-ui
    ports:
      - 3000:3000
    depends_on:
      - watcharr

  # Caddy - Used as reverse proxy to services
  # and to re-route requests
  caddy:
    image: caddy:2.6
    container_name: watcharr-caddy
    ports:
      - "8080:80"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
```

**Caddyfile**

```
:80 {
  # Reverse proxy for frontend
  reverse_proxy watcharr:3000

  # Route requests to /api/ to backend
  handle_path /api/* {
    reverse_proxy watcharr-server:3080
  }
}
```

**.env**

```
# Used to sign JWT tokens. Make sure to make
# it strong, just like a very long, complicated password.
JWT_SECRET=MAKE_ME_RANDOM_AND_LONG

# Optional: Point to your Jellyfin install
# to enable it as an auth provider.
JELLYFIN_HOST=https://my.jellyfin.example
```
