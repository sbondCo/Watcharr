<script lang="ts">
  import type { TMDBSeasonDetails, TMDBShowSeason } from "@/types";
  import axios from "axios";
  import Spinner from "./Spinner.svelte";
  import Error from "./Error.svelte";
  import SeasonsListEpisode from "./SeasonsListEpisode.svelte";

  export let tvId: number;
  export let seasons: TMDBShowSeason[];

  let activeSeason = 1;
  let seasonDetailsReq: Promise<TMDBSeasonDetails>;

  async function sdr(seasonNum: number) {
    return (await axios.get(`/content/tv/${tvId}/season/${seasonNum}`)).data as TMDBSeasonDetails;
  }

  $: {
    seasonDetailsReq = sdr(activeSeason);
  }
</script>

<div class="ctr">
  <ul class="seasons">
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

  <div class="episodes">
    {#await seasonDetailsReq}
      <Spinner />
    {:then season}
      {#if season?.episodes?.length > 0}
        <ul>
          {#each season.episodes as ep}
            <SeasonsListEpisode {ep} />
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
        max-width: 150px;
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
