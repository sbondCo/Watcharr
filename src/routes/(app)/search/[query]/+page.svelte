<script lang="ts">
  import Poster from "@/lib/poster/Poster.svelte";
  import PosterList from "@/lib/poster/PosterList.svelte";
  import { searchQuery, watchedList } from "@/store";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import axios from "axios";
  import { getWatchedDependedProps, getPlayedDependedProps } from "@/lib/util/helpers";
  import PersonPoster from "@/lib/poster/PersonPoster.svelte";
  import type {
    ContentSearch,
    ContentSearchMovie,
    ContentSearchPerson,
    ContentSearchTv,
    GameSearch,
    MediaType,
    PublicUser
  } from "@/types";
  import UsersList from "@/lib/UsersList.svelte";
  import { onDestroy, onMount } from "svelte";
  import Error from "@/lib/Error.svelte";
  import GamePoster from "@/lib/poster/GamePoster.svelte";

  export let data;

  $: searchQ = $searchQuery;
  $: wList = $watchedList;

  interface CombinedSearchResults {
    id: number;
    name: string;
    type: MediaType | "game";
    og: (ContentSearchMovie | ContentSearchTv | ContentSearchPerson | GameSearch)[];
  }

  type GameWithMediaType = GameSearch & { media_type: "game" };
  type CombinedResult =
    | ContentSearchMovie
    | ContentSearchTv
    | ContentSearchPerson
    | GameWithMediaType;

  // async function search(query: string) {
  //   return (await axios.get(`/content/${query}`)).data as ContentSearch;
  // }

  // async function searchGames(query: string) {
  //   return (await axios.get(`/game/${query}`)).data as GameSearch[];
  // }

  // TODO only search games as well if that server feature is enabled.
  async function search(query: string) {
    const r = await Promise.all([
      axios.get<ContentSearch>(`/content/${query}`),
      axios.get<GameSearch[]>(`/game/search/${query}`)
    ]);
    const games: GameWithMediaType[] = r[1].data.map((g) => ({
      ...g,
      media_type: "game"
    }));
    const d = new Array<CombinedResult>().concat
      .apply([], [r[0].data.results, games])
      ?.sort((a, b) => {
        let name = "";
        if (a.media_type === "game" || a.media_type === "tv" || a.media_type === "person") {
          name = a.name ?? "";
        } else if (a.media_type === "movie") {
          name = a.title ?? "";
        }

        let name2 = "";
        if (b.media_type === "game" || b.media_type === "tv" || b.media_type === "person") {
          name2 = b.name ?? "";
        } else if (b.media_type === "movie") {
          name2 = b.title ?? "";
        }

        if (name < name2) {
          return 1;
        }
        if (name > name2) {
          return -1;
        }
        return 0;
      });
    return d;
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

<div class="content">
  <div class="inner">
    {#if data.slug}
      {#await searchUsers(data.slug) then results}
        {#if results?.length > 0}
          <UsersList users={results} />
        {/if}
      {:catch err}
        <PageError pretty="Failed to load users!" error={err} />
      {/await}

      {#await search(data.slug)}
        <Spinner />
      {:then results}
        <h2>Results</h2>
        <PosterList>
          {#if results?.length > 0}
            {#each results as w (w.id)}
              {#if w.media_type === "person"}
                <PersonPoster id={w.id} name={w.name} path={w.profile_path} />
              {:else if w.media_type === "game"}
                <GamePoster
                  media={{
                    id: w.id,
                    coverId: w.cover.image_id,
                    name: w.name,
                    summary: w.summary,
                    firstReleaseDate: w.first_release_date
                  }}
                  {...getPlayedDependedProps(w.id, wList)}
                />
              {:else}
                <Poster media={w} {...getWatchedDependedProps(w.id, w.media_type, wList)} />
              {/if}
            {/each}
          {:else}
            No Search Results!
          {/if}
        </PosterList>
      {:catch err}
        <Error pretty="Failed to load results!" error={err} />
      {/await}

      <!-- {#await searchGames(data.slug)}
        <Spinner />
      {:then results}
        <h2>Results</h2>
        <PosterList>
          {#if results?.length > 0}
            {#each results as w (w.id)}
              <GamePoster media={w} />
            {/each}
          {:else}
            No Search Results!
          {/if}
        </PosterList>
      {:catch err}
        <Error pretty="Failed to load results!" error={err} />
      {/await} -->

      <!-- {#await search(data.slug)}
        <Spinner />
      {:then results}
        <h2>Results</h2>
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
        <Error pretty="Failed to load results!" error={err} />
      {/await} -->
    {:else}
      <h2>No Search Query!</h2>
    {/if}
  </div>
</div>

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;

    .inner {
      width: 100%;
      max-width: 1200px;

      h2 {
        margin-left: 15px;
      }
    }
  }
</style>
