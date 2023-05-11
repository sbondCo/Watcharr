<script lang="ts">
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import { removeWatched, updateWatched } from "@/lib/api";
  import { watchedList } from "@/store";

  $: watched = $watchedList;
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
