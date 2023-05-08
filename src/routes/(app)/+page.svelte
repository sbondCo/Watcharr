<script lang="ts">
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import { baseURL, removeWatched, updateWatched } from "@/lib/api";
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
        poster={w.content.poster_path ? baseURL + "/img" + w.content.poster_path : undefined}
        title={w.content.title}
        desc={w.content.overview}
        rating={w.rating}
        status={w.status}
        link="/{w.content.type}/{w.content.tmdbId}"
        onStatusChanged={(t) => updateWatched(w.content.tmdbId, w.content.type, t)}
        onRatingChanged={(r) => updateWatched(w.content.tmdbId, w.content.type, undefined, r)}
        onDeleteClicked={() => removeWatched(w.id)}
      />
    {/each}
  {:else}
    You don't have any watched content yet!
  {/if}
</PosterList>
