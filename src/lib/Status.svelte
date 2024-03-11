<script lang="ts">
  import Icon from "./Icon.svelte";
  import type { WatchedStatus } from "../types";
  import tooltip from "./actions/tooltip";
  import { toUnderstandableStatus } from "./util/helpers";

  export let status: WatchedStatus | undefined;
  export let isForGame: boolean = false;
  export let onChange: (newStatus: WatchedStatus) => void;

  function handleStatusClick(s: WatchedStatus) {
    if (s === status) return;
    onChange(s);
  }
</script>

<div class="status">
  <button
    class={status && status !== "PLANNED" ? "not-active" : ""}
    on:click={() => handleStatusClick("PLANNED")}
    use:tooltip={{ text: toUnderstandableStatus("PLANNED", isForGame), pos: "top" }}
  >
    <Icon i="calendar" />
  </button>
  <button
    class={status && status !== "WATCHING" ? "not-active" : ""}
    on:click={() => handleStatusClick("WATCHING")}
    use:tooltip={{ text: toUnderstandableStatus("WATCHING", isForGame), pos: "top" }}
  >
    <Icon i="clock" />
  </button>
  <button
    class={status && status !== "FINISHED" ? "not-active" : ""}
    on:click={() => handleStatusClick("FINISHED")}
    use:tooltip={{ text: toUnderstandableStatus("FINISHED", isForGame), pos: "top" }}
  >
    <Icon i="check" />
  </button>
  <button
    class={status && status !== "HOLD" ? "not-active" : ""}
    on:click={() => handleStatusClick("HOLD")}
    use:tooltip={{ text: toUnderstandableStatus("HOLD", isForGame), pos: "top" }}
  >
    <Icon i="pause" />
  </button>
  <button
    class={status && status !== "DROPPED" ? "not-active" : ""}
    on:click={() => handleStatusClick("DROPPED")}
    use:tooltip={{ text: toUnderstandableStatus("DROPPED", isForGame), pos: "top" }}
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
      padding: 5px 10px;
    }
  }
</style>
