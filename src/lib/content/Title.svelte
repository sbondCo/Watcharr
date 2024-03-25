<script lang="ts">
  export let homepage: string | undefined;
  export let title: string;
  export let releaseYear: string | number;
  export let voteAverage: number;
  export let voteCount: number;

  // if voteAvg bigger than 10, it is out of 100, so no need to * by 10
  const vote = Math.round(voteAverage > 10 ? voteAverage : voteAverage * 10) / 10;
</script>

<span class="title-container">
  <span class="title">
    {#if homepage}
      <a href={homepage} target="_blank">{title}</a>
    {:else}
      <span class="t">{title}</span>
    {/if}
    <span class="year">{releaseYear}</span>
  </span>
  <span class="rating" title={`TMDB Rating: ${vote} out of 10 (based on ${voteCount} votes)`}>
    <span>*</span>
    {vote}
  </span>
</span>

<style lang="scss">
  .title-container {
    display: flex;
    gap: 10px;

    .title {
      a,
      span.t {
        color: white;
        text-decoration: none;
        font-size: 30px;
        font-weight: bold;
        padding-right: 3px;
      }

      span.year {
        font-size: 20px;
        color: rgba($color: #fff, $alpha: 0.7);
      }
    }

    .rating {
      display: flex;
      align-items: start;
      justify-content: center;
      gap: 5px;
      color: green;
      margin-left: auto;
      font-size: 22px;
      color: gold;
      font-weight: bolder;

      span {
        font-family: "Rampart One";
        -webkit-text-stroke: 1px gold;
        font-size: 40px;
        line-height: 0.7;
        margin-top: 7px;
      }
    }
  }
</style>