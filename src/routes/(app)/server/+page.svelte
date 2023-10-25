<script lang="ts">
  import Checkbox from "@/lib/Checkbox.svelte";
  import Icon from "@/lib/Icon.svelte";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { notify } from "@/lib/util/notify";
  import type { ServerConfig, SonarrSettings } from "@/types";
  import axios from "axios";
  import SonarrModal from "./modals/SonarrModal.svelte";
  import SettingsList from "@/lib/settings/SettingsList.svelte";
  import Setting from "@/lib/settings/Setting.svelte";
  import SettingButton from "@/lib/settings/SettingButton.svelte";

  let serverConfig: ServerConfig;
  let sonarrModalOpen = false;
  let sonarrServerEditing: SonarrSettings;
  let sonarrModalEditing = false;
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
    <SettingsList>
      <h2>Server Settings</h2>
      {#await getServerConfig()}
        <Spinner />
      {:then}
        <h3>General</h3>
        <Setting
          title="Jellyfin Host"
          desc="Point to your Jellyfin server to enable related features. Don't change server after
        already using another."
        >
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
        </Setting>
        <Setting title="TMDB Key" desc="Provide your own TMDB API Key">
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
        </Setting>
        <Setting title="Signup" desc="Allow signing up with web ui" row>
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
        </Setting>
        <Setting title="Debug" desc="Enable debug logging" row>
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
        </Setting>
        <h3>Services</h3>
        {#if serverConfig.SONARR?.length > 0}
          {#each serverConfig.SONARR as server}
            <SettingButton
              title={server.name}
              desc={`Configure server at ${server.host}`}
              onClick={() => {
                sonarrServerEditing = server;
                sonarrModalEditing = true;
                sonarrModalOpen = true;
              }}
            />
          {/each}
        {/if}
        <SettingButton
          title="Sonarr"
          desc="Add a Sonarr server."
          icon="add"
          onClick={() => {
            let name = "Sonarr";
            if (serverConfig.SONARR?.length > 0) {
              // if this still exists ya on yur own
              name = `Sonarr${serverConfig.SONARR.length + 1}`;
            }
            sonarrServerEditing = { name };
            sonarrModalEditing = false;
            sonarrModalOpen = true;
          }}
        />

        {#if sonarrModalOpen}
          <SonarrModal
            servarr={sonarrServerEditing}
            isEditing={sonarrModalEditing}
            onClose={() => (sonarrModalOpen = false)}
          />
        {/if}
      {:catch err}
        <PageError error={err} pretty="Failed to load server config" />
      {/await}
    </SettingsList>
  </div>
</div>

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;
    padding: 0 30px 30px 30px;

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
