<!-- 
  /import/trakt collects the username for the trakt
  user to import data from.
 -->

<script lang="ts">
  import JobWatcherModal from "@/lib/JobWatcherModal.svelte";
  import { notify } from "@/lib/util/notify";
  import { userInfo } from "@/store";
  import type { JobCreatedResponse } from "@/types";
  import axios from "axios";

  $: user = $userInfo;

  let modalOpen = false;
  let traktUsername = "";

  async function startJob(): Promise<{ jobId: string } | undefined> {
    const r = await axios.post<JobCreatedResponse>("/import/trakt", { username: traktUsername });
    console.log("startSync: Response:", r.data);
    if (!r.data.jobId) {
      notify({
        type: "error",
        text: "No job id was returned! Cannot watch job, if it even started."
      });
      return;
    }
    return { jobId: r.data.jobId };
  }
</script>

<div class="content">
  <div class="inner">
    <h2>Trakt Import</h2>
    <h5 class="norm">
      Provide the username to your <b>public</b> Trakt profile to start the import job.
    </h5>

    <input
      type="text"
      placeholder={user?.username ?? "Trakt Username"}
      bind:value={traktUsername}
    />
    <button on:click={() => (modalOpen = true)}>Start Import</button>
  </div>

  {#if modalOpen}
    <JobWatcherModal
      modalTitle="Trakt Import"
      messages={{ starting: "Trakt import job is starting" }}
      getJobId={startJob}
      onClose={() => (modalOpen = false)}
    />
  {/if}
</div>

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;
    padding: 0 30px 30px 30px;

    .inner {
      display: flex;
      flex-flow: column;
      min-width: 400px;
      max-width: 400px;
      overflow: hidden;

      @media screen and (max-width: 420px) {
        min-width: 100%;
      }
    }
  }

  input,
  button {
    margin-top: 15px;
  }

  ul {
    display: flex;
    flex-flow: column;
    gap: 5px;
    margin: 10px;
    list-style: none;

    li {
      display: flex;
      flex-flow: row;
      align-items: center;
      padding: 10px;
      background-color: $accent-color;
      border-radius: 5px;

      a {
        margin-left: auto;

        button {
          width: max-content;
        }
      }
    }
  }
</style>
