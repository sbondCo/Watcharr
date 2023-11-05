<script lang="ts">
  import axios from "axios";
  import Modal from "./Modal.svelte";
  import type { DropDownItem, SonarrSettings, SonarrTestResponse, TMDBShowDetails } from "@/types";
  import { notify } from "./util/notify";
  import DropDown from "./DropDown.svelte";

  const animeKeywordId = 210024;

  export let content: TMDBShowDetails;
  export let onClose: () => void;

  let servarrs: SonarrSettings[];
  let selectedServarrIndex: number;
  let inputsDisabled = true;
  let selectedServerCfg: SonarrTestResponse | undefined;

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

<Modal title="Request" desc="Add Show To Sonarr" {onClose}>
  <div>Content</div>

  <DropDown
    placeholder="Select a server"
    active={selectedServarrIndex}
    options={servarrs?.length > 0
      ? servarrs.map((s, i) => {
          return { id: i, value: s.name };
        })
      : []}
  />

  {servarrs ? servarrs[selectedServarrIndex]?.host : ""}

  {JSON.stringify(content.external_ids, undefined, 2)}

  {content.keywords.results?.find((k) => k.id == animeKeywordId) ? "anime" : "standard"}
  <button on:click={request}>Request</button>
</Modal>
