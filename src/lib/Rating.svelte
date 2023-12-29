<script lang="ts">
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

  function handleStarHover(
    ev: MouseEvent & {
      currentTarget: EventTarget & HTMLButtonElement;
    },
    r: number
  ) {
    hoveredRating = r;
    const start = ratingContainer?.getBoundingClientRect()?.x;
    const starl = ev?.currentTarget?.getBoundingClientRect()?.left;
    const rb = ratingText?.getBoundingClientRect();
    ratingText.style.left = `${starl - start - rb.width / 2 + 10}px`;
  }

  function handleStarHoverEnd() {
    hoveredRating = undefined;
    ratingText.style.left = `${0}px`;
  }
</script>

<div class="rating-container" bind:this={ratingContainer}>
  <span bind:this={ratingText}>
    {#if typeof hoveredRating === "number"}
      {ratingDesc[hoveredRating - 1]}
    {:else if typeof rating === "number" && rating > 0}
      {ratingDesc[rating - 1]}
    {:else}
      Select Your Rating
    {/if}
  </span>
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
      transition: left 100ms ease-in;
      text-align: center;
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
