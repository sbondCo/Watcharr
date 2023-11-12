<script lang="ts">
  import axios from "axios";
  import Modal from "../Modal.svelte";
  import type { RadarrSettings, RadarrTestResponse, TMDBMovieDetails } from "@/types";
  import { notify } from "../util/notify";
  import DropDown from "../DropDown.svelte";
  import Setting from "../settings/Setting.svelte";
  import Spinner from "../Spinner.svelte";

  export let content: TMDBMovieDetails;
  export let onClose: () => void;

  let servarrs: RadarrSettings[];
  let selectedServarrIndex: number;
  let inputsDisabled = true;
  let selectedServerCfg: RadarrTestResponse | undefined;
  let addRequestRunning = false;

  async function getServers() {
    try {
      inputsDisabled = true;
      const r = await axios.get("/arr/rad");
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
      const r = await axios.get<RadarrTestResponse>(`/arr/rad/config/${name}`);
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
          "movie request.. no root folder found with id:",
          server.rootFolder,
          "rf:",
          rootFolder
        );
        notify({ id: nid, text: "No Root Folder Found", type: "error" });
        return;
      }
      await axios.post("/arr/rad/request", {
        serverName: server.name,
        title: content.title,
        year: new Date(content.release_date)?.getFullYear(),
        tmdbId: content.id,
        qualityProfile: server.qualityProfile,
        rootFolder: rootFolder.path
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

<Modal title="Request" desc={content.title} {onClose}>
  <div class="req-ctr">
    {#if servarrs}
      {@const server = servarrs[selectedServarrIndex]}

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

    button {
      margin-top: auto;
      margin-left: auto;
      width: max-content;
    }
  }
</style>
