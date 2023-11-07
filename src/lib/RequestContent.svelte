<script lang="ts">
  import axios from "axios";
  import Modal from "./Modal.svelte";
  import type {
    DropDownItem,
    ListBoxItem,
    SonarrSettings,
    SonarrTestResponse,
    TMDBShowDetails
  } from "@/types";
  import { notify } from "./util/notify";
  import DropDown from "./DropDown.svelte";
  import Setting from "./settings/Setting.svelte";
  import Spinner from "./Spinner.svelte";
  import ListBox from "./ListBox.svelte";

  const animeKeywordId = 210024;

  export let content: TMDBShowDetails;
  export let onClose: () => void;

  let servarrs: SonarrSettings[];
  let selectedServarrIndex: number;
  let inputsDisabled = true;
  let selectedServerCfg: SonarrTestResponse | undefined;
  let seasonItems: ListBoxItem[] = content.seasons.map((s) => {
    return {
      id: s.id,
      value: false,
      displayValue: s.name
    };
  });

  async function getServers() {
    try {
      inputsDisabled = true;
      const r = await axios.get("/arr/son");
      if (r.data?.length > 0) {
        servarrs = r.data;
        selectedServarrIndex = 0;
      } else {
        notify({ text: "No servers founds", type: "error" });
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
    if (!servarrs || !servarrs[selectedServarrIndex]) {
      notify({ text: "Must select a server", type: "error" });
      return;
    }
    if (!selectedServerCfg) {
      notify({ text: "No selected server config found", type: "error" });
      return;
    }
    const server = servarrs[selectedServarrIndex];
    axios.post("/arr/son/request", {
      serverName: "Sonarr",
      tvdbId: content.external_ids.tvdb_id,
      seriesType: content.keywords.results?.find((k) => k.id == animeKeywordId)
        ? "anime"
        : "standard",
      qualityProfile: server.qualityProfile,
      rootFolder: selectedServerCfg.rootFolders[0].path,
      languageProfile: server.languageProfile
    });
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

      <ListBox bind:options={seasonItems} allCheckBox="All Seasons" />

      {JSON.stringify(seasonItems)}

      {#if servarrs?.length > 1}
        <Setting title="Select the server to use">
          <DropDown
            placeholder="Select a server"
            active={selectedServarrIndex}
            options={servarrs?.length > 0
              ? servarrs.map((s, i) => {
                  return { id: i, value: s.name };
                })
              : []}
          />
        </Setting>
      {/if}

      <!-- {servarrs ? JSON.stringify(server) : ""} -->

      <!-- {JSON.stringify(content.external_ids, undefined, 2)} -->

      <!-- {content.keywords.results?.find((k) => k.id == animeKeywordId) ? "anime" : "standard"} -->
      <button on:click={request}>Request</button>
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

    button {
      margin-top: auto;
      margin-left: auto;
      width: max-content;
    }
  }
</style>
