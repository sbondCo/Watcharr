<script lang="ts">
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import { watchedList } from "@/store";
  import type { ContentSearch } from "./+page";
  import { removeWatched, updateWatched } from "@/lib/util/api";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import axios from "axios";
  import { getWatchedDependedProps } from "@/lib/util/helpers";
  import PersonPoster from "@/lib/PersonPoster.svelte";
  import type { PublicUser } from "@/types";
  import UsersList from "@/lib/UsersList.svelte";

  export let data;

  $: wList = $watchedList;

  async function search(query: string) {
    return (await axios.get(`/content/${query}`)).data as ContentSearch;
  }

  async function searchUsers(query: string) {
    return (await axios.get(`/user/search/${query}`)).data as PublicUser[];
  }
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
