<script lang="ts">
  import Icon from "./Icon.svelte";
  import type { WatchedStatus } from "../types";

  export let status: WatchedStatus | undefined;
  export let onChange: (newStatus: WatchedStatus) => void;
</script>

<div class="status">
  <button
    class={status && status !== "PLANNED" ? "not-active" : ""}
    on:click={() => onChange("PLANNED")}
  >
    <Icon i="calendar" />
  </button>
  <button
    class={status && status !== "WATCHING" ? "not-active" : ""}
    on:click={() => onChange("WATCHING")}
  >
    <Icon i="clock" />
  </button>
  <button
    class={status && status !== "FINISHED" ? "not-active" : ""}
    on:click={() => onChange("FINISHED")}
  >
    <Icon i="check" />
  </button>
  <button class={status && status !== "HOLD" ? "not-active" : ""} on:click={() => onChange("HOLD")}>
    <Icon i="pause" />
  </button>
  <button
    class={status && status !== "DROPPED" ? "not-active" : ""}
    on:click={() => onChange("DROPPED")}
  >
    <Icon i="thumb-down" />
  </button>
</div>

<style lang="scss">
  .status {
    display: flex;
    flex-flow: row;
    gap: 10px;
    width: 100%;
    height: 100%;

    button {
      font-size: 10px;

      &:hover {
        fill: white;
      }
    }
  }
</style>
