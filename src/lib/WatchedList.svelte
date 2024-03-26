<script lang="ts">
  import { goto } from "$app/navigation";
  import Icon from "@/lib/Icon.svelte";
  import Poster from "@/lib/poster/Poster.svelte";
  import PosterList from "@/lib/poster/PosterList.svelte";
  import { activeFilters, activeSort, serverFeatures, userSettings } from "@/store";
  import type { Watched, WatchedEpisode, WatchedSeason } from "@/types";
  import GamePoster from "./poster/GamePoster.svelte";
  import { get } from "svelte/store";
  import { seasonAndEpToReadable } from "./util/helpers";

  export let list: Watched[];
  export let isPublicList: boolean = false;

  $: sort = $activeSort;
  $: filters = $activeFilters;
  $: watched = list;
  $: settings = $userSettings;
  $: features = $serverFeatures;

  /**
   * Checks if content has been watched previously
   * by analyzing the watched entrys activity (with
   * the latest AI improvements added in of course.)
   */
  function contentWatchedPreviously(w: Watched) {
    let wp = false;
    const relatedActivity = w.activity.filter(
      (a) =>
        a.type === "ADDED_WATCHED" ||
        a.type === "IMPORTED_ADDED_WATCHED" ||
        a.type === "IMPORTED_WATCHED" ||
        a.type === "STATUS_CHANGED"
    );
    for (let i = 0; i < relatedActivity.length; i++) {
      const ra = relatedActivity[i];
      if (ra.type === "IMPORTED_ADDED_WATCHED") {
        wp = true;
        break;
      } else if (ra.type === "ADDED_WATCHED" || ra.type === "IMPORTED_WATCHED") {
        const data = JSON.parse(ra.data);
        if (data?.status == "FINISHED") {
          wp = true;
          break;
        }
      } else if (ra.type === "STATUS_CHANGED") {
        if (ra.data === "FINISHED") {
          wp = true;
          break;
        }
      }
    }
    return wp;
  }

  // Monsterous code for filters. Soz.
  $: (watched, filters, sort), filt();

  function filt() {
    // Set watched to list and sort it.
    watched = list.sort((a, b) => {
      if (sort[0] === "DATEADDED" && sort[1] === "UP") {
        return Date.parse(a.createdAt) - Date.parse(b.createdAt);
      } else if (sort[0] === "ALPHA") {
        const atitle = a.content ? a.content.title : a.game ? a.game.name : "";
        const btitle = b.content ? b.content.title : b.game ? b.game.name : "";
        if (sort[1] === "UP") {
          return atitle.localeCompare(btitle);
        } else if (sort[1] === "DOWN") {
          return btitle.localeCompare(atitle);
        }
      } else if (sort[0] === "LASTCHANGED") {
        if (sort[1] === "UP") return Date.parse(a.updatedAt) - Date.parse(b.updatedAt);
        else if (sort[1] === "DOWN") return Date.parse(b.updatedAt) - Date.parse(a.updatedAt);
      } else if (sort[0] === "RATING") {
        if (sort[1] === "UP") return (a.rating ?? 0) - (b.rating ?? 0);
        else if (sort[1] === "DOWN") return (b.rating ?? 0) - (a.rating ?? 0);
      }
      // default DATEADDED DOWN
      return Date.parse(b.createdAt) - Date.parse(a.createdAt);
    });
    // If games type filter enabled, but games disabled on server, make sure we remove it from active filters.
    if (!features.games) {
      const af = get(activeFilters);
      af.type = af.type?.filter((a) => a !== "game");
      filters.type = filters.type.filter((f) => f !== "game");
    }
    // Now apply filters to watch list.
    if (filters.status.length > 0 && filters.type.length > 0) {
      // If status and type filters applied, combine both.
      if (settings?.includePreviouslyWatched && filters.status.includes("finished")) {
        watched = watched.filter(
          (w) =>
            (filters.status.includes(w.status?.toLowerCase()) || contentWatchedPreviously(w)) &&
            filters.type.includes(w.content ? w.content.type : w.game ? "game" : "")
        );
      } else {
        watched = watched.filter(
          (w) =>
            filters.status.includes(w.status?.toLowerCase()) &&
            filters.type.includes(w.content ? w.content.type : w.game ? "game" : "")
        );
      }
    } else if (filters.type.length > 0) {
      // Only filter type
      watched = watched.filter((w) =>
        filters.type.includes(w.content ? w.content.type : w.game ? "game" : "")
      );
    } else if (filters.status.length > 0) {
      // Only filter status
      if (settings?.includePreviouslyWatched && filters.status.includes("finished")) {
        watched = watched.filter(
          (w) => filters.status.includes(w.status?.toLowerCase()) || contentWatchedPreviously(w)
        );
      } else {
        watched = watched.filter((w) => filters.status.includes(w.status?.toLowerCase()));
      }
    }
  }

  // Get biggest season watching or biggest season watched.
  // This could probably be simpler but -_-
  function getLatestWatchedInTv(
    ws: WatchedSeason[] | undefined,
    we: WatchedEpisode[] | undefined
  ): string {
    if ((!ws || ws.length <= 0) && (!we || we.length <= 0)) {
      return "";
    }

    let biggestSeasonWatched = -1;
    let biggestSeasonWatching = -1;
    if (ws && ws.length > 0) {
      for (let i = 0; i < ws.length; i++) {
        const s = ws[i];
        if (s.status === "WATCHING") {
          if (s.seasonNumber > biggestSeasonWatching) {
            biggestSeasonWatching = s.seasonNumber;
          }
        } else if (s.status === "FINISHED") {
          if (s.seasonNumber > biggestSeasonWatched) {
            biggestSeasonWatched = s.seasonNumber;
          }
        }
      }
    }
    const season = biggestSeasonWatching >= 0 ? biggestSeasonWatching : biggestSeasonWatched;

    // Look for biggest watched/watching episode in season if any.
    // Does same thing as above.
    let episode: WatchedEpisode | undefined;
    if (we && we.length > 0) {
      let biggestEpisodeWatched: WatchedEpisode | undefined;
      let biggestEpisodeWatching: WatchedEpisode | undefined;
      for (let i = 0; i < we.length; i++) {
        const s = we[i];
        if (season >= 0 && s.seasonNumber !== season) continue;
        if (s.status === "WATCHING") {
          if (!biggestEpisodeWatching) {
            biggestEpisodeWatching = s;
          }
          if (
            s.episodeNumber > biggestEpisodeWatching.episodeNumber ||
            s.seasonNumber > biggestEpisodeWatching.seasonNumber
          ) {
            biggestEpisodeWatching = s;
          }
        } else if (s.status === "FINISHED") {
          if (!biggestEpisodeWatched) {
            biggestEpisodeWatched = s;
          }
          if (
            s.episodeNumber > biggestEpisodeWatched.episodeNumber ||
            s.seasonNumber > biggestEpisodeWatched.seasonNumber
          ) {
            biggestEpisodeWatched = s;
          }
        }
      }
      if (biggestEpisodeWatched || biggestEpisodeWatching) {
        episode =
          biggestEpisodeWatching !== undefined ? biggestEpisodeWatching : biggestEpisodeWatched;
      }
    }

    if (season >= 0 && episode) {
      return seasonAndEpToReadable(season, episode.episodeNumber);
    } else if (season >= 0) {
      return `Season ${season}`;
    } else if (episode) {
      return seasonAndEpToReadable(episode.seasonNumber, episode.episodeNumber);
    } else {
      return "";
    }
  }
</script>

<PosterList>
  {#if watched?.length > 0}
    {#each watched as w (w.id)}
      {#if w.game}
        <GamePoster
          id={w.id}
          rating={w.rating}
          status={w.status}
          media={{
            id: w.game.igdbId,
            coverId: w.game.coverId,
            name: w.game.name,
            summary: w.game.summary,
            firstReleaseDate: w.game.releaseDate,
            poster: w.game.poster
          }}
          disableInteraction={isPublicList}
          extraDetails={{
            dateAdded: w.createdAt,
            dateModified: w.updatedAt
          }}
          fluidSize={true}
        />
      {:else if w.content}
        <Poster
          id={w.id}
          media={{
            id: w.content.tmdbId,
            poster_path: w.content.poster_path,
            title: w.content.title,
            overview: w.content.overview,
            media_type: w.content.type,
            release_date: w.content.release_date,
            first_air_date: w.content.first_air_date
          }}
          rating={w.rating}
          status={w.status}
          disableInteraction={isPublicList}
          extraDetails={{
            dateAdded: w.createdAt,
            dateModified: w.updatedAt,
            lastWatched: getLatestWatchedInTv(w.watchedSeasons, w.watchedEpisodes)
          }}
          fluidSize={true}
        />
      {/if}
    {/each}
  {:else}
    <div class="empty-list">
      <Icon i="reel" wh={80} />
      {#if isPublicList}
        <h2 class="norm">This watched list is empty!</h2>
        <h4 class="norm">Come back later to see if they have added anything.</h4>
      {:else}
        <h2 class="norm">Your watched list is empty!</h2>
        <h4 class="norm">Try searching for something you would like to add.</h4>
        <button on:click={() => goto("/import")}>Import</button>
      {/if}
    </div>
  {/if}
</PosterList>

<style lang="scss">
  .empty-list {
    display: flex;
    flex-flow: column;
    gap: 5px;
    align-items: center;

    h2 {
      margin-top: 10px;
    }

    h4 {
      font-weight: normal;
    }

    button {
      width: max-content;
      padding-left: 20px;
      padding-right: 20px;
      margin-top: 15px;
    }
  }
</style>
