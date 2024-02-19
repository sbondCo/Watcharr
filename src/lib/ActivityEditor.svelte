<script lang="ts">
  import { updateActivity, removeActivity } from "@/lib/util/api";
  import Modal from "./Modal.svelte";
  import type { Activity } from "@/types";

  export let watchedId: number;
  export let activity: Activity;
  export let activityMessage: string;
  export let onClose: () => void;

  let isDateTimeChanged: boolean;
  let currentDateObject = new Date(Date.parse(activity.customDate ?? activity.createdAt));
  let currentDateString = dateToInputDateString(currentDateObject);
  let currentTimeString = dateToInputTimeString(currentDateObject);
  let selectedDateString = currentDateString;
  let selectedTimeString = currentTimeString;

  $: {
    isDateTimeChanged =
      currentDateString != selectedDateString || currentTimeString != selectedTimeString;
  }

  function dateToInputDateString(date: Date) {
    const year = date.getFullYear();
    const month = (date.getMonth() + 1).toString().padStart(2, "0");
    const day = date.getDate().toString().padStart(2, "0");
    return `${year}-${month}-${day}`;
  }

  function dateToInputTimeString(date: Date) {
    const hours = date.getHours().toString().padStart(2, "0");
    const minutes = date.getMinutes().toString().padStart(2, "0");
    return `${hours}:${minutes}`;
  }

  function update(date: string, time: string) {
    const epochMillis = Date.parse(`${date} ${time}`);
    const dateObj = new Date(epochMillis);
    updateActivity(watchedId, activity.id, dateObj);
    onClose();
  }

  function remove() {
    removeActivity(watchedId, activity.id);
    onClose();
  }
</script>

<Modal title="Edit Activity" desc={activityMessage} maxWidth="400px" {onClose}>
  <div class="centered">
    <h3>Date</h3>
    <input id="activity-date" type="date" bind:value={selectedDateString} />
    <h3>Time</h3>
    <input id="activity-time" type="time" bind:value={selectedTimeString} />

    <div class="button-row">
      <button class="danger" on:click={remove}>Delete</button>
      <button
        on:click={() => update(selectedDateString, selectedTimeString)}
        disabled={!isDateTimeChanged}>Update</button
      >
    </div>
  </div>
</Modal>

<style lang="scss">
  .centered {
    display: flex;
    flex-flow: column;
    gap: 10px;
    height: 100%;

    h3 {
      font-size: 16px;
      font-family:
        sans-serif,
        system-ui,
        -apple-system,
        BlinkMacSystemFont;
    }

    .button-row {
      display: flex;
      flex-flow: row;
      justify-content: space-between;
      margin-top: 10px;

      button {
        margin-top: auto;
        width: max-content;
      }
    }
  }
</style>
