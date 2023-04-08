<script lang="ts">
  import req from "@/lib/api";
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import { watchedList } from "@/store";
  import type { ContentType, Rating, WatchedAddRequest, WatchedStatus } from "@/types";
  import { get } from "svelte/store";
  import type { ContentSearch } from "./+page";

  export let data: ContentSearch;

  let wList = get(watchedList);

  function addWatched(
    contentId: number,
    contentType: ContentType,
    status: WatchedStatus,
    rating: Rating
  ) {
    req("/watched", "POST", {
      contentId,
      contentType,
      rating,
      status
    } as WatchedAddRequest);
  }

  function getWatchedDependedProps(wid: number) {
    const wel = wList.find((wl) => wl.content.id === wid);
    if (!wel) return {};
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
  {#each data.results as w (w.id)}
    <Poster
      poster={"https://image.tmdb.org/t/p/w500" + w.poster_path}
      title={w.title ?? w.name}
      desc={w.overview}
      onBtnClicked={(t, r) => addWatched(w.id, w.title ? "movie" : "tv", t, r)}
      {...getWatchedDependedProps(w.id)}
    />
  {/each}
</PosterList>
