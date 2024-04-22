<script lang="ts">
  import Poster from "@/lib/poster/Poster.svelte";
  import PosterList from "@/lib/poster/PosterList.svelte";
  import { searchQuery, serverFeatures, watchedList } from "@/store";
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
  import { get } from "svelte/store";
  import { notify } from "@/lib/util/notify.js";
  import Icon from "@/lib/Icon.svelte";
  import { onNavigate } from "$app/navigation";

  type GameWithMediaType = GameSearch & { media_type: "game" };
  type CombinedResult =
    | ContentSearchMovie
    | ContentSearchTv
    | ContentSearchPerson
    | GameWithMediaType;
  type SearchFilterTypes = MediaType | "game";

  export let data;

  let allSearchResults: CombinedResult[];
  let searchResults: CombinedResult[];
  let activeSearchFilter: SearchFilterTypes | undefined;

  $: searchQ = $searchQuery;
  $: wList = $watchedList;
  $: (allSearchResults, activeSearchFilter), filterResults();

  async function search(query: string) {
    const f = get(serverFeatures);
    if (!f.games) {
      console.log("Search: Only for movies/tv");
      allSearchResults = (await axios.get<ContentSearch>(`/content/${query}`)).data.results;
      searchResults = allSearchResults;
      return;
    }
    console.log("Search: For movies/tv and games");
    // To get around promise.all rejecting both promises when one fails,
    // catch them separately and return empty object so we can still
    // display the other media types.
    const r = await Promise.all([
      axios.get<ContentSearch>(`/content/${query}`).catch((err) => {
        console.error("Movies/Tv search failed!", err);
        notify({ text: "Movie/Tv Search Failed!", type: "error" });
        return { data: { results: [] } };
      }),
      axios.get<GameSearch[]>(`/game/search/${query}`).catch((err) => {
        console.error("Game search failed!", err);
        notify({ text: "Game Search Failed!", type: "error" });
        return { data: [] };
      })
    ]);
    const games: GameWithMediaType[] = r[1].data.map((g) => ({
      ...g,
      media_type: "game"
    }));
    allSearchResults = new Array<CombinedResult>().concat.apply([], [r[0].data.results, games]);
    searchResults = allSearchResults;
  }

  async function searchUsers(query: string) {
    return (await axios.get(`/user/search/${query}`)).data as PublicUser[];
  }

  function setActiveSearchFilter(to: SearchFilterTypes) {
    if (activeSearchFilter === to) {
      activeSearchFilter = undefined;
      return;
    }
    activeSearchFilter = to;
  }

  function filterResults() {
    if (!activeSearchFilter) {
      searchResults = allSearchResults;
      return;
    }
    searchResults = allSearchResults.filter((s) => s.media_type === activeSearchFilter);
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
  <title>Search Results{data?.slug ? ` for '${data?.slug}'` : ""}</title>
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
      {:then}
        <div class="results-filters-header">
          <h2>Results</h2>
          <div>
            {#if allSearchResults?.length > 0}
              {#if allSearchResults.find((s) => s.media_type === "movie") || activeSearchFilter === "movie"}
                <button
                  class="plain"
                  data-active={activeSearchFilter === "movie"}
                  on:click={() => setActiveSearchFilter("movie")}
                >
                  <Icon i="film" wh={20} /> Movies
                </button>
              {/if}
              {#if allSearchResults.find((s) => s.media_type === "tv") || activeSearchFilter === "tv"}
                <button
                  class="plain"
                  data-active={activeSearchFilter === "tv"}
                  on:click={() => setActiveSearchFilter("tv")}
                >
                  <Icon i="tv" wh={20} /> TV Shows
                </button>
              {/if}
              {#if allSearchResults.find((s) => s.media_type === "game") || activeSearchFilter === "game"}
                <button
                  class="plain"
                  data-active={activeSearchFilter === "game"}
                  on:click={() => setActiveSearchFilter("game")}
                >
                  <Icon i="gamepad" wh={20} /> Games
                </button>
              {/if}
              {#if allSearchResults.find((s) => s.media_type === "person") || activeSearchFilter === "person"}
                <button
                  class="plain"
                  data-active={activeSearchFilter === "person"}
                  on:click={() => setActiveSearchFilter("person")}
                >
                  <Icon i="people-nocircle" wh={20} /> People
                </button>
              {/if}
            {/if}
          </div>
        </div>
        <PosterList>
          {#if searchResults?.length > 0}
            {#each searchResults as w (w.id)}
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
                  fluidSize
                />
              {:else}
                <Poster
                  media={w}
                  {...getWatchedDependedProps(w.id, w.media_type, wList)}
                  fluidSize
                />
              {/if}
            {/each}
          {:else}
            No Search Results!
          {/if}
        </PosterList>
      {:catch err}
        <Error pretty="Failed to load results!" error={err} />
      {/await}
    {:else}
      <h2>No Search Query!</h2>
    {/if}
  </div>
</div>

<style lang="scss">
  .results-filters-header {
    display: flex;
    flex-flow: row;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px;

    div {
      display: flex;
      flex-flow: row;
      flex-wrap: wrap;
      gap: 10px;
      margin: 0 15px;

      button {
        display: flex;
        flex-flow: row;
        flex-wrap: wrap;
        gap: 8px;
        align-items: center;
        height: fit-content;
        padding: 8px 12px;
        border-radius: 8px;
        font-size: 14px;
        color: $text-color;
        fill: $text-color;
        transition:
          background-color 150ms ease,
          color 150ms ease,
          outline 150ms ease;

        &:hover,
        &[data-active="true"] {
          color: $bg-color;
          fill: $bg-color;
          background-color: $accent-color-hover;
        }

        &[data-active="true"] {
          outline: 3px solid $accent-color;
        }

        @media screen and (max-width: 500px) {
          flex-flow: column;
        }
      }

      @media screen and (max-width: 500px) {
        width: 100%;
        justify-content: center;
      }
    }
  }

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
