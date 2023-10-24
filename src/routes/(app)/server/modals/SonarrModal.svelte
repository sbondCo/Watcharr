<script lang="ts">
  import Checkbox from "@/lib/Checkbox.svelte";
  import DropDown from "@/lib/DropDown.svelte";
  import Modal from "@/lib/Modal.svelte";
  import Setting from "@/lib/settings/Setting.svelte";
  import SettingsList from "@/lib/settings/SettingsList.svelte";
  import type { DropDownItem, ServerConfig, SonarrSettings, SonarrTestResponse } from "@/types";
  import axios from "axios";

  export let servarr: SonarrSettings;
  export let onUpdate: <K extends keyof ServerConfig>(
    name: K,
    value: ServerConfig[K],
    done?: () => void
  ) => void;
  export let onClose: () => void;

  let error: string;
  let formDisabled = false;

  let qualityProfiles: DropDownItem[] = [];
  let languageProfiles: DropDownItem[] = [];
  let rootFolders: DropDownItem[] = [];

  function testIfHostAndKeySet() {
    if (servarr.host && servarr.key) {
      console.log("host and key set.. loading settings data from sonarr");
      getSettingsData();
    } else {
      qualityProfiles = [];
    }
  }

  async function getSettingsData() {
    try {
      formDisabled = true;
      const res = await axios.post<SonarrTestResponse>("/arr/son/test", {
        host: servarr.host,
        key: servarr.key
      });
      qualityProfiles = res.data.qualityProfiles.map((d) => {
        return { id: d.id, value: d.name };
      });
      languageProfiles = res.data.languageProfiles.map((d) => {
        return { id: d.id, value: d.name };
      });
      rootFolders = res.data.rootFolders.map((d) => {
        return { id: d.id, value: d.path };
      });
      formDisabled = false;
      error = "";
    } catch (err) {
      console.error("getSettingsData failed!", err);
      error = "Request Failed!";
      formDisabled = false;
    }
  }

  async function save() {}
</script>

<Modal title="Sonarr" desc="Setup a connection to your Sonarr server" {onClose}>
  {#if error}
    <span class="error">{error}!</span>
  {/if}
  <SettingsList>
    <Setting title="Name" desc="Give your server a memorable name.">
      <input
        type="text"
        placeholder="https://sonarr.example.com"
        bind:value={servarr.name}
        disabled={formDisabled}
      />
    </Setting>
    <Setting title="Sonarr Host" desc="Point to your Sonarr server to enable tv show requesting.">
      <input
        type="text"
        placeholder="https://sonarr.example.com"
        bind:value={servarr.host}
        disabled={formDisabled}
      />
    </Setting>
    <Setting title="Sonarr Key" desc="API key for your Sonarr instance.">
      <input
        type="text"
        placeholder="dGhhbmtzIGZvciB1c2luZyB3YXRjaGFyciA6KQ=="
        bind:value={servarr.key}
        disabled={formDisabled}
      />
    </Setting>
    <Setting title="Quality Profile" desc="Default quality profile when adding shows.">
      <DropDown
        placeholder={qualityProfiles?.length == 0
          ? "Add a Host and Key, then press test to view"
          : "Select a quality profile"}
        options={qualityProfiles}
        bind:active={servarr.qualityProfile}
        disabled={formDisabled}
      />
    </Setting>
    <Setting title="Root Folder" desc="Default root folder when adding shows.">
      <DropDown
        placeholder={rootFolders?.length == 0
          ? "Add a Host and Key, then press test to view"
          : "Select a root folder"}
        options={rootFolders}
        bind:active={servarr.rootFolder}
        disabled={formDisabled}
      />
    </Setting>
    <Setting title="Language Profile" desc="Default language profile when adding shows.">
      <DropDown
        placeholder={languageProfiles?.length == 0
          ? "Add a Host and Key, then press test to view"
          : "Select a language profile"}
        options={languageProfiles}
        bind:active={servarr.languageProfile}
        disabled={formDisabled}
      />
    </Setting>
    <Setting title="Automatic Search" desc="Start missing episode search automatically?" row>
      <!-- <Checkbox
        name="searchForMissingEps"
        disabled={formDisabled}
        value={settings?.hideSpoilers}
      /> -->
    </Setting>
    <div class="btns">
      <button class="secondary" on:click={() => testIfHostAndKeySet()}>Test</button>
      <button on:click={() => save()}>Save</button>
    </div>
  </SettingsList>
</Modal>

<style lang="scss">
  .btns {
    display: flex;
    flex-flow: row;
    gap: 10px;
    margin-left: auto;

    button {
      width: max-content;
      padding-left: 15px;
      padding-right: 15px;
    }
  }

  .error {
    display: flex;
    justify-content: center;
    width: 100%;
    padding: 10px;
    background-color: rgb(221, 48, 48);
    text-transform: capitalize;
    color: white;
  }
</style>
