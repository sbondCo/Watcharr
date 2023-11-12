<script lang="ts">
  import tooltip from "../actions/tooltip";

  export let rating: number | undefined = undefined;
  export let handleStarClick: (rating: number) => void;
  export let disableInteraction: boolean = false;
  export let minimal = false;
  export let direction: "top" | "bot" = "top";
  export let btnTooltip: string = "";

  let ratingsShown = false;
</script>

<button
  class={["rating", minimal ? (!rating ? "minimal" : "minimal-space") : ""].join(" ")}
  on:click={(ev) => {
    ev.stopPropagation();
    ratingsShown = !ratingsShown;
  }}
  on:mouseleave={(ev) => {
    ratingsShown = false;
    ev.currentTarget.blur();
  }}
  use:tooltip={{ text: btnTooltip, pos: "top", condition: !!btnTooltip && !ratingsShown }}
>
  <span>*</span>
  {#if !minimal}
    <span class={!rating && disableInteraction ? "unrated-text" : ""}>
      {rating ? rating : disableInteraction ? "Unrated" : "Rate"}
    </span>
  {:else if rating}
    <span>
      {rating}
    </span>
  {/if}
  {#if ratingsShown}
    <div class={["small-scrollbar", direction].join(" ")}>
      {#each [10, 9, 8, 7, 6, 5, 4, 3, 2, 1] as v}
        <button
          class="plain{rating === v ? ' active' : ''}"
          on:click={(ev) => {
            ev.stopPropagation();
            handleStarClick(v);
            ratingsShown = false;
          }}
        >
          {v}
        </button>
      {/each}
    </div>
  {/if}
</button>

<style lang="scss">
  button {
    padding: 3px;
    position: relative;
    font-family: "Rampart One";
    width: 100%;
    height: 100%;

    &.minimal span:first-child {
      letter-spacing: unset;
    }

    &.minimal-space span:first-child {
      letter-spacing: 5px;
    }

    span {
      &:first-child {
        color: $text-color;
        font-size: 39px;
        letter-spacing: 10px;
        line-height: 52px;
        height: 42px;
      }

      &:nth-child(2) {
        color: $text-color;
        font-size: 22px;
        height: 35px; // quick fix to make the rating num look centered - text-stroke makes it look not centered
      }
    }

    &:hover span,
    &:focus-visible span {
      color: $poster-rating-color;
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

      &.bot {
        top: calc(100% + 2px);
        border-radius: 0 0 4px 4px;
      }

      button {
        width: 100%;
        color: $text-color;
        fill: $text-color;
        -webkit-text-stroke: 0.5px $text-color;
        font-size: 20px;

        & :global(svg) {
          width: 100%;
          padding: 0 4.5px;
        }

        &:hover,
        &:focus-visible {
          background-color: rgb(100, 100, 100, 0.25);
        }
      }
    }
  }
</style>
