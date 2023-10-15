<script lang="ts">
  import Checkbox from "@/lib/Checkbox.svelte";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { notify } from "@/lib/util/notify";
  import type { ServerConfig } from "@/types";
  import axios from "axios";

  let serverConfig: ServerConfig;
  // Disabled vars for disabling inputs until api request completes
  let signupDisabled = false;
  let debugDisabled = false;
  let jfDisabled = false;
  let soDisabled = false;
  let soaDisabled = false;
  let raDisabled = false;
  let raaDisabled = false;
  let tmdbkDisabled = false;

  async function getServerConfig() {
    serverConfig = (await axios.get(`/server/config`)).data as ServerConfig;
  }

  export function updateServerConfig<K extends keyof ServerConfig>(
    name: K,
    value: ServerConfig[K],
    done?: () => void
  ) {
    console.log("Updating server setting", name, "to", value);
    const originalValue = serverConfig[name];
    const nid = notify({ type: "loading", text: "Updating" });
    axios
      .post("/server/config", { key: name, value: value })
      .then((r) => {
        if (r.status === 200) {
          serverConfig[name] = value;
          serverConfig = serverConfig;
          notify({ id: nid, type: "success", text: "Updated" });
          if (typeof done !== "undefined") done();
        }
      })
      .catch((err) => {
        console.error("Failed to update user setting", err);
        notify({ id: nid, type: "error", text: "Couldn't Update" });
        serverConfig[name] = originalValue;
        serverConfig = serverConfig;
        if (typeof done !== "undefined") done();
      });
  }
</script>

<div class="content">
  <div class="inner">
    <div class="settings">
      <h2>Server Settings</h2>
      {#await getServerConfig()}
        <Spinner />
      {:then}
        <div>
          <h4 class="norm">Jellyfin Host</h4>
          <h5 class="norm">
            Point to your Jellyfin server to enable related features. Don't change server after
            already using another.
          </h5>
          <input
            type="text"
            placeholder="https://jellyfin.example.com"
            bind:value={serverConfig.JELLYFIN_HOST}
            on:blur={() => {
              jfDisabled = true;
              updateServerConfig("JELLYFIN_HOST", serverConfig.JELLYFIN_HOST, () => {
                jfDisabled = false;
              });
            }}
            disabled={jfDisabled}
          />
        </div>
        <div>
          <h4 class="norm">Sonarr Host</h4>
          <h5 class="norm">Point to your Sonarr server to enable tv show requesting.</h5>
          <input
            type="text"
            placeholder="https://sonarr.example.com"
            bind:value={serverConfig.SONARR_HOST}
            on:blur={() => {
              soDisabled = true;
              updateServerConfig("SONARR_HOST", serverConfig.SONARR_HOST, () => {
                soDisabled = false;
              });
            }}
            disabled={soDisabled}
          />
        </div>
        {#if serverConfig.SONARR_HOST}
          <div>
            <h4 class="norm">Sonarr Key</h4>
            <h5 class="norm">API key for your Sonarr instance.</h5>
            <input
              type="text"
              placeholder="dGhhbmtzIGZvciB1c2luZyB3YXRjaGFyciA6KQ=="
              bind:value={serverConfig.SONARR_KEY}
              on:blur={() => {
                soaDisabled = true;
                updateServerConfig("SONARR_KEY", serverConfig.SONARR_KEY, () => {
                  soaDisabled = false;
                });
              }}
              disabled={soaDisabled}
            />
          </div>
        {/if}
        <div>
          <h4 class="norm">Radarr Host</h4>
          <h5 class="norm">Point to your Radarr server to enable movie requesting.</h5>
          <input
            type="text"
            placeholder="https://radarr.example.com"
            bind:value={serverConfig.RADARR_HOST}
            on:blur={() => {
              raDisabled = true;
              updateServerConfig("RADARR_HOST", serverConfig.RADARR_HOST, () => {
                raDisabled = false;
              });
            }}
            disabled={raDisabled}
          />
        </div>
        {#if serverConfig.RADARR_HOST}
          <div>
            <h4 class="norm">Radarr Key</h4>
            <h5 class="norm">API key for your Radarr instance.</h5>
            <input
              type="text"
              placeholder="bm8gcmVhbGx5IGkgYXBwcmVjaWF0ZSB5b3UgOik="
              bind:value={serverConfig.RADARR_KEY}
              on:blur={() => {
                raaDisabled = true;
                updateServerConfig("RADARR_KEY", serverConfig.RADARR_KEY, () => {
                  raaDisabled = false;
                });
              }}
              disabled={raaDisabled}
            />
          </div>
        {/if}
        <div>
          <h4 class="norm">TMDB Key</h4>
          <h5 class="norm">Provide your own TMDB API Key</h5>
          <input
            type="password"
            placeholder="TMDB Key"
            bind:value={serverConfig.TMDB_KEY}
            on:blur={() => {
              tmdbkDisabled = true;
              updateServerConfig("TMDB_KEY", serverConfig.TMDB_KEY, () => {
                tmdbkDisabled = false;
              });
            }}
            disabled={tmdbkDisabled}
          />
        </div>
        <div class="row">
          <div>
            <h4 class="norm">Signup</h4>
            <h5 class="norm">Allow signing up with web ui</h5>
          </div>
          <Checkbox
            name="SIGNUP_ENABLED"
            disabled={signupDisabled}
            value={serverConfig.SIGNUP_ENABLED}
            toggled={(on) => {
              signupDisabled = true;
              updateServerConfig("SIGNUP_ENABLED", on, () => {
                signupDisabled = false;
              });
            }}
          />
        </div>
        <div class="row">
          <div>
            <h4 class="norm">Debug</h4>
            <h5 class="norm">Enable debug logging</h5>
          </div>
          <Checkbox
            name="DEBUG"
            disabled={debugDisabled}
            value={serverConfig.DEBUG}
            toggled={(on) => {
              debugDisabled = true;
              updateServerConfig("DEBUG", on, () => {
                debugDisabled = false;
              });
            }}
          />
        </div>
      {:catch err}
        <PageError error={err} pretty="Failed to load server config" />
      {/await}
    </div>
  </div>
</div>

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;
    padding: 0 30px 0 30px;

    .inner {
      min-width: 400px;
      max-width: 400px;
      overflow: hidden;

      h2 {
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
      }

      & > div:not(:first-of-type) {
        margin-top: 30px;
      }

      @media screen and (max-width: 440px) {
        width: 100%;
        min-width: unset;
      }
    }
  }

  .settings {
    display: flex;
    flex-flow: column;
    gap: 20px;
    width: 100%;

    h3 {
      font-variant: small-caps;
    }

    h5 {
      font-weight: normal;
    }

    & > div {
      margin: 0 15px;
    }

    div {
      input[type="text"],
      input[type="password"] {
        margin-top: 1px;
      }

      &.row {
        display: flex;
        flex-flow: row;
        gap: 10px;
        align-items: center;

        & > div:first-of-type {
          margin-right: auto;
        }

        &.btns button {
          width: min-content;
        }
      }
    }
  }
</style>
