<script lang="ts">
  import { onMount } from "svelte";

  export let rating: number | undefined;
  export let onChange: (newRating: number) => void;

  let hoveredRating: number | undefined;
  let shownRating: number | undefined;
  let ratingContainer: HTMLDivElement;
  let ratingText: HTMLSpanElement;

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
    onChange(r);
  }

  $: {
    if (hoveredRating !== undefined) shownRating = hoveredRating;
    else if (rating !== undefined) shownRating = rating;
    else shownRating = undefined;
  }

  function resetRatingText() {
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
    hoveredRating = r;
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

  onMount(() => {
    resetRatingText();
  });
</script>

<div class="rating-container" bind:this={ratingContainer}>
  <span bind:this={ratingText}></span>
  <div class="rating">
    {#each [10, 9, 8, 7, 6, 5, 4, 3, 2, 1] as v}
      <button
        class="plain{shownRating === v ? ' lit' : ''}"
        on:click={() => handleStarClick(v)}
        on:mouseenter={(ev) => handleStarHover(ev, v)}
        on:mouseleave={() => handleStarHoverEnd()}
      >
        *
      </button>
    {/each}
  </div>
</div>

<style lang="scss">
  .rating-container {
    display: flex;
    flex-flow: column;
    overflow: visible;

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

  .rating {
    display: flex;
    flex-flow: row-reverse;
    align-items: center;
    justify-content: center;
    color: $text-color;
    -webkit-text-stroke: 1.5px $text-color;
    cursor: pointer;
    overflow: hidden;
    margin: 10px 0 10px 0;
    padding: 1px;

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

      @media screen and (max-width: 400px) {
        font-size: 45px;
      }
    }
  }
</style>
