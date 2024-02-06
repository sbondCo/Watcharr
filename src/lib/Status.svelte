<script lang="ts">
  import Icon from "./Icon.svelte";
  import type { Watched, WatchedStatus } from "../types";
  import tooltip from "./actions/tooltip";
  import { toUnderstandableWatchedStatus } from "./util/helpers";

  export let w: Watched | undefined;
  export let onChange: (newStatus: WatchedStatus) => void;

  function handleStatusClick(s: WatchedStatus) {
    if (s === w?.status) return;
    onChange(s);
  }
</script>

<div class="status">
  <button
    class={w?.status && w?.status !== "PLANNED" ? "not-active" : ""}
    on:click={() => handleStatusClick("PLANNED")}
    use:tooltip={{ text: toUnderstandableWatchedStatus(w, "PLANNED"), pos: "top" }}
  >
    <Icon i="calendar" />
  </button>
  <button
    class={w?.status && w?.status !== "WATCHING" ? "not-active" : ""}
    on:click={() => handleStatusClick("WATCHING")}
    use:tooltip={{ text: toUnderstandableWatchedStatus(w, "WATCHING"), pos: "top" }}
  >
    <Icon i="clock" />
  </button>
  <button
    class={w?.status && w?.status !== "FINISHED" ? "not-active" : ""}
    on:click={() => handleStatusClick("FINISHED")}
    use:tooltip={{ text: toUnderstandableWatchedStatus(w, "FINISHED"), pos: "top" }}
  >
    <Icon i="check" />
  </button>
  <button
    class={w?.status && w?.status !== "HOLD" ? "not-active" : ""}
    on:click={() => handleStatusClick("HOLD")}
    use:tooltip={{ text: toUnderstandableWatchedStatus(w, "HOLD"), pos: "top" }}
  >
    <Icon i="pause" />
  </button>
  <button
    class={w?.status && w?.status !== "DROPPED" ? "not-active" : ""}
    on:click={() => handleStatusClick("DROPPED")}
    use:tooltip={{ text: toUnderstandableWatchedStatus(w, "DROPPED"), pos: "top" }}
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
