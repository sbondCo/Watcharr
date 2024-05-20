<script lang="ts">
  import Modal from "@/lib/Modal.svelte";
  import Setting from "@/lib/settings/Setting.svelte";
  import SettingsList from "@/lib/settings/SettingsList.svelte";
  import { toRelativeTime } from "@/lib/util/helpers";
  import type { AllTasksResponse } from "@/types";
  import axios from "axios";
  import { onMount } from "svelte";

  export let onClose: () => void;

  $: now = Date.now();

  let error: string;
  let formDisabled = false;
  let taskSchedule: AllTasksResponse[] = [];

  async function getAllTasks() {
    try {
      formDisabled = true;
      const res = await axios.get<AllTasksResponse[]>("/task/");
      taskSchedule = res.data;
      formDisabled = false;
      error = "";
    } catch (err) {
      console.error("getAllTasks failed!", err);
      error = `Failed to get all tasks from server`;
      formDisabled = false;
    }
  }

  onMount(() => {
    getAllTasks();
    const nowInterval = setInterval(() => {
      now = Date.now();
    }, 1000);
    const getTasksInterval = setInterval(() => {
      getAllTasks();
    }, 5000);
    return () => {
      clearInterval(nowInterval);
      clearInterval(getTasksInterval);
    };
  });
</script>

<Modal
  title="Tasks Schedule"
  desc="Want a routine task to occur more or less frequently? Configure it below."
  {onClose}
>
  {#if error}
    <span class="error">{error}!</span>
  {/if}
  <SettingsList>
    {#each taskSchedule as task}
      {@const nextRun = toRelativeTime((new Date(task.nextRun).getTime() - now) / 1000)}
      <Setting title={task.name}>
        Runs every&nbsp;
        <input type="text" placeholder="60" value={task.seconds} disabled={formDisabled} />
        &nbsp;seconds. Next{nextRun === "now" ? "" : " in"}
        {nextRun}.
      </Setting>
    {/each}
  </SettingsList>
</Modal>

<style lang="scss">
  input {
    width: 200px;
    margin-top: 5px;
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
