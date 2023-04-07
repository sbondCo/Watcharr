<script lang="ts">
  import req from "@/lib/api";
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import type { Rating, WatchedStatus, WatchedUpdateRequest } from "@/types";

  export let data: import("./$types").PageData;

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
  {#if data?.watched && data.watched.length > 0}
    {#each data.watched as w}
      <Poster
        poster={"http://localhost:3080/img" + w.content.poster_path}
        title={w.content.title}
        desc={w.content.overview}
        rating={w.rating}
        status={w.status}
        onBtnClicked={(type) => {
          updateWatched(w.id, type);
        }}
        onRatingChanged={(rating) => {
          updateWatched(w.id, undefined, rating);
        }}
      />
    {/each}
  {:else}
    You don't have any watched content yet!
  {/if}
</PosterList>
