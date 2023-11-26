<script lang="ts">
  import { activeFilters } from "@/store";
  import type { Filters } from "@/types";
  import { get } from "svelte/store";

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
</script>

<div class="menu">
  <div class="inner">
    <h4 class="norm sm-caps">type</h4>
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
    </button>
    <button
      class={`plain ${filter.status.includes("finished") ? "on" : ""}`}
      on:click={() => filterClicked("status", "finished")}
    >
      finished
    </button>
    <button
      class={`plain ${filter.status.includes("hold") ? "on" : ""}`}
      on:click={() => filterClicked("status", "hold")}
    >
      held
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
    width: 180px;
    right: 35px;

    &:before {
      left: 3px;
    }
  }

  div.inner {
    h4 {
      margin-bottom: 8px;

      &:not(:first-of-type) {
        margin-top: 8px;
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
        left: 12px;
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
          border-right: 0;
          border-radius: 5px 0 0 5px;
        }

        &:last-of-type {
          border-radius: 0 5px 5px 0;
        }
      }
    }
  }
</style>
