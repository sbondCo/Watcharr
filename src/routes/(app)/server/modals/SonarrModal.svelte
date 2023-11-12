<script lang="ts">
  import Checkbox from "@/lib/Checkbox.svelte";
  import DropDown from "@/lib/DropDown.svelte";
  import Modal from "@/lib/Modal.svelte";
  import Setting from "@/lib/settings/Setting.svelte";
  import SettingsList from "@/lib/settings/SettingsList.svelte";
  import { notify } from "@/lib/util/notify";
  import type { DropDownItem, SonarrSettings, SonarrTestResponse } from "@/types";
  import axios from "axios";

  export let servarr: SonarrSettings;
  export let isEditing: boolean;
  export let onClose: () => void;

  let error: string;
  let formDisabled = false;

  let qualityProfiles: DropDownItem[] = [];
  let languageProfiles: DropDownItem[] = [];
  let rootFolders: DropDownItem[] = [];

  $: {
    if (isEditing) {
      getSettingsData();
    }
  }

  function testIfHostAndKeySet() {
    if (servarr.host && servarr.key) {
      console.log("host and key set.. loading settings data from sonarr");
      getSettingsData();
    } else {
      qualityProfiles = [];
      languageProfiles = [];
      rootFolders = [];
      error = "host and key must be provided to test the connection";
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
      error = `Request Failed, check your Host and Key`;
      formDisabled = false;
    }
  }

  function checkForm() {
    let errs: string[] = [];
    if (!servarr.name) {
      errs.push("name");
    }
    if (!servarr.host) {
      errs.push("host");
    }
    if (!servarr.key) {
      errs.push("key");
    }
    if (!servarr.qualityProfile) {
      errs.push("qualityProfile");
    }
    if (!servarr.rootFolder) {
      errs.push("rootFolder");
    }
    if (!servarr.languageProfile) {
      errs.push("languageProfile");
    }
    if (errs.length > 0) {
      error = `Missing required params: ${errs.join(", ")}`;
    } else {
      error = "";
    }
  }

  async function save() {
    checkForm();
    // if no error set from checkForm func, continue
    if (!error) {
      console.log(servarr);
      try {
        const res = await axios.post(`/arr/son/${isEditing ? "edit" : "add"}`, servarr);
        if (res.status === 200) {
          notify({
            type: "success",
            text: isEditing ? "Changes saved!" : "Server added successfully!"
          });
          onClose();
        }
      } catch (err: any) {
        console.error("Failed to save server!", err);
        error = `Failed to ${isEditing ? "edit" : "add"}`;
        if (err?.response?.data?.error) {
          error = err.response.data.error;
        }
      }
    }
  }

  async function remove() {
    try {
      const res = await axios.post(`/arr/son/rm/${servarr.name}`);
      if (res.status === 200) {
        notify({
          type: "success",
          text: "Removed server"
        });
        onClose();
      }
    } catch (err: any) {
      console.error("Failed to remove server!", err);
      error = `Failed to remove`;
      if (err?.response?.data?.error) {
        error = err.response.data.error;
      }
    }
  }
</script>

<Modal
  title={isEditing ? `Edit ${servarr.name}` : "Add Sonarr Server"}
  desc="Setup a connection to your Sonarr server"
  {onClose}
>
  {#if error}
    <span class="error">{error}!</span>
  {/if}
  <SettingsList>
    {#if !isEditing}
      <Setting title="Name" desc="Give your server a memorable name.">
        <input
          type="text"
          placeholder="https://sonarr.example.com"
          bind:value={servarr.name}
          disabled={formDisabled}
        />
      </Setting>
    {/if}
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
      <Checkbox
        name="automaticSearch"
        disabled={formDisabled}
        bind:value={servarr.automaticSearch}
      />
    </Setting>
    <div class="btns">
      {#if isEditing}
        <button class="danger" on:click={() => remove()}>Delete</button>
      {/if}
      <button id="test" class="secondary" on:click={() => testIfHostAndKeySet()}>Test</button>
      <button on:click={() => save()}>{isEditing ? "Save" : "Add Server"}</button>
    </div>
  </SettingsList>
</Modal>

<style lang="scss">
  .btns {
    display: flex;
    flex-flow: row;
    gap: 10px;

    #test {
      margin-left: auto;
    }

    button {
      width: max-content;
      padding-left: 15px;
      padding-right: 15px;
    }
  }

  .error {
    position: sticky;
    top: 0;
    display: flex;
    justify-content: center;
    width: 100%;
    padding: 10px;
    background-color: rgb(221, 48, 48);
    text-transform: capitalize;
    color: white;
    margin-bottom: 15px;
  }
</style>
