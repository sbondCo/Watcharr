<script lang="ts">
  import Checkbox from "@/lib/Checkbox.svelte";
  import Icon from "@/lib/Icon.svelte";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { notify } from "@/lib/util/notify";
  import type { ServerConfig } from "@/types";
  import axios from "axios";
  import SonarrModal from "./modals/SonarrModal.svelte";

  let serverConfig: ServerConfig;
  let sonarrModalOpen = false;
  // Disabled vars for disabling inputs until api request completes
  let signupDisabled = false;
  let debugDisabled = false;
  let jfDisabled = false;
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
        <h3>General</h3>
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
        <h3>Services</h3>
        <div>
          <button class="plain configure" on:click={() => (sonarrModalOpen = !sonarrModalOpen)}>
            <div>
              <h4 class="norm">Sonarr</h4>
              <h5 class="norm">Configure your Sonarr server.</h5>
            </div>
            <Icon i="arrow" facing="right" />
          </button>
        </div>

        {#if sonarrModalOpen}
          <SonarrModal
            onUpdate={updateServerConfig}
            {serverConfig}
            onClose={() => (sonarrModalOpen = false)}
          />
        {/if}
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
</style>
