<script lang="ts">
  import { wlDetailedView } from "@/store";
  import type { WLDetailedViewOption } from "@/types";
  import { get } from "svelte/store";
  import { page } from "$app/stores";

  function detailClicked(d: WLDetailedViewOption) {
    let dv = get(wlDetailedView);
    if (dv.includes(d)) {
      dv = dv.filter((a) => a !== d);
    } else {
      dv.push(d);
    }
    wlDetailedView.update((a) => (a = dv));
  }

  $: dve = $wlDetailedView;
</script>

<div class={`menu${$page.url?.pathname.startsWith("/search/") ? " on-search-page" : ""}`}>
  <div class="inner">
    <h4 class="norm sm-caps">Shown Details</h4>
    <button
      class={`plain ${dve?.includes("statusRating") ? "on" : ""}`}
      on:click={() => detailClicked("statusRating")}
    >
      Status & Rating
    </button>
    <button
      class={`plain ${dve?.includes("lastWatched") ? "on" : ""}`}
      on:click={() => detailClicked("lastWatched")}
    >
      Watching Season
    </button>
    <button
      class={`plain ${dve?.includes("dateAdded") ? "on" : ""}`}
      on:click={() => detailClicked("dateAdded")}
    >
      Date Added
    </button>
    <button
      class={`plain ${dve?.includes("dateModified") ? "on" : ""}`}
      on:click={() => detailClicked("dateModified")}
    >
      Date Modified
    </button>
  </div>
</div>

<style lang="scss">
  div.menu {
    width: 200px;
    right: 92px;

    &:before {
      left: 3px;
    }
  }

  div.menu.on-search-page:before {
    left: 86px;
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
  }
</style>
