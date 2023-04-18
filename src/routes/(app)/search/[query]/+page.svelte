<script lang="ts">
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import { watchedList } from "@/store";
  import type { ContentType, Watched, WatchedAddRequest, WatchedStatus } from "@/types";
  import type { ContentSearch } from "./+page";
  import axios from "axios";
  import { updateWatched } from "@/lib/api";

  export let data: ContentSearch;

  $: wList = $watchedList;

  // Not passing wList from #each loop caused it not to have reactivity.
  // Passing it through must allow it to recognize it as a dependency?
  function getWatchedDependedProps(wid: number, list: Watched[]) {
    const wel = list.find((wl) => wl.content.id === wid);
    if (!wel) return {};
    console.log(wel.content.title, wel.status, wel.rating);
    return {
      status: wel.status,
      rating: wel.rating
    };
  }
</script>

<svelte:head>
  <title>Content Search</title>
</svelte:head>

<PosterList>
  {#if data?.results?.length > 0}
    {#each data.results as w (w.id)}
      <Poster
        poster={w.poster_path ? "https://image.tmdb.org/t/p/w500" + w.poster_path : undefined}
        title={w.title ?? w.name}
        desc={w.overview}
        onBtnClicked={(t, r) => updateWatched(w.id, w.title ? "movie" : "tv", t, r)}
        onRatingChanged={(r) => updateWatched(w.id, w.title ? "movie" : "tv", undefined, r)}
        {...getWatchedDependedProps(w.id, wList)}
      />
    {/each}
  {:else}
    No Search Results!
  {/if}
</PosterList>
