<script lang="ts">
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import { removeWatched, updateWatched } from "@/lib/util/api";
  import { activeFilter, watchedList } from "@/store";

  $: filter = $activeFilter;
  $: watched = $watchedList.sort((a, b) => {
    if (filter[0] === "DATEADDED" && filter[1] === "UP") {
      return Date.parse(a.createdAt) - Date.parse(b.createdAt);
    } else if (filter[0] === "ALPHA") {
      if (filter[1] === "UP") return a.content.title.localeCompare(b.content.title);
      else if (filter[1] === "DOWN") return b.content.title.localeCompare(a.content.title);
    }
    // default DATEADDED DOWN
    return Date.parse(b.createdAt) - Date.parse(a.createdAt);
  });
</script>

<svelte:head>
  <title>Watched List</title>
</svelte:head>

<PosterList>
  {#if watched?.length > 0}
    {#each watched as w (w.id)}
      <Poster
        media={{
          id: w.content.tmdbId,
          poster_path: w.content.poster_path,
          title: w.content.title,
          overview: w.content.overview,
          media_type: w.content.type
        }}
        rating={w.rating}
        status={w.status}
        onStatusChanged={(t) => updateWatched(w.content.tmdbId, w.content.type, t)}
        onRatingChanged={(r) => updateWatched(w.content.tmdbId, w.content.type, undefined, r)}
        onDeleteClicked={() => removeWatched(w.id)}
      />
    {/each}
  {:else}
    You don't have any watched content yet!
  {/if}
</PosterList>
