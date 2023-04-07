<script lang="ts">
  import req from "@/lib/api";
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import { watchedList } from "@/store";
  import type { Rating, Watched, WatchedStatus, WatchedUpdateRequest } from "@/types";

  let watched: Watched[];

  watchedList.subscribe((wl) => (watched = wl));

  function updateWatched(id: number, status?: WatchedStatus, rating?: Rating) {
    if (!status && !rating) return;
    let obj = {} as WatchedUpdateRequest;
    if (status) obj.status = status;
    if (rating) obj.rating = rating;
    req(`/watched/${id}`, "PUT", obj);
  }
</script>

<svelte:head>
  <title>Watched List</title>
</svelte:head>

<PosterList>
  {#if watched && watched.length > 0}
    {#each watched as w (w.id)}
      <Poster
        poster={"http://localhost:3080/img" + w.content.poster_path}
        title={w.content.title}
        desc={w.content.overview}
        rating={w.rating}
        status={w.status}
        onBtnClicked={(type) => {
          updateWatched(w.id, type);
          w.status = type;
          $watchedList = watched;
        }}
        onRatingChanged={(rating) => {
          updateWatched(w.id, undefined, rating);
          w.rating = rating;
          $watchedList = watched;
        }}
      />
    {/each}
  {:else}
    You don't have any watched content yet!
  {/if}
</PosterList>
