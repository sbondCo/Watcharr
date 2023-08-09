<script lang="ts">
  import Icon from "./Icon.svelte";
  import type { WatchedStatus } from "../types";
  import tooltip from "./actions/tooltip";

  export let status: WatchedStatus | undefined;
  export let onChange: (newStatus: WatchedStatus) => void;
</script>

<div class="status">
  <button
    class={status && status !== "PLANNED" ? "not-active" : ""}
    on:click={() => onChange("PLANNED")}
    use:tooltip={{ text: "Planned", pos: "top" }}
  >
    <Icon i="calendar" />
  </button>
  <button
    class={status && status !== "WATCHING" ? "not-active" : ""}
    on:click={() => onChange("WATCHING")}
    use:tooltip={{ text: "Watching", pos: "top" }}
  >
    <Icon i="clock" />
  </button>
  <button
    class={status && status !== "FINISHED" ? "not-active" : ""}
    on:click={() => onChange("FINISHED")}
    use:tooltip={{ text: "Finished", pos: "top" }}
  >
    <Icon i="check" />
  </button>
  <button
    class={status && status !== "HOLD" ? "not-active" : ""}
    on:click={() => onChange("HOLD")}
    use:tooltip={{ text: "On Hold", pos: "top" }}
  >
    <Icon i="pause" />
  </button>
  <button
    class={status && status !== "DROPPED" ? "not-active" : ""}
    on:click={() => onChange("DROPPED")}
    use:tooltip={{ text: "Dropped", pos: "top" }}
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
    }
  }
</style>
