<script lang="ts">
  import req from "@/lib/api";
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import { watchedList } from "@/store";
  import type { ContentType, Rating, Watched, WatchedAddRequest, WatchedStatus } from "@/types";
  import type { ContentSearch } from "./+page";

  export let data: ContentSearch;

  $: wList = $watchedList;

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
    } as WatchedAddRequest)
      .then((resp) => {
        console.log("Added watched:", resp.data);
        $watchedList.push(resp.data as Watched);
        $watchedList = $watchedList;
      })
      .catch((err) => {
        console.error(err);
      });
  }

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
  {#each data.results as w (w.id)}
    <Poster
      poster={"https://image.tmdb.org/t/p/w500" + w.poster_path}
      title={w.title ?? w.name}
      desc={w.overview}
      onBtnClicked={(t, r) => addWatched(w.id, w.title ? "movie" : "tv", t, r)}
      {...getWatchedDependedProps(w.id, wList)}
    />
  {/each}
</PosterList>
