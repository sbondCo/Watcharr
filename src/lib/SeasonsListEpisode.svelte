<script lang="ts">
  import type { TMDBSeasonDetailsEpisode } from "@/types";

  export let ep: TMDBSeasonDetailsEpisode;
</script>

<li>
  {#if ep.still_path}
    <img src={`https://www.themoviedb.org/t/p/w227_and_h127_bestv2/${ep.still_path}`} alt="" />
  {:else}
    <div class="no-still" />
  {/if}
  <span>
    <b>{ep.episode_number}</b>
    {ep.name}
  </span>
  <span
    class="rating"
    title={`TMDB Rating: ${ep.vote_average} out of 10 (based on ${ep.vote_count} votes)`}
  >
    <span>*</span>
    {Math.round(ep.vote_average * 10) / 10}
  </span>
</li>

<style lang="scss">
  li {
    display: flex;
    flex-flow: row;
    gap: 8px;

    img,
    .no-still {
      width: 227px;
      min-width: 227px;
      height: 127px;
      min-height: 127px;
      border-radius: 10px;
      background-color: rgb(0, 0, 0);
      object-fit: fill;

      @media screen and (max-width: 590px) {
        width: 80%;
        height: auto;
      }

      @media screen and (max-width: 450px) {
        width: 100%;
      }
    }

    span {
      padding: 3px 5px;

      @media screen and (max-width: 590px) {
        text-align: center;
      }
    }

    .rating {
      display: flex;
      align-items: start;
      justify-content: center;
      font-size: 15px;
      color: $rating-color;
      font-weight: bolder;
      overflow: hidden;

      span {
        font-family: "Rampart One";
        -webkit-text-stroke: 1px $rating-color;
        font-size: 25px;
        line-height: 0.7;
        margin-top: 1.5px;
      }
    }
  }

  @media screen and (max-width: 590px) {
    li {
      align-items: center;
      flex-flow: column;
      width: 100%;
      height: 100%;
    }
  }
</style>
