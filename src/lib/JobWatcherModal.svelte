<script lang="ts">
  import Icon from "@/lib/Icon.svelte";
  import Modal from "@/lib/Modal.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { notify } from "@/lib/util/notify";
  import { JobStatus, type GetJobResponse, type JobCreatedResponse } from "@/types";
  import axios from "axios";
  import { onDestroy, onMount } from "svelte";
  import { watchedList } from "@/store";

  export let modalTitle: string;
  // Promise that returns the job id to watch.
  export let getJobId: () => Promise<{ jobId: string } | undefined>;
  export let onClose: () => void;
  export let messages: {
    starting: string;
  };

  let step: "starting" | "errored" | "job-running" | "done" | "modal-closing" = "starting";
  let jobId: string | undefined;
  let currentTask: string | undefined;
  let latestJobStatus: GetJobResponse | undefined;

  async function startSync() {
    try {
      const r = await getJobId();
      console.log("startSync: Response:", r);
      if (!r?.jobId) {
        step = "errored";
        console.error("startSync: No jobId returned!");
        return;
      }
      jobId = r.jobId;
      step = "job-running";
      startJobWatcher();
    } catch (err) {
      console.error("startSync failed!", err);
      step = "errored";
    }
  }

  async function startJobWatcher() {
    if (!jobId) {
      console.error("startJobWatcher: No Job Id");
      notify({ text: "Unable to start job watcher, no job id.", type: "error" });
      return;
    }
    console.log("startJobWatcher: Starting..");
    let seqfailedJobReqs = 0;
    while (step === "job-running") {
      try {
        const r = await axios.get<GetJobResponse>(`/job/${jobId}`);
        console.log("jobWatcher: Got job data:", r.data);
        latestJobStatus = r.data;
        currentTask = r.data?.currentTask;
        if (r.data?.status === JobStatus.DONE) {
          step = "done";
        } else if (r.data?.status === JobStatus.CANCELLED) {
          step = "errored";
        }
        // If we get here without erroring, we can reset it to 0.
        seqfailedJobReqs = 0;
      } catch (err) {
        console.error("jobWatcher: Get job request failed!", seqfailedJobReqs, err);
        seqfailedJobReqs++;
      }
      if (seqfailedJobReqs >= 10) {
        console.error("jobWatcher: Failed 10 times in a row!");
        notify({
          text: "Status checker has failed 10 times in a row!",
          type: "error",
          time: 30000
        });
        step = "errored";
        break;
      }
      await new Promise((r) => setTimeout(r, 1000));
    }
    if (step !== "modal-closing") {
      // Update our watched list
      const nid = notify({ text: "Fetching updated watched list.", type: "loading" });
      try {
        const w = await axios.get("/watched");
        if (w?.data?.length > 0) {
          watchedList.update((wl) => (wl = w.data));
        }
        notify({ id: nid, text: "Fetched updated watched list.", type: "success" });
      } catch (err) {
        console.error("jobWatcher: Getting updated watched list failed!", err);
        notify({ id: nid, text: "Getting updated watched list failed!", type: "error" });
      }
    }
  }

  function modalClose() {
    if (step === "job-running") {
      notify({
        text: "Sync will continue in the background.. please refresh the page periodically to view your updated list or come back later.",
        time: 10000
      });
    }
    step = "modal-closing";
    onClose();
  }

  onMount(() => {
    startSync();
  });

  onDestroy(() => {
    step = "starting";
    jobId = undefined;
    currentTask = undefined;
    latestJobStatus = undefined;
  });
</script>

<Modal title={modalTitle} maxWidth="700px" onClose={modalClose}>
  <div class="ctr">
    {#if step === "done"}
      <Icon i="check" wh={60} />
    {:else if step === "errored"}
      <Icon i="close" wh={70} />
    {:else}
      <Spinner />
    {/if}
    <div>
      {#if step === "starting"}
        <h4 class="norm">Starting</h4>
        <span>{messages.starting}</span>
      {:else if step === "job-running"}
        <h4 class="norm">Running</h4>
        {#if currentTask}
          <span>{currentTask}</span>
        {/if}
      {:else if step === "done"}
        {#if !latestJobStatus?.errors || latestJobStatus?.errors?.length <= 0}
          <h4 class="norm">Finished</h4>
          <span>We have finished. Looks like there were no errors!</span>
        {:else}
          <h4 class="norm">
            Finished With {latestJobStatus?.errors?.length} Error{latestJobStatus?.errors
              ?.length === 1
              ? ""
              : "s"}
          </h4>
          <span>Job finished, but with errors:</span>
          <ul>
            {#each latestJobStatus?.errors as e}
              <li>{e}</li>
            {/each}
          </ul>
        {/if}
      {:else if step === "errored"}
        <h4 class="norm">We Errored!</h4>
        <span>We errored before starting the job or the job was cancelled.</span>
        {#if latestJobStatus?.errors && latestJobStatus?.errors?.length > 0}
          <ul>
            {#each latestJobStatus?.errors as e}
              <li>{e}</li>
            {/each}
          </ul>
        {/if}
      {:else}
        <h4 class="norm">Unknown State!</h4>
        <span>We're not sure of the current job status.</span>
      {/if}
    </div>
  </div>
</Modal>

<style lang="scss">
  .ctr {
    display: flex;
    flex-flow: row;
    gap: 20px;
    justify-content: start;
    align-items: start;
    margin-top: 25px;
    margin-bottom: 15px;
    margin-left: 15px;

    & > div:last-of-type {
      display: flex;
      flex-flow: column;
      gap: 8px;
      padding: 8px 0;

      & > span {
        font-style: italic;

        &::first-letter {
          text-transform: uppercase;
        }
      }

      & > ul {
        padding-left: 25px;

        li::first-letter {
          text-transform: uppercase;
        }
      }
    }
  }
</style>
