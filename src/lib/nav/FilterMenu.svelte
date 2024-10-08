<script lang="ts">
  import { activeFilters, clearActiveFilters, serverFeatures } from "@/store";
  import type { Filters } from "@/types";
  import { get } from "svelte/store";
  import Icon from "../Icon.svelte";
  import tooltip from "../actions/tooltip";

  function filterClicked(type: keyof Filters, f: string) {
    const af = get(activeFilters);
    if (af[type]?.includes(f)) {
      af[type] = af[type]?.filter((a) => a !== f);
    } else {
      af[type]?.push(f);
    }
    activeFilters.update((a) => (a = af));
  }

  $: filter = $activeFilters;
  $: features = $serverFeatures;
</script>

<div class="menu">
  <div class="inner">
    <div class="title">
      <h4 class="norm sm-caps">type</h4>
      {#if filter?.type?.length > 0 || filter?.status?.length > 0}
        <button
          class="plain"
          use:tooltip={{ text: "Clear", pos: "left" }}
          on:click={() => clearActiveFilters()}
        >
          <Icon i="close-circle" wh={18} />
        </button>
      {/if}
    </div>
    <div class="type-filter">
      <button
        class={`${filter.type.includes("tv") ? "active" : ""}`}
        on:click={() => filterClicked("type", "tv")}
      >
        SHOW
      </button>
      <button
        class={`${filter.type.includes("movie") ? "active" : ""}`}
        on:click={() => filterClicked("type", "movie")}
      >
        MOVIE
      </button>
      {#if features.games}
        <button
          class={`${filter.type.includes("game") ? "active" : ""}`}
          on:click={() => filterClicked("type", "game")}
        >
          GAME
        </button>
      {/if}
    </div>
    <h4 class="norm sm-caps">status</h4>
    <button
      class={`plain ${filter.status.includes("planned") ? "on" : ""}`}
      on:click={() => filterClicked("status", "planned")}
    >
      planned
    </button>
    <button
      class={`plain ${filter.status.includes("watching") ? "on" : ""}`}
      on:click={() => filterClicked("status", "watching")}
    >
      watching
      {#if features.games}
        (playing)
      {/if}
    </button>
    <button
      class={`plain ${filter.status.includes("finished") ? "on" : ""}`}
      on:click={() => filterClicked("status", "finished")}
    >
      finished
      {#if features.games}
        (played)
      {/if}
    </button>
    <button
      class={`plain ${filter.status.includes("hold") ? "on" : ""}`}
      on:click={() => filterClicked("status", "hold")}
    >
      on hold
    </button>
    <button
      class={`plain ${filter.status.includes("dropped") ? "on" : ""}`}
      on:click={() => filterClicked("status", "dropped")}
    >
      dropped
    </button>
  </div>
</div>

<style lang="scss">
  div.menu {
    width: 200px;
    right: 47px;

    &:before {
      left: 38px;
    }
  }

  div.inner {
    & > h4 {
      margin-top: 8px;
      margin-bottom: 8px;
    }

    .title {
      display: flex;
      flex-flow: row;
      align-items: center;
      margin-bottom: 8px;
      gap: 5px;
      /* Always height of when clear filters btn is shown so there is no jump */
      min-height: 26px;

      button {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 28px;
        height: 26px;
        padding: 2px 3px;
        border-radius: 8px;

        &.manage-on {
          color: #f3555a;
          background-color: $text-color;
        }

        &:first-of-type {
          margin-left: auto;
        }
      }
    }

    & > button {
      text-transform: capitalize;
      position: relative;

      &.on::before {
        content: "\2713";
      }

      &::before {
        position: absolute;
        top: 4px;
        left: 7.5px;
        font-family:
          system-ui,
          -apple-system,
          BlinkMacSystemFont;
        font-size: 18px;
      }
    }

    .type-filter {
      display: flex;
      flex-flow: row;
      width: 100%;

      button {
        border-radius: 0;
        padding: 8px 0;
        width: 100%;

        &:first-of-type {
          border-radius: 5px 0 0 5px;
        }

        &:not(:first-of-type) {
          border-left: unset;
        }

        &:last-of-type {
          border-radius: 0 5px 5px 0;
        }
      }
    }
  }
</style>
