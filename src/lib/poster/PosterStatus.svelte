<script lang="ts">
  import type { WatchedStatus } from "@/types";
  import Icon from "../Icon.svelte";
  import tooltip from "../actions/tooltip";
  import { watchedStatuses } from "../util/helpers";

  export let status: WatchedStatus | undefined = undefined;
  export let handleStatusClick: (status: WatchedStatus | "DELETE") => void;

  let statusesShown = false;
</script>

<button
  class="status"
  on:click={(ev) => {
    ev.stopPropagation();
    statusesShown = !statusesShown;
  }}
  on:mouseleave={(ev) => {
    statusesShown = false;
    ev.currentTarget.blur();
  }}
>
  {#if status}
    <Icon i={watchedStatuses[status]} />
  {:else}
    <span class="no-icon">+</span>
  {/if}
  {#if statusesShown}
    <div class={["small-scrollbar", status ? "has-status" : ""].join(" ")}>
      {#each Object.entries(watchedStatuses) as [statusName, icon]}
        <button
          class="plain{status && status !== statusName ? ' not-active' : ''}"
          on:click={() => handleStatusClick(statusName)}
          use:tooltip={{ text: statusName }}
        >
          <Icon i={icon} />
        </button>
      {/each}
      {#if status}
        <button
          class="plain not-active"
          on:click={() => handleStatusClick("DELETE")}
          use:tooltip={{ text: "Delete" }}
        >
          <Icon i="trash" />
        </button>
      {/if}
    </div>
  {/if}
</button>

<style lang="scss">
  button {
    padding: 3px;
    position: relative;
    font-family: "Rampart One";
    width: 40%;

    .no-icon {
      color: $text-color;
      font-size: 30px;
      height: 52px;
    }

    &:hover .no-icon,
    &:focus-visible .no-icon {
      color: white;
      fill: white;
    }

    div {
      display: flex;
      flex-flow: column;
      position: absolute;
      width: 100%;
      height: 200px;
      background-color: $bg-color;
      top: calc(-100% - 170px);
      list-style: none;
      border-radius: 4px 4px 0 0;
      overflow: auto;
      scrollbar-width: thin;

      button {
        width: 100%;
        color: $text-color;
        fill: $text-color;
        -webkit-text-stroke: 0.5px $text-color;

        & :global(svg) {
          width: 100%;
          padding: 0 2px;
        }

        &:hover,
        &:focus-visible {
          background-color: rgb(100, 100, 100, 0.5);
        }
      }

      &.has-status :global(svg) {
        padding: 0 4.5px;
      }
    }
  }
</style>
