<script lang="ts">
  import type { TMDBSeasonDetails, TMDBShowSeason, Watched, WatchedStatus } from "@/types";
  import axios from "axios";
  import Spinner from "./Spinner.svelte";
  import Error from "./Error.svelte";
  import SeasonsListEpisode from "./SeasonsListEpisode.svelte";
  import PosterStatus from "./poster/PosterStatus.svelte";
  import { notify } from "./util/notify";
  import { get } from "svelte/store";
  import { watchedList } from "@/store";

  export let tvId: number;
  export let seasons: TMDBShowSeason[];
  export let watchedItem: Watched; // Watched list item id

  let activeSeason = 1;
  let seasonDetailsReq: Promise<TMDBSeasonDetails>;

  async function sdr(seasonNum: number) {
    return (await axios.get(`/content/tv/${tvId}/season/${seasonNum}`)).data as TMDBSeasonDetails;
  }

  function handleStatusClick(type: WatchedStatus | "DELETE", seasonNumber: number) {
    if (type === "DELETE") {
      const ws = watchedItem.watchedSeasons?.find((s) => s.seasonNumber === seasonNumber);
      if (!ws) {
        notify({ text: "Failed to find watched season id. Please try refreshing.", type: "error" });
        return;
      }
      const nid = notify({ text: `Saving`, type: "loading" });
      axios
        .delete(`/watched/season/${ws.id}`)
        .then((r) => {
          const wList = get(watchedList);
          const wEntry = wList.find((w) => w.id === watchedItem.id);
          if (!wEntry) {
            notify({
              id: nid,
              text: `Request succeeded, but failed to find local data. Please refresh.`,
              type: "error"
            });
            return;
          }
          if (r.status === 200) {
            wEntry.watchedSeasons = wEntry.watchedSeasons?.filter((s) => s.id !== ws.id);
            watchedList.update((w) => w);
            notify({ id: nid, text: `Removed!`, type: "success" });
          }
        })
        .catch((err) => {
          console.error(err);
          notify({ id: nid, text: "Failed To Remove!", type: "error" });
        });
      return;
    }
    const nid = notify({ text: `Saving`, type: "loading" });
    axios
      .post(`/watched/season`, {
        watchedId: watchedItem.id,
        seasonNumber: seasonNumber,
        status: type
      })
      .then((r) => {
        const wList = get(watchedList);
        const wEntry = wList.find((w) => w.id === watchedItem.id);
        if (!wEntry) {
          notify({
            id: nid,
            text: `Request succeeded, but failed to find local data. Please refresh.`,
            type: "error"
          });
          return;
        }
        if (r.status === 200) {
          wEntry.watchedSeasons = r.data;
          watchedList.update((w) => w);
          notify({ id: nid, text: `Saved!`, type: "success" });
        }
      })
      .catch((err) => {
        console.error(err);
        notify({ id: nid, text: "Failed To Update!", type: "error" });
      });
  }

  $: {
    seasonDetailsReq = sdr(activeSeason);
  }
</script>

<div class="ctr">
  <ul class="seasons">
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
      <div class="episodes-topbar">
        <h3>{season.name}</h3>
        <div>
          <PosterStatus
            status={watchedItem?.watchedSeasons?.find(
              (s) => s.seasonNumber === season.season_number
            )?.status}
            btnTooltip="Season Status"
            handleStatusClick={(t) => handleStatusClick(t, season.season_number)}
            direction="bot"
            width="100%"
            small
          />
        </div>
      </div>
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
    width: 100%;

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

  .episodes-topbar {
    display: flex;
    align-items: center;
    margin-bottom: 10px;

    div {
      width: 45px;
      min-height: 40px;
      height: 40px;
      overflow: visible;
      margin-left: auto;
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
