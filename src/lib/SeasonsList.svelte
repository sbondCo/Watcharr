<script lang="ts">
  import type { TMDBSeasonDetails, TMDBShowSeason } from "@/types";
  import axios from "axios";
  import Spinner from "./Spinner.svelte";
  import Error from "./Error.svelte";

  export let tvId: number;
  export let seasons: TMDBShowSeason[];

  let activeSeason = 1;
  let seasonDetailsReq: Promise<TMDBSeasonDetails>;
  let seasonsEl: HTMLUListElement, episodesEl: HTMLDivElement;

  async function sdr(seasonNum: number) {
    return (await axios.get(`/content/tv/${tvId}/season/${seasonNum}`)).data as TMDBSeasonDetails;
  }

  $: {
    seasonDetailsReq = sdr(activeSeason);
  }

  // onMount(() => {
  //   episodesEl.style.height = `${seasonsEl.clientHeight}px`;
  //   seasonsEl.addEventListener("resize", () => {
  //     console.log("resize");
  //     episodesEl.style.height = `${seasonsEl.clientHeight}px`;
  //   });
  // });
</script>

<div class="ctr">
  <ul class="seasons" bind:this={seasonsEl}>
    <!-- {#each seasons.sort((a, b) => Date.parse(b.air_date) - Date.parse(a.air_date)) as season} -->
    {#each seasons as season}
      <button
        class={`plain${activeSeason === season.season_number ? " active" : ""}`}
        on:click={() => {
          activeSeason = season.season_number;
        }}
      >
        <h1>{season.name}</h1>
        {#if season.air_date}
          <h2>{new Date(Date.parse(season.air_date)).getFullYear()}</h2>
        {:else if season.season_number > 0}
          <h2>TBD</h2>
        {/if}
      </button>
    {/each}
    <div class="last" />
  </ul>

  <div class="episodes" bind:this={episodesEl}>
    {#await seasonDetailsReq}
      <Spinner />
    {:then season}
      {#if season?.episodes?.length > 0}
        <ul>
          {#each season.episodes as ep}
            <li>
              {#if ep.still_path}
                <img
                  src={`https://www.themoviedb.org/t/p/w227_and_h127_bestv2/${ep.still_path}`}
                  alt=""
                />
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
          {/each}
        </ul>
      {:else}
        <h3 class="norm">No episodes in this season yet!</h3>
      {/if}
    {:catch err}
      <Error pretty="Failed to load season details!" error={err} />
    {/await}
  </div>
</div>

<style lang="scss">
  .ctr {
    display: flex;
    flex-flow: row;
    gap: 20px;
    width: 100%;
  }

  .episodes {
    overflow: auto;

    ul {
      display: flex;
      flex-flow: column;
      list-style: none;
      gap: 20px;
    }

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
  }

  ul.seasons {
    display: flex;
    flex-flow: column;
    list-style: none;
    gap: 8px;
    min-width: fit-content;
    height: 100vh;
    overflow: auto;
    position: sticky;
    top: 0px;

    button {
      display: flex;
      flex-flow: row;
      gap: 18px;
      align-items: center;
      padding: 10px;
      border: 2px solid #302d2d;
      border-radius: 8px;
      cursor: pointer;
      transition: background-color 100ms ease;

      &:first-of-type {
        margin-top: 10px;
      }

      h1 {
        font-size: 18px;
        font-family: sans-serif;
      }

      h2 {
        font-size: 12px;
        font-family: sans-serif;
        margin-left: auto;
      }

      &:hover,
      &.active {
        color: white;
        background-color: black;
      }

      &.active {
        position: sticky;
        top: 10px;
        bottom: 10px;
      }
    }

    /* hack to get extra scroll space under last el */
    .last {
      padding: 1px;
    }
  }

  @media screen and (max-width: 960px) {
    .ctr {
      flex-flow: column;
    }

    ul.seasons {
      flex-flow: row;
      flex-wrap: wrap;
      position: unset;
      height: unset;
      justify-content: center;

      button {
        &:first-of-type {
          margin-top: unset;
        }
        &.active {
          position: unset;
        }
      }
    }
  }
</style>
