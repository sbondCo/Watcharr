<script lang="ts">
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import { watchedList } from "@/store";
  import type { ContentSearch } from "./+page";
  import { removeWatched, updateWatched } from "@/lib/api";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import axios from "axios";
  import { getWatchedDependedProps } from "@/lib/helpers";
  import PersonPoster from "@/lib/PersonPoster.svelte";

  export let data;

  $: wList = $watchedList;

  async function search(query: string) {
    return (await axios.get(`/content/${query}`)).data as ContentSearch;
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
        {#if w.media_type === "person"}
          <PersonPoster id={w.id} name={w.name} path={w.profile_path} />
        {:else}
          <Poster
            media={w}
            onStatusChanged={(t) => updateWatched(w.id, w.media_type, t)}
            onRatingChanged={(r) => updateWatched(w.id, w.media_type, undefined, r)}
            onDeleteClicked={() => {
              const wl = wList.find((wi) => wi.content.tmdbId === w.id);
              if (!wl) {
                console.error("Failed to find item in watched list, cant remove!");
                return;
              }
              removeWatched(wl.id);
            }}
            {...getWatchedDependedProps(w.id, w.media_type, wList)}
          />
        {/if}
      {/each}
    {:else}
      No Search Results!
    {/if}
  </PosterList>
{:catch err}
  <PageError pretty="Failed to load watched list!" error={err} />
{/await}
