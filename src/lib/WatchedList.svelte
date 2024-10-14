<script lang="ts">
  import { goto } from "$app/navigation";
  import Icon from "@/lib/Icon.svelte";
  import Poster from "@/lib/poster/Poster.svelte";
  import PosterList from "@/lib/poster/PosterList.svelte";
  import { activeFilters, activeSort, serverFeatures, userSettings } from "@/store";
  import { WatchedStatus } from "@/types";
  import type { Watched } from "@/types";
  import GamePoster from "./poster/GamePoster.svelte";
  import { get } from "svelte/store";
  import { getLatestWatchedInTv } from "./util/helpers";
  import { notify } from "./util/notify";

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
    try {
      // Set watched to list and sort it.
      watched = list
        .sort((a, b) => {
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
          } else if (sort[0] === "LASTFIN") {
            const aLastFinishActivity = a.activity
              ?.sort(
                (aa, bb) =>
                  Date.parse(bb.customDate ?? bb.updatedAt) -
                  Date.parse(aa.customDate ?? aa.updatedAt)
              )
              ?.find(
                (aa) =>
                  (aa.type === "STATUS_CHANGED" && aa.data === "FINISHED") ||
                  (aa.type === "ADDED_WATCHED" && aa.data?.includes("FINISHED"))
              );
            const bLastFinishActivity = b.activity
              ?.sort(
                (aa, bb) =>
                  Date.parse(bb.customDate ?? bb.updatedAt) -
                  Date.parse(aa.customDate ?? aa.updatedAt)
              )
              ?.find(
                (aa) =>
                  (aa.type === "STATUS_CHANGED" && aa.data === "FINISHED") ||
                  (aa.type === "ADDED_WATCHED" && aa.data?.includes("FINISHED"))
              );
            if (!aLastFinishActivity) return 1;
            if (!bLastFinishActivity) return -1;
            const alfaDate = aLastFinishActivity.customDate ?? aLastFinishActivity.updatedAt;
            const blfaDate = bLastFinishActivity.customDate ?? bLastFinishActivity.updatedAt;
            if (sort[1] === "UP") return Date.parse(alfaDate) - Date.parse(blfaDate);
            else if (sort[1] === "DOWN") return Date.parse(blfaDate) - Date.parse(alfaDate);
          } else if (sort[0] === "RATING") {
            if (sort[1] === "UP") return (a.rating ?? 0) - (b.rating ?? 0);
            else if (sort[1] === "DOWN") return (b.rating ?? 0) - (a.rating ?? 0);
          }
          // default DATEADDED DOWN
          return Date.parse(b.createdAt) - Date.parse(a.createdAt);
        })
        .sort((a, b) => {
          if (a.pinned && !b.pinned) return -1;
          if (!a.pinned && b.pinned) return 1;
          return 0;
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
    } catch (err) {
      console.error("filt: Failed to filter/sort current list!", err);
      notify({ text: "Failed to filter/sort list!", type: "error", time: 6000 });
    }
  }
</script>

{#if watched?.length === 0}
  <div class="central-div">
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
{:else}
  {#each filters.status.length > 0 ? ["ALL"] : Object.values(WatchedStatus) as status}
    {#if filters.status.length === 0}
      <div class="central-div">
        <h2 class="norm first-upper-case">{status}</h2>
      </div>
    {/if}
    <PosterList>
      {#each watched as w (w.id)}
        {#if status.includes(w.status) || filters.status.length > 0}
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
              pinned={w.pinned}
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
              pinned={w.pinned}
            />
          {/if}
        {/if}
      {/each}
    </PosterList>
  {/each}
{/if}

<style lang="scss">
  .central-div {
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

  .first-upper-case {
    text-transform: lowercase;
  }

  .first-upper-case::first-letter {
    text-transform: uppercase;
  }
</style>
