<script lang="ts">
  import type { WatchedStatus } from "@/types";
  import Icon from "../Icon.svelte";
  import tooltip from "../actions/tooltip";
  import { toUnderstandableStatus, watchedStatuses } from "../util/helpers";

  export let status: WatchedStatus | undefined = undefined;
  export let handleStatusClick: (status: WatchedStatus | "DELETE") => void;
  export let direction: "top" | "bot" = "top";
  export let width = "40%";
  export let small = false;
  export let btnTooltip: string = "";
  export let disableInteraction: boolean = false;
  export let isForGame = false;

  let statusesShown = false;
</script>

<button
  class={["status", disableInteraction ? "interaction-disabled" : ""].join(" ")}
  style={`width: ${width};`}
  on:click={(ev) => {
    ev.stopPropagation();
    statusesShown = !statusesShown;
  }}
  on:mouseleave={(ev) => {
    statusesShown = false;
  }}
  use:tooltip={{ text: btnTooltip, pos: "top", condition: !!btnTooltip && !statusesShown }}
>
  {#if status}
    <Icon i={watchedStatuses[status]} />
  {:else}
    <span class={["no-icon", small ? "small" : ""].join(" ")}>+</span>
  {/if}
  <div
    class={[
      statusesShown ? "shown" : "",
      "small-scrollbar",
      status ? "has-status" : "",
      direction
    ].join(" ")}
  >
    {#each Object.entries(watchedStatuses) as [statusName, icon]}
      <button
        class="plain{status && status !== statusName ? ' not-active' : ''}"
        on:click={() => handleStatusClick(statusName)}
        use:tooltip={{
          text: toUnderstandableStatus(statusName, isForGame)
        }}
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
</button>

<style lang="scss">
  button {
    padding: 3px;
    position: relative;
    font-family: "Rampart One";
    height: 100%;

    &.interaction-disabled {
      pointer-events: none;
      cursor: default;
      background-color: transparent;
      border: unset;
      fill: white;
      color: white;

      span {
        color: white !important;
      }
    }

    .no-icon {
      color: $text-color;
      font-size: 30px;
      height: 52px;

      &.small {
        height: 30px;
        line-height: 22px;
      }
    }

    &:hover .no-icon,
    &:focus-visible .no-icon {
      color: $bg-color;
      fill: $bg-color;
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
      z-index: 40;
      box-shadow: 0px 0px 1px #000;

      &:not(.shown) {
        display: none;
      }

      &.bot {
        top: calc(100% + 2px);
        border-radius: 0 0 4px 4px;
      }

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
