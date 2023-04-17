<script lang="ts">
  export let rating: number | undefined;
  let hoveredRating: number | undefined;
  let shownRating: number | undefined;

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
    rating = r;
  }

  $: {
    if (hoveredRating !== undefined) shownRating = hoveredRating;
    else if (rating !== undefined) shownRating = rating;
  }

  function handleStarHover(r: number) {
    hoveredRating = r;
  }

  function handleStarHoverEnd() {
    hoveredRating = undefined;
  }
</script>

<div class="rating-container">
  <span>
    {#if typeof hoveredRating === "number"}
      {ratingDesc[hoveredRating]}
    {:else if typeof rating === "number"}
      {ratingDesc[rating]}
    {:else}
      Select Your Rating
    {/if}
  </span>
  <div class="rating">
    {#each [9, 8, 7, 6, 5, 4, 3, 2, 1, 0] as v}
      <button
        class="plain{shownRating === v ? ' lit' : ''}"
        on:click={() => handleStarClick(v)}
        on:mouseenter={() => handleStarHover(v)}
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
  }

  .rating {
    display: flex;
    flex-flow: row-reverse;
    align-items: center;
    justify-content: center;
    color: black;
    -webkit-text-stroke: 1.5px black;
    cursor: pointer;
    overflow: hidden;
    margin: 10px 0 10px 0;

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

      @media screen and (max-width: 500px) {
        font-size: 40px;
      }
    }
  }
</style>
