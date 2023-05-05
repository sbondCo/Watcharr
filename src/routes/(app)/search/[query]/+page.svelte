<script lang="ts">
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import { watchedList } from "@/store";
  import type { MediaType, Watched } from "@/types";
  import type { ContentSearch } from "./+page";
  import { updateWatched } from "@/lib/api";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import axios from "axios";

  export let data;

  $: wList = $watchedList;

  async function search(query: string) {
    return (await axios.get(`/content/${query}`)).data as ContentSearch;
  }

  // Not passing wList from #each loop caused it not to have reactivity.
  // Passing it through must allow it to recognize it as a dependency?
  function getWatchedDependedProps(wid: number, wtype: MediaType, list: Watched[]) {
    const wel = list.find((wl) => wl.content.tmdbId === wid && wl.content.type === wtype);
    if (!wel) return {};
    console.log(wid, wtype, wel?.content.title, wel?.status, wel?.rating);
    return {
      status: wel.status,
      rating: wel.rating
    };
  }
</script>

<svelte:head>
  <title>Content Search</title>
</svelte:head>

{#await search(data.slug)}
  <Spinner />
{:then results}
  <PosterList>
    {#if results?.results?.length > 0}
      {#each results.results as w (w.id)}
        <Poster
          poster={w.poster_path ? "https://image.tmdb.org/t/p/w500" + w.poster_path : undefined}
          title={w.title ?? w.name}
          desc={w.overview}
          link="/{w.media_type}/{w.id}"
          onStatusChanged={(t) => updateWatched(w.id, w.media_type, t)}
          onRatingChanged={(r) => updateWatched(w.id, w.media_type, undefined, r)}
          {...getWatchedDependedProps(w.id, w.media_type, wList)}
        />
      {/each}
    {:else}
      No Search Results!
    {/if}
  </PosterList>
{:catch err}
  <PageError pretty="Failed to load watched list!" error={err} />
{/await}
