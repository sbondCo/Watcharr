<script lang="ts">
  import type { TMDBSeasonDetails, TMDBShowSeason } from "@/types";
  import axios from "axios";
  import Spinner from "./Spinner.svelte";
  import PageError from "./PageError.svelte";
  import Error from "./Error.svelte";
  import { onMount } from "svelte";

  export let tvId: number;
  export let seasons: TMDBShowSeason[];

  let activeSeason = seasons[0].season_number;
  let seasonDetailsReq = sdr(activeSeason);
  let seasonsEl: HTMLUListElement, episodesEl: HTMLDivElement;

  async function sdr(seasonNum: number) {
    return (await axios.get(`/content/tv/${tvId}/season/${seasonNum}`)).data as TMDBSeasonDetails;
  }

  $: {
    console.log("as change");
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
        <h2>{new Date(Date.parse(season.air_date)).getFullYear()}</h2>
      </button>
    {/each}
  </ul>

  <div class="episodes" bind:this={episodesEl}>
    {#await seasonDetailsReq}
      <Spinner />
    {:then season}
      {#if season}
        <ul>
          {#each season.episodes as ep}
            <li>
              <img
                src={`https://www.themoviedb.org/t/p/w227_and_h127_bestv2/${ep.still_path}`}
                alt=""
              />
              <span>
                <b>{ep.episode_number}</b>
                {ep.name}
              </span>
            </li>
          {/each}
        </ul>
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
    width: 100%;
  }

  .episodes {
    padding: 0px 20px;
    overflow: auto;

    ul {
      display: flex;
      flex-flow: column;
      list-style: none;
      gap: 8px;
    }

    li {
      display: flex;
      flex-flow: row;
      gap: 12px;

      img {
        width: 227px;
        height: 127px;
        border-radius: 10px;
      }

      span {
        padding: 8px 5px;
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
    padding: 10px 0;

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

      h1 {
        font-size: 18px;
      }

      h2 {
        font-size: 12px;
        font-family: sans-serif;
      }

      &:hover,
      &.active {
        color: white;
        background-color: black;
      }
    }
  }
</style>
