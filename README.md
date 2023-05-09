<h1 align="center">Watcharr</h1>
<p align="center"><img src="./static/logo-col.png" alt="logo" /></p>


I'm the place you store your watched content for later rememberance, sadly I am still in my infancy. I aspire to be easily self hosted.

I am built with Go and Svelte(Kit).

Feel free to abuse this demo instance (nicely), which runs on the latest `dev` build: [https://watcharr.lab.sbond.co/](https://watcharr.lab.sbond.co/)

# Screenshots

<h3 align="center">Watched List</h3>
<p align="center">
<img src="./screenshot/homepage.png" alt="Watched List" />

| Watched Show Hover                                                            | Watched Show Status Change                                                                           | Show Details                                                               |
| ----------------------------------------------------------------------- | --------------------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| <img src="./screenshot/homepage-poster-hover.png" alt="Watched List" /> | <img src="./screenshot/homepage-poster-change-status.png" alt="Changing Show Status" /> | <img src="./screenshot/show-details-page.png" alt="Content Details Page" /> |
</p>

# Set Up

Currently we have docker images for each version that are fairly simple to setup, but it is a lot more complex that I would like it to be. Hopefully this improves in the future.

For now though, we have two container images ([UI](https://github.com/sbondCo/Watcharr/pkgs/container/watcharr-ui) & [Server](https://github.com/sbondCo/Watcharr/pkgs/container/watcharr-ui)) that work together through the use of a proxy.

Here's a simple setup to get you started based from visible files in the repo, which you can also check out. We use Caddy here, but use whatever you are comfortable with!

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
