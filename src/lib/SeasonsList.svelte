<script lang="ts">
  import type {
    TMDBSeasonDetails,
    TMDBShowSeason,
    Watched,
    WatchedSeasonAddResponse,
    WatchedStatus
  } from "@/types";
  import axios from "axios";
  import Spinner from "./Spinner.svelte";
  import Error from "./Error.svelte";
  import SeasonsListEpisode from "./SeasonsListEpisode.svelte";
  import PosterStatus from "./poster/PosterStatus.svelte";
  import { notify } from "./util/notify";
  import { get } from "svelte/store";
  import { watchedList } from "@/store";
  import PosterRating from "./poster/PosterRating.svelte";

  export let tvId: number;
  export let seasons: TMDBShowSeason[];
  export let watchedItem: Watched; // Watched list item id

  let activeSeason = 1;
  let seasonDetailsReq: Promise<TMDBSeasonDetails>;

  async function sdr(seasonNum: number) {
    return (await axios.get(`/content/tv/${tvId}/season/${seasonNum}`)).data as TMDBSeasonDetails;
  }

  // Add/update watched season
  function updateWatchedSeason(seasonNumber: number, status?: WatchedStatus, rating?: number) {
    const nid = notify({ text: `Saving`, type: "loading" });
    axios
      .post<WatchedSeasonAddResponse>(`/watched/season`, {
        watchedId: watchedItem.id,
        seasonNumber: seasonNumber,
        status,
        rating
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
          wEntry.watchedSeasons = r.data.watchedSeasons;
          if (wEntry.activity?.length > 0) {
            wEntry.activity.push(r.data.addedActivity);
          } else {
            wEntry.activity = [r.data.addedActivity];
          }
          watchedList.update((w) => w);
          notify({ id: nid, text: `Saved!`, type: "success" });
        }
      })
      .catch((err) => {
        console.error(err);
        notify({ id: nid, text: "Failed To Update!", type: "error" });
      });
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
    updateWatchedSeason(seasonNumber, type);
  }

  function handleStarClick(rating: number, seasonNumber: number) {
    updateWatchedSeason(seasonNumber, undefined, rating);
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
        {#if watchedItem}
          {@const ws = watchedItem?.watchedSeasons?.find(
            (s) => s.seasonNumber === season.season_number
          )}
          {#if ws}
            <div class="rating" style={ws?.rating ? "width: 65px" : "width: 45px"}>
              <PosterRating
                rating={ws?.rating}
                btnTooltip="Season Rating"
                handleStarClick={(r) => handleStarClick(r, season.season_number)}
                minimal={true}
                direction="bot"
              />
            </div>
          {/if}
          <div class="status">
            <PosterStatus
              status={ws?.status}
              btnTooltip="Season Status"
              handleStatusClick={(t) => handleStatusClick(t, season.season_number)}
              direction="bot"
              width="100%"
              small
            />
          </div>
        {/if}
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
    transition: top 200ms ease-in-out;

    button {
      display: flex;
      flex-flow: row;
      flex-wrap: wrap;
      gap: 0 18px;
      align-items: center;
      padding: 10px;
      border: 2px solid #302d2d;
      border-radius: 8px;
      cursor: pointer;
      max-width: 220px;
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

  .episodes-topbar {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 20px;
    min-height: 40px;

    div {
      transition: width 100ms ease;

      &:first-of-type {
        margin-left: auto;
      }

      &.rating {
        height: 40px;
        min-height: 40px;
      }

      &.status {
        width: 45px;
        min-height: 40px;
        height: 40px;
        overflow: visible;
      }
    }
  }

  @media screen and (min-width: 960px) {
    :global(body.nav-shown) ul.seasons {
      top: $nav-height;
      height: calc(100vh - $nav-height);
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
