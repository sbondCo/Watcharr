<script lang="ts">
  import { activeSort } from "@/store";
  import { get } from "svelte/store";

  function sortClicked(type: string, modeType: string = "UPDOWN") {
    const af = get(activeSort);
    let mode: string;
    if (modeType === "UPDOWN") {
      mode = "UP";
      if (af[0] == type) {
        if (af[1] === "UP") {
          mode = "DOWN";
        } else if (af[1] === "DOWN") {
          mode = "";
        }
      }
    } else if (modeType === "TOGGLE") {
      mode = "ON";
      if (af[0] == type) {
        if (af[1] === "ON") {
          mode = "OFF";
        }
      }
    } else {
      console.error("filterClicked() ran without a valid modeType:", modeType);
      return;
    }
    activeSort.update((af) => (af = [type, mode]));
  }

  $: sort = $activeSort;
</script>

<div class="menu sort-menu">
  <div>
    <button
      class={`plain ${sort[0] == "DATEADDED" ? sort[1].toLowerCase() : ""}`}
      on:click={() => sortClicked("DATEADDED")}
    >
      Date Added
    </button>
    <button
      class={`plain ${sort[0] == "LASTCHANGED" ? sort[1].toLowerCase() : ""}`}
      on:click={() => sortClicked("LASTCHANGED")}
    >
      Last Changed
    </button>
    <button
      class={`plain ${sort[0] == "LASTFIN" ? sort[1].toLowerCase() : ""}`}
      on:click={() => sortClicked("LASTFIN")}
    >
      Last Finished
    </button>
    <button
      class={`plain ${sort[0] == "RATING" ? sort[1].toLowerCase() : ""}`}
      on:click={() => sortClicked("RATING")}
    >
      Rating
    </button>
    <button
      class={`plain ${sort[0] == "ALPHA" ? sort[1].toLowerCase() : ""}`}
      on:click={() => sortClicked("ALPHA")}
    >
      Alphabetical
    </button>
  </div>
</div>

<style lang="scss">
  div {
    width: 180px;
    right: 53px;

    &:before {
      left: 21px;
    }

    & > div {
      & > button {
        position: relative;

        &.down::before {
          content: "\2193";
        }

        &.up::before {
          content: "\2191";
        }

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
  }
</style>
