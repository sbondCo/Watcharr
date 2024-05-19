<script lang="ts">
  import Modal from "@/lib/Modal.svelte";
  import Setting from "@/lib/settings/Setting.svelte";
  import SettingsList from "@/lib/settings/SettingsList.svelte";
  import type { AllTasksResponse } from "@/types";
  import axios from "axios";
  import { onMount } from "svelte";

  export let onClose: () => void;

  let error: string;
  let formDisabled = false;

  let taskSchedule: AllTasksResponse[] = [];

  async function getSettingsData() {
    try {
      formDisabled = true;
      const res = await axios.get<AllTasksResponse[]>("/task/");
      taskSchedule = res.data;
      formDisabled = false;
      error = "";
    } catch (err) {
      console.error("getSettingsData failed!", err);
      error = `Request Failed, check your Host and Key`;
      formDisabled = false;
    }
  }

  onMount(() => {
    getSettingsData();
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
      <Setting title={task.name}>
        <input type="text" placeholder="60" value={task.seconds} disabled={formDisabled} />
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
