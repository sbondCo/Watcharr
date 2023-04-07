<script lang="ts">
  import req from "@/lib/api";
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import type { ContentType, Rating, WatchedAddRequest, WatchedStatus } from "@/types";
  import type { ContentSearch } from "./+page";

  export let data: ContentSearch;

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
</script>

<PosterList>
  {#each data.results as w}
    <Poster
      poster={"https://image.tmdb.org/t/p/w500" + w.poster_path}
      title={w.title ?? w.name}
      desc={w.overview}
      onBtnClicked={(t, r) => addWatched(w.id, w.title ? "movie" : "tv", t, r)}
    />
  {/each}
</PosterList>
