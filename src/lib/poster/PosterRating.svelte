<script lang="ts">
  import { userSettings } from "@/store";
  import tooltip from "../actions/tooltip";
  import { RatingStep, RatingSystem } from "@/types";
  import Icon from "../Icon.svelte";
  import { toShowableRating, toWhichThumb } from "../rating/helpers";

  export let rating: number | undefined = undefined;
  export let handleStarClick: (rating: number) => void;
  export let disableInteraction: boolean = false;
  /**
   * When not minimal, we will use user settings to
   * display ratings as they want.
   */
  export let minimal = false;
  export let direction: "top" | "bot" = "top";
  export let btnTooltip: string = "";
  export let hideStarWhenRated = false;

  let ratingsShown = false;

  $: settings = $userSettings;
  $: isUsingThumbs = settings && settings.ratingSystem === RatingSystem.Thumbs;
</script>

<button
  class={[
    "rating",
    minimal ? (!rating ? "minimal" : "minimal-space") : "",
    disableInteraction ? "interaction-disabled" : ""
  ].join(" ")}
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
  {#if !isUsingThumbs}
    <span class="star" style={hideStarWhenRated && rating ? "display: none" : ""}>*</span>
  {/if}
  {#if !minimal}
    <span class={[!rating && disableInteraction ? "unrated-text" : "", "rating-text"].join(" ")}>
      {#if rating}
        {#if isUsingThumbs}
          {@const r = toWhichThumb(rating)}
          {#if r === -1}
            <Icon i="thumb-down" />
          {:else if r === 0}
            <span
              style="display: flex; transform: translate(2px, -7px); font-size: 40px; font-family: 'Shrikhand';"
            >
              -
            </span>
          {:else if r === 1}
            <Icon i="thumb-up" />
          {/if}
        {:else}
          {toShowableRating(rating)}
        {/if}
      {:else if disableInteraction}
        Unrated
      {:else}
        Rate
      {/if}
    </span>

    {#if ratingsShown}
      <div class={["small-scrollbar", direction, isUsingThumbs ? "is-using-thumbs" : ""].join(" ")}>
        {#if isUsingThumbs}
          <button
            on:click={() => handleStarClick(1)}
            class="plain{rating && rating > 0 && rating < 5 ? ' active' : ''}"
            style="display: flex; justify-content: center;"
          >
            <i style="display: flex; width: 35px;"><Icon i="thumb-down" /></i>
          </button>
          <button
            on:click={() => handleStarClick(5)}
            class="plain{rating && rating > 4 && rating < 9 ? ' active' : ''}"
            style="display: flex; justify-content: center;"
          >
            <span
              style="display: flex; transform: translate(0px, -2px); font-size: 40px; height: 40px; font-family: 'Shrikhand';"
            >
              -
            </span>
          </button>
          <button
            on:click={() => handleStarClick(9)}
            class="plain{rating && rating > 8 ? ' active' : ''}"
            style="display: flex; justify-content: center;"
          >
            <i style="display: flex; width: 35px;"><Icon i="thumb-up" /></i>
          </button>
        {:else}
          {@const stars =
            settings?.ratingSystem == RatingSystem.OutOf5
              ? [5, 4, 3, 2, 1]
              : [10, 9, 8, 7, 6, 5, 4, 3, 2, 1]}
          {#each stars as v}
            <button
              class="plain{rating === v ? ' active' : ''}"
              on:click={(ev) => {
                ev.stopPropagation();
                handleStarClick(settings?.ratingSystem === RatingSystem.OutOf5 ? v * 2 : v);
                ratingsShown = false;
              }}
            >
              {#if settings?.ratingSystem === RatingSystem.OutOf100}
                {v * 10}
              {:else if settings?.ratingSystem === RatingSystem.OutOf5}
                {v}
              {:else}
                {v}
              {/if}
            </button>
          {/each}
        {/if}
      </div>
    {/if}
  {:else if rating}
    <span class="rating-text">
      {rating}
    </span>
  {/if}

  <!-- Ratings popup for usage with `minimal` -->
  {#if minimal && ratingsShown}
    <div class={["small-scrollbar", direction, isUsingThumbs ? "is-using-thumbs" : ""].join(" ")}>
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

      .unrated-text {
        display: flex;
        align-items: center;
        font-size: 15px !important;
      }
    }

    &.minimal span:first-child {
      letter-spacing: unset;
    }

    &.minimal-space span:first-child {
      letter-spacing: 5px;
    }

    span {
      &.star {
        color: $text-color;
        font-size: 39px;
        letter-spacing: 10px;
        line-height: 52px;
        height: 42px;
      }

      &.rating-text {
        color: $text-color;
        font-size: 22px;
        height: 35px; // quick fix to make the rating num look centered - text-stroke makes it look not centered

        & :global(svg) {
          height: 100%;
          padding: 5px;
        }
      }
    }

    &:hover span,
    &:focus-visible span {
      color: $poster-rating-color;
      fill: $poster-rating-color;
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

      &.is-using-thumbs {
        height: 150px;
        top: calc(-100% - 120px);

        button span {
          /* Overriding color so dash for thumbs ratings stays text-color */
          color: $text-color;
        }
      }
    }
  }
</style>
