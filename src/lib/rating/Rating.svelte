<script lang="ts">
  import { userSettings } from "@/store";
  import { RatingSystem } from "@/types";
  import { onMount } from "svelte";

  $: settings = $userSettings;

  export let rating: number | undefined;
  export let onChange: (newRating: number) => void;

  let hoveredRating: number | undefined;
  let shownRating: number | undefined;
  let ratingContainer: HTMLDivElement;
  let ratingText: HTMLSpanElement;
  let highlightContainer: HTMLDivElement;
  let normalContainer: HTMLDivElement;
  /**
   * Percentage star step.
   */
  let starStep = 10;
  $: stars =
    settings?.ratingSystem == RatingSystem.OutOf5
      ? [5, 4, 3, 2, 1]
      : [10, 9, 8, 7, 6, 5, 4, 3, 2, 1];

  const ratingDesc = [
    "Apalling",
    "Horrible",
    "Very Bad",
    "Bad",
    "Average",
    "Fine",
    "Good",
    "Very Good",
    "Great",
    "Masterpiece"
  ];

  function handleStarClick(r: number) {
    return;
    if (r === rating) return;
    onChange(r);
  }

  $: {
    if (hoveredRating !== undefined) shownRating = hoveredRating;
    else if (rating !== undefined) {
      shownRating = rating;
      resetRatingText();
    } else shownRating = undefined;
  }

  function resetRatingText() {
    if (!ratingText) {
      console.warn("resetRatingText: ratingText is not defined yet.");
      return;
    }
    if (typeof rating === "number" && rating > 0) {
      ratingText.innerText = ratingDesc[rating - 1];
    } else {
      ratingText.innerText = "Select Your Rating";
    }
  }

  function handleStarHover(
    ev: MouseEvent & {
      currentTarget: EventTarget & HTMLButtonElement;
    },
    r: number
  ) {
    // hoveredRating = r;
    // We set innerText instead of letting svelte update dom for us
    // since we need the new width of span right now.
    ratingText.innerText = ratingDesc[r - 1];
    const start = ratingContainer?.getBoundingClientRect()?.x;
    const starl = ev?.currentTarget?.getBoundingClientRect()?.left;
    const rb = ratingText?.getBoundingClientRect();
    ratingText.style.left = `${starl - start - rb.width / 2 + 11.5}px`;
    ratingText.style.transform = "unset";
  }

  function handleStarHoverEnd() {
    hoveredRating = undefined;
    ratingText.style.left = "50%";
    ratingText.style.transform = "translateX(-50%)";
    resetRatingText();
  }

  function handleMouseOver(
    ev: (TouchEvent | MouseEvent) & {
      currentTarget: EventTarget & HTMLDivElement;
    }
  ) {
    // console.log(ev);
    const rect = ev.currentTarget.getBoundingClientRect();
    const x = (ev instanceof MouseEvent ? ev.clientX : ev.touches[0].clientX) - rect.left; // rel to start of container
    const perc = Math.ceil(Math.round((x * 100) / rect.width) / starStep) * starStep;
    // console.log("PERC", Math.round((x * 100) / rect.width), perc);
    if (perc > 1) {
      let percToHighlight = perc;
      let percToHide = 100 - perc;
      if (starStep == 5) {
        // On step of 5, it looks nicer when we take highlight back one percent,
        // otherwise more than half the star looks highlighted... ew. This is only visual.
        percToHighlight--;
        percToHide++;
      }
      highlightContainer.style.display = "flex";
      highlightContainer.style.width = `${percToHighlight}%`;
      // We shrink this container too because of what seems to be a bug
      // where the text-stroke draws on the upper layer too, instead
      // of being hidden by the highlighted stars overlay.
      normalContainer.style.width = `${percToHide}%`;
    } else {
      highlightContainer.style.display = "none";
      highlightContainer.style.width = "0";
      normalContainer.style.width = "100%";
    }
    let hovR = perc;
    if (settings?.ratingSystem === RatingSystem.OutOf5) {
      hovR = hovR / 20;
    } else {
      hovR = hovR / 10;
    }
    hoveredRating = hovR;
    // console.log(rect);
    // console.log(x, `${perc}%`);
    // console.log(ev);
  }

  $: {
    // console.log("block", starStep, settings?.ratingStep, settings.ratingSystem);
    try {
      if (settings) {
        if (typeof settings.ratingSystem === "number") {
          // Set default star step for system.
          switch (settings.ratingSystem) {
            case RatingSystem.OutOf100:
              starStep = 1;
              break;

            case RatingSystem.OutOf5:
              starStep = 20;
              break;

            case RatingSystem.OutOf10:
            default:
              starStep = 5;
              break;
          }
        }
        // Override default with user set step if supported by this system.
        if (typeof settings.ratingStep === "number") {
          if (
            settings.ratingSystem === RatingSystem.OutOf5 ||
            settings.ratingSystem === RatingSystem.OutOf10 ||
            !settings.ratingSystem
          ) {
            if (
              settings.ratingStep === 0.1 ||
              settings.ratingStep === 0.5 ||
              settings.ratingStep === 1
            ) {
              starStep =
                settings.ratingSystem === RatingSystem.OutOf5
                  ? settings.ratingStep * 20
                  : settings.ratingStep * 10;
              console.debug("Set starStep from setting:", starStep);
            } else {
              starStep = settings.ratingSystem === RatingSystem.OutOf5 ? 20 : 10;
              console.debug("Set starStep using default:", starStep);
            }
          }
        }
      }
    } catch (err) {
      console.error("Failed to set startStep from settings:", err);
    }
  }

  onMount(() => {
    resetRatingText();
  });
</script>

<!-- {settings?.ratingSystem}
step: {starStep}
stepSetting: {settings?.ratingStep}
hoveredRating: {hoveredRating} -->

<!-- TODO make sure stars work good on mobile, should be able to hold down and adjust like a slider -->

<div class="rating-container" bind:this={ratingContainer}>
  <span bind:this={ratingText}></span>
  <div
    class="rating-wrap"
    on:pointermove={(ev) => handleMouseOver(ev)}
    on:touchmove={(ev) => handleMouseOver(ev)}
    on:mouseleave={() => handleStarHoverEnd()}
    on:touchend={() => handleStarHoverEnd()}
    role="button"
    tabindex="0"
  >
    <!-- The unlit stars. -->
    <div bind:this={normalContainer} class="rating the-normal-one">
      {#each stars as v}
        <button
          class="plain{shownRating === v ? ' TODOREMOVElit' : ''}"
          on:mouseenter={(ev) => handleStarHover(ev, v)}
          on:touchstart={(ev) => handleStarHover(ev, v)}
        >
          *
        </button>
      {/each}
    </div>
    <!-- Overlays on stars above to show them as highlighted. -->
    <div bind:this={highlightContainer} class="rating the-highlight-one">
      {#each stars as _}
        <button class="plain lit">*</button>
      {/each}
    </div>
    <!-- Hidden stars, just to keep correct layout since the two above are abolute. -->
    <div class="rating the-hidden-one-for-layout-reasons">
      {#each stars as v}
        <button class="plain" style="opacity: 0;" on:click={() => handleStarClick(v)}>*</button>
      {/each}
    </div>
  </div>
</div>

<style lang="scss">
  .rating-container {
    display: flex;
    flex-flow: column;
    overflow: visible;
    /* TODO responsivise for smol screenz */
    width: 377px;

    & > span {
      position: relative;
      transition:
        left 100ms ease-in,
        transform 100ms ease-in;
      max-width: max-content;
      left: 50%;
      transform: translateX(-50%);
    }
  }

  .rating-wrap {
    position: relative;
    user-select: none;
    cursor: pointer;
    width: max-content;
    margin-left: auto;
    margin-right: auto;
  }

  .rating {
    display: flex;
    align-items: center;
    color: $text-color;
    overflow: hidden;
    margin: 10px 0 10px 0;
    padding: 1px;

    &.the-highlight-one {
      width: 0%;
      display: none;
      overflow: hidden;
      -webkit-text-stroke: 1.5px gold;
      pointer-events: none;
    }

    &:not(.the-highlight-one) {
      flex-flow: row-reverse;
      justify-content: center;
    }

    &:not(.the-hidden-one-for-layout-reasons) {
      position: absolute;
      left: 0;
      top: 0;
    }

    &.the-normal-one {
      justify-content: unset;
      white-space: nowrap;
      left: unset;
      right: 0;
      -webkit-text-stroke: 1.5px $text-color;
      pointer-events: none;
    }

    button {
      font-size: 55px;
      font-family: "Rampart One";
      letter-spacing: 10px;
      line-height: 52px;
      height: 38px;

      &:global(.lit),
      &:global(.lit ~ button) {
        color: gold;
        -webkit-text-stroke: 1.5px gold;
      }

      @media screen and (max-width: 450px) {
        font-size: 50px;
      }

      @media screen and (max-width: 420px) {
        font-size: 45px;
      }
    }
  }
</style>
