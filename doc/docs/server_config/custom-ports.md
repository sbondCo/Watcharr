---
sidebar_position: 12
description: How to change the ports Watcharr runs on.
---

# Port Configuration

If you're running outside of docker, you may run in to port conflicts with the defaults in Watcharr. You can override these with settings in the `watcharr.json`.

## Watcharr Service Itself

The main part of Watcharr.

Partial Example:

```json
{
  "API_PORT": 4090,
  "API_HOST": "127.0.0.1",
  ... // The rest of the json
}
```

### API_PORT

Default: `3080`
Description: Configures the port the Watcharr service listens to.
Location: This value is at the root of the `watcharr.json`.

### API_HOST

Default: `0.0.0.0`
Description: What interface IP to bind to, if running behind a proxy `127.0.0.1` is probably more appropriate.
Location: This value is at the root of the `watcharr.json`.

## The Node.JS Web Asset Server

This is what serves js/html/image files.

### WEB_ASSET_SERVER

This is an object at the root of the `watcharr.json` it contains the following options.

Partial Example:

```json
{
  "WEB_ASSET_SERVER": {
    "port": 3333,
    "host": "10.88.0.1"
  }
  ... // The rest of the json
}
```

#### port

Default: `3000`
Description: The port the web asset server listens on. The main watcharr executable is the only thing that talks to this directly.
Location: Within the `WEB_ASSET_SERVER` object.

#### host

Default: `127.0.0.1`
Description: What interface IP to bind to.
Location: Within the `WEB_ASSET_SERVER` object.
