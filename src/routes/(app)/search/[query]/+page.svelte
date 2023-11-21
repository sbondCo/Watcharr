<script lang="ts">
  import Poster from "@/lib/poster/Poster.svelte";
  import PosterList from "@/lib/poster/PosterList.svelte";
  import { searchQuery, watchedList } from "@/store";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import axios from "axios";
  import { getWatchedDependedProps } from "@/lib/util/helpers";
  import PersonPoster from "@/lib/poster/PersonPoster.svelte";
  import type { ContentSearch, PublicUser } from "@/types";
  import UsersList from "@/lib/UsersList.svelte";
  import { onDestroy, onMount } from "svelte";

  export let data;

  $: searchQ = $searchQuery;
  $: wList = $watchedList;

  async function search(query: string) {
    return (await axios.get(`/content/${query}`)).data as ContentSearch;
  }

  async function searchUsers(query: string) {
    return (await axios.get(`/user/search/${query}`)).data as PublicUser[];
  }

  onMount(() => {
    if (!searchQ && data.slug) {
      searchQuery.set(data.slug);
    }
  });

  onDestroy(() => {
    searchQuery.set("");
  });
</script>

<svelte:head>
  <title>Content Search</title>
</svelte:head>

{#if data.slug}
  {#await searchUsers(data.slug) then results}
    {#if results?.length > 0}
      <UsersList users={results} />
    {/if}
  {:catch err}
    <PageError pretty="Failed to load watched list!" error={err} />
  {/await}

  {#await search(data.slug)}
    <Spinner />
  {:then results}
    <h2 style="margin-left: 30px;">Results</h2>
    <PosterList>
      {#if results?.results?.length > 0}
        {#each results.results as w (w.id)}
          {#if w.media_type === "person"}
            <PersonPoster id={w.id} name={w.name} path={w.profile_path} />
          {:else}
            <Poster media={w} {...getWatchedDependedProps(w.id, w.media_type, wList)} />
          {/if}
        {/each}
      {:else}
        No Search Results!
      {/if}
    </PosterList>
  {:catch err}
    <PageError pretty="Failed to load watched list!" error={err} />
  {/await}
{:else}
  <h2>No Search Query!</h2>
{/if}
