<script lang="ts">
  import type {
    ArrDetailsResponse,
    ArrInfoResponse,
    ArrRequestResponse,
    ContentType
  } from "@/types";
  import axios from "axios";
  import { onMount } from "svelte";
  import Icon from "../Icon.svelte";
  import tooltip from "../actions/tooltip";
  import { msToAmountsOfTime } from "../util/helpers";
  import { notify } from "../util/notify";

  export let type: ContentType;
  export let tmdbId: number;
  export let openRequestModal: () => void;

  let existingRequest: ArrRequestResponse | undefined;
  let info: ArrInfoResponse | undefined;
  let status: ArrDetailsResponse | "available" | undefined;
  let estimatedCompletionIn: string | undefined;

  async function getInfo() {
    try {
      if (!existingRequest) {
        console.warn("getInfo called before existingRequest exists.");
        return;
      }
      const resp = await axios.get<ArrInfoResponse>(
        `/arr/${type === "movie" ? "rad" : "son"}/info/${existingRequest.serverName}/${existingRequest.arrId}`
      );
      if (resp?.data) {
        info = resp.data;
        if (info.hasFile) {
          status = "available";
        } else {
          getStatus();
        }
      }
    } catch (err) {
      console.error("ArrRequestButton: getInfo failed!", err);
      notify({
        text: `Failed when getting info from ${type === "movie" ? "Radarr" : "Sonarr"}`,
        type: "error"
      });
    }
  }

  async function getStatus(ev: MouseEvent | undefined = undefined) {
    try {
      if (!existingRequest) {
        console.warn("getStatus called before existingRequest exists.");
        return;
      }
      const statusResp = await axios.get<ArrDetailsResponse>(
        `/arr/${type === "movie" ? "rad" : "son"}/status/${existingRequest.serverName}/${existingRequest.arrId}`
      );
      if (statusResp?.data) {
        status = statusResp.data;
        const estMs =
          new Date(status.estimatedCompletionTime).getTime() - new Date(Date.now()).getTime();
        const est = msToAmountsOfTime(estMs);
        if (est.days > 0) {
          estimatedCompletionIn = `${est.days} Day${est.days > 1 ? "s" : ""}`;
        } else if (est.hours > 0) {
          estimatedCompletionIn = `${est.hours} Hour${est.hours > 1 ? "s" : ""}`;
        } else if (est.minutes > 0) {
          estimatedCompletionIn = `${est.minutes} Minute${est.minutes > 1 ? "s" : ""}`;
        } else if (est.seconds > 0) {
          estimatedCompletionIn = `${est.seconds} Second${est.seconds > 1 ? "s" : ""}`;
        } else {
          estimatedCompletionIn = undefined;
        }
      }
      if (ev?.target) {
        setTimeout(() => {
          (ev?.target as HTMLButtonElement)?.blur();
        }, 250);
      }
    } catch (err) {
      console.error("ArrRequestButton: getStatus failed!", err);
      notify({
        text: `Failed to fetch content status from ${type === "movie" ? "Radarr" : "Sonarr"}`,
        type: "error"
      });
    }
  }

  async function lookForExisting() {
    try {
      const existingRequestResp = await axios.get<ArrRequestResponse>(
        `/arr/${type === "movie" ? "rad" : "son"}/request/${tmdbId}`
      );
      if (existingRequestResp?.data && existingRequestResp?.data?.arrId) {
        existingRequest = existingRequestResp?.data;
        getInfo();
      }
    } catch (err) {
      console.error("ArrRequestButton: lookForExisting failed!", err);
      notify({
        text: `Failed when looking for an existing request for this content`,
        type: "error"
      });
    }
  }

  onMount(() => {
    lookForExisting();
  });
</script>

{#if typeof status === "object"}
  <button
    on:click={getStatus}
    use:tooltip={{
      text:
        status.status === "downloading"
          ? estimatedCompletionIn
            ? `Done in ${estimatedCompletionIn}`
            : "Estimation Unavailable"
          : status.status === "paused"
            ? "Download has been paused"
            : "",
      pos: "bot"
    }}
  >
    <div><Icon i="refresh" /></div>
    <span>
      {status.status}
      {status.status === "downloading" ? (status?.progress ? `(${status?.progress}%)` : "") : ""}
    </span>
  </button>
{:else if status === "available"}
  <button disabled>Available</button>
{:else}
  <button on:click={openRequestModal}>Request</button>
{/if}

<style lang="scss">
  button {
    max-width: fit-content;
    text-transform: capitalize;
    position: relative;

    & > div {
      display: none;
      position: absolute;
      left: 50%;
      top: 50%;
      transform: translate(-50%, -50%);
    }

    &:hover:not(:focus) {
      & > div {
        display: block;
      }

      & > span {
        visibility: hidden;
      }
    }
  }
</style>
