<script lang="ts">
  import Icon from "@/lib/Icon.svelte";
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import { activeFilter, watchedList } from "@/store";

  $: filter = $activeFilter;
  $: watched = $watchedList.sort((a, b) => {
    if (filter[0] === "DATEADDED" && filter[1] === "UP") {
      return Date.parse(a.createdAt) - Date.parse(b.createdAt);
    } else if (filter[0] === "ALPHA") {
      if (filter[1] === "UP") return a.content.title.localeCompare(b.content.title);
      else if (filter[1] === "DOWN") return b.content.title.localeCompare(a.content.title);
    } else if (filter[0] === "LASTCHANGED") {
      if (filter[1] === "UP") return Date.parse(a.updatedAt) - Date.parse(b.updatedAt);
      else if (filter[1] === "DOWN") return Date.parse(b.updatedAt) - Date.parse(a.updatedAt);
    } else if (filter[0] === "RATING") {
      if (filter[1] === "UP") return (a.rating ?? 0) - (b.rating ?? 0);
      else if (filter[1] === "DOWN") return (b.rating ?? 0) - (a.rating ?? 0);
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
      />
    {/each}
  {:else}
    <div class="empty-list">
      <Icon i="reel" wh={80} />
      <h2 class="norm">Your watched list is empty!</h2>
      <h4 class="norm">Try searching for something you would like to add.</h4>
    </div>
  {/if}
</PosterList>

<style lang="scss">
  .empty-list {
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
  }
</style>
