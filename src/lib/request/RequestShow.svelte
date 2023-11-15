<script lang="ts">
  import axios from "axios";
  import Modal from "../Modal.svelte";
  import type { ListBoxItem, SonarrSettings, SonarrTestResponse, TMDBShowDetails } from "@/types";
  import { notify } from "../util/notify";
  import DropDown from "../DropDown.svelte";
  import Setting from "../settings/Setting.svelte";
  import Spinner from "../Spinner.svelte";
  import ListBox from "../ListBox.svelte";

  const animeKeywordId = 210024;

  export let content: TMDBShowDetails;
  export let onClose: () => void;

  let servarrs: SonarrSettings[];
  let selectedServarrIndex: number;
  let inputsDisabled = true;
  let selectedServerCfg: SonarrTestResponse | undefined;
  let seasonItems: ListBoxItem[] = content.seasons.map((s) => {
    return {
      id: s.season_number,
      value: false,
      displayValue: s.name
    };
  });
  let addRequestRunning = false;

  async function getServers() {
    try {
      inputsDisabled = true;
      const r = await axios.get("/arr/son");
      if (r.data?.length > 0) {
        servarrs = r.data;
        selectedServarrIndex = 0;
      } else {
        notify({ text: "No servers found", type: "error" });
      }
      inputsDisabled = false;
    } catch (err) {
      notify({ text: "Failed to load servers", type: "error" });
    }
  }

  async function getConfig(name: string) {
    try {
      inputsDisabled = true;
      const r = await axios.get<SonarrTestResponse>(`/arr/son/config/${name}`);
      selectedServerCfg = r.data;
      inputsDisabled = false;
    } catch (err) {}
  }

  async function request() {
    let nid;
    try {
      if (!servarrs || !servarrs[selectedServarrIndex]) {
        notify({ text: "Must select a server", type: "error" });
        return;
      }
      if (!selectedServerCfg) {
        notify({ text: "No selected server config found", type: "error" });
        return;
      }
      addRequestRunning = true;
      nid = notify({ text: "Requesting", type: "loading" });
      const server = servarrs[selectedServarrIndex];
      const rootFolder = selectedServerCfg.rootFolders?.find((f) => f.id === server.rootFolder);
      if (!rootFolder) {
        console.error(
          "show request.. no root folder found with id:",
          server.rootFolder,
          "rf:",
          rootFolder
        );
        notify({ id: nid, text: "No Root Folder Found", type: "error" });
        return;
      }
      await axios.post("/arr/son/request", {
        serverName: server.name,
        title: content.name,
        year: new Date(content.first_air_date)?.getFullYear(),
        tvdbId: content.external_ids.tvdb_id,
        seriesType: content.keywords.results?.find((k) => k.id == animeKeywordId)
          ? "anime"
          : "standard",
        qualityProfile: server.qualityProfile,
        rootFolder: rootFolder.path,
        languageProfile: server.languageProfile,
        seasons: seasonItems.map((s) => {
          return {
            seasonNumber: s.id,
            monitored: s.value
          };
        })
      });
      notify({ id: nid, text: "Request complete", type: "success" });
      addRequestRunning = false;
      onClose();
    } catch (err) {
      console.error("content request failed!", err);
      addRequestRunning = false;
      notify({ id: nid, text: "Request failed!", type: "error" });
    }
  }

  $: {
    if (typeof selectedServarrIndex !== "undefined" && servarrs?.length > 0) {
      const s = servarrs[selectedServarrIndex];
      if (!s) {
        selectedServerCfg = undefined;
      } else {
        getConfig(s.name);
      }
    }
  }

  getServers();
</script>

<Modal title="Request" desc={content.name} {onClose}>
  <div class="req-ctr">
    {#if servarrs}
      {@const server = servarrs[selectedServarrIndex]}

      <div class="seasons-list">
        <ListBox bind:options={seasonItems} allCheckBox="All Seasons" />
      </div>

      {#if servarrs?.length > 1}
        <Setting title="Select the server to use">
          <DropDown
            placeholder="Select a server"
            bind:active={selectedServarrIndex}
            options={servarrs?.length > 0
              ? servarrs.map((s, i) => {
                  return { id: i, value: s.name };
                })
              : []}
          />
        </Setting>
      {/if}

      <button on:click={request} disabled={addRequestRunning}>Request</button>
    {:else}
      <Spinner />
    {/if}
  </div>
</Modal>

<style lang="scss">
  .req-ctr {
    display: flex;
    flex-flow: column;
    gap: 10px;
    height: 100%;

    .seasons-list {
      max-height: 500px;
      overflow: auto;
    }

    button {
      margin-top: auto;
      margin-left: auto;
      width: max-content;
    }
  }
</style>
