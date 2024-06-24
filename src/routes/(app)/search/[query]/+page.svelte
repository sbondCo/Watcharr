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
    MoviesSearchResponse,
    PeopleSearchResponse,
    PublicUser,
    ShowsSearchResponse
  } from "@/types";
  import UsersList from "@/lib/UsersList.svelte";
  import { onDestroy, onMount } from "svelte";
  import Error from "@/lib/Error.svelte";
  import GamePoster from "@/lib/poster/GamePoster.svelte";
  import { get } from "svelte/store";
  import { notify } from "@/lib/util/notify.js";
  import Icon from "@/lib/Icon.svelte";
  import { afterNavigate } from "$app/navigation";

  type GameWithMediaType = GameSearch & { media_type: "game" };
  type CombinedResult =
    | ContentSearchMovie
    | ContentSearchTv
    | ContentSearchPerson
    | GameWithMediaType;
  type SearchFilterTypes = MediaType | "game";

  export let data;

  let allSearchResults: CombinedResult[] = [];
  let searchResults: CombinedResult[] = [];
  let activeSearchFilter: SearchFilterTypes | undefined;
  let curPage = 0;
  let maxContentPage = 1;
  let searchRunning = false;
  let contentSearchErr: any;

  const infiniteScrollThreshold = 150;

  $: searchQ = $searchQuery;
  $: wList = $watchedList;

  async function searchMovies(query: string, page: number) {
    try {
      const movies = await axios.get<MoviesSearchResponse>(
        `/content/search/movie/${query}?page=${page}`
      );
      return movies;
    } catch (err) {
      console.error("Movies search failed!", err);
      notify({ text: "Movie Search Failed!", type: "error" });
      contentSearchErr = err;
      throw err;
    }
  }

  async function searchTv(query: string, page: number) {
    try {
      const shows = await axios.get<ShowsSearchResponse>(
        `/content/search/tv/${query}?page=${page}`
      );
      return shows;
    } catch (err) {
      console.error("Tv search failed!", err);
      notify({ text: "Tv Search Failed!", type: "error" });
      contentSearchErr = err;
      throw err;
    }
  }

  async function searchPeople(query: string, page: number) {
    try {
      const people = await axios.get<PeopleSearchResponse>(
        `/content/search/person/${query}?page=${page}`
      );
      return people;
    } catch (err) {
      console.error("People search failed!", err);
      notify({ text: "People Search Failed!", type: "error" });
      contentSearchErr = err;
      throw err;
    }
  }

  async function searchMulti(query: string, page: number) {
    try {
      return await axios.get<ContentSearch>(`/content/search/multi/${query}?page=${page}`);
    } catch (err) {
      console.error("Movies/Tv search failed!", err);
      notify({ text: "Movie/Tv Search Failed!", type: "error" });
      contentSearchErr = err;
      throw err;
    }
  }

  async function searchGames(query: string, page: number) {
    try {
      const f = get(serverFeatures);
      if (!f.games) {
        console.debug("game search is not enabled on this server");
        return { data: [] };
      }
      const games = await axios.get<GameSearch[]>(`/game/search/${query}`);
      return {
        data: games.data.map((g) => ({
          ...g,
          media_type: "game"
        })) as GameWithMediaType[]
      };
    } catch (err) {
      console.error("Game search failed!", err);
      notify({ text: "Game Search Failed!", type: "error" });
      // This could cause the multi error to be replaced with this one
      // if both requests fail, but it shouldn't matter, retry button
      // should still do it's job if needed.
      contentSearchErr = err;
      throw err;
    }
  }

  async function search(query: string) {
    if (searchRunning) {
      console.debug("search: already running");
      return;
    }
    if (curPage === maxContentPage) {
      console.debug("search: max page reached");
      return;
    }
    searchRunning = true;
    try {
      if (activeSearchFilter) {
        // If we have a search filter selected, search for just one specific type of content.
        console.log("Search: A filter is active:", activeSearchFilter);
        let cdata;
        if (activeSearchFilter === "movie") {
          cdata = (await searchMovies(query, curPage + 1)).data;
          allSearchResults.push(...cdata.results);
        } else if (activeSearchFilter === "tv") {
          cdata = (await searchTv(query, curPage + 1)).data;
          allSearchResults.push(...cdata.results);
        } else if (activeSearchFilter === "person") {
          cdata = (await searchPeople(query, curPage + 1)).data;
          // HACK couldn't be bothered to fix this type error
          allSearchResults.push(...(cdata.results as unknown as CombinedResult[]));
        } else if (activeSearchFilter === "game") {
          allSearchResults.push(...(await searchGames(query, curPage + 1)).data);
        } else {
          console.error("Active search filter is invalid:", activeSearchFilter);
        }
        maxContentPage = cdata ? cdata.total_pages ?? 1 : 1;
      } else {
        // If no search filter is applied, do default multi+game combined search.
        console.log("Search: No filter is applied.");
        const r = await Promise.allSettled([
          searchMulti(query, curPage + 1),
          searchGames(query, curPage + 1)
        ]);
        if (r[0].status == "fulfilled") {
          if (r[0].value.data.total_pages) {
            maxContentPage = r[0].value.data.total_pages;
          }
          allSearchResults.push(...r[0].value.data.results);
        }
        if (r[1].status == "fulfilled") {
          allSearchResults.push(...r[1].value.data);
        }
      }
      console.debug("allSearchResults:", allSearchResults);
      searchResults = allSearchResults;
      curPage++;
      searchRunning = false;
      // If results don't fill the page enough to enable scrolling,
      // the user could be stuck and not be able to get more results
      // to show, run `infiniteScroll` to load more if we can.
      // Smol timeout to give ui time to render so end of page calc
      // can be accurate.
      setTimeout(() => {
        infiniteScroll();
      }, 250);
    } catch (err) {
      console.error("search failed!", err);
      contentSearchErr = err;
      searchRunning = false;
    }
  }

  async function doCleanSearch() {
    if (!data.slug) {
      console.error("doCleanSearch: No query to use.");
      return;
    }
    curPage = 0;
    allSearchResults = [];
    searchResults = [];
    search(data.slug);
  }

  async function searchUsers(query: string) {
    return (await axios.get(`/user/search/${query}`)).data as PublicUser[];
  }

  function setActiveSearchFilter(to: SearchFilterTypes) {
    if (activeSearchFilter === to) {
      activeSearchFilter = undefined;
    } else {
      activeSearchFilter = to;
    }
    doCleanSearch();
  }

  async function infiniteScroll() {
    // If an error is being shown, no more infinite scroll.
    if (contentSearchErr) {
      return;
    }
    if (
      window.innerHeight + Math.round(window.scrollY) + infiniteScrollThreshold >=
      document.body.offsetHeight
    ) {
      console.log("reached end");
      window.removeEventListener("scroll", infiniteScroll);
      if (data.slug) await search(data.slug);
      window.addEventListener("scroll", infiniteScroll);
    }
  }

  onMount(() => {
    if (!searchQ && data.slug) {
      searchQuery.set(data.slug);
    }
    doCleanSearch();

    window.addEventListener("scroll", infiniteScroll);
    window.addEventListener("resize", infiniteScroll);

    return () => {
      window.removeEventListener("scroll", infiniteScroll);
      window.removeEventListener("resize", infiniteScroll);
    };
  });

  afterNavigate(() => {
    console.log("Query changed, performing new search");
    if (!searchQ && data.slug) {
      searchQuery.set(data.slug);
    }
    doCleanSearch();
  });

  onDestroy(() => {
    searchQuery.set("");
  });
</script>

<svelte:head>
  <title>Search Results{data?.slug ? ` for '${data?.slug}'` : ""}</title>
</svelte:head>

<span style="position: sticky;top: 70px;">{curPage} / {maxContentPage}</span>

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

      <div class={`results-filters-header${searchRunning ? " search-running" : ""}`}>
        <h2>Results</h2>
        <div>
          <button
            class="plain"
            data-active={activeSearchFilter === "movie"}
            on:click={() => setActiveSearchFilter("movie")}
          >
            <Icon i="film" wh={20} /> Movies
          </button>
          <button
            class="plain"
            data-active={activeSearchFilter === "tv"}
            on:click={() => setActiveSearchFilter("tv")}
          >
            <Icon i="tv" wh={20} /> TV Shows
          </button>
          <button
            class="plain"
            data-active={activeSearchFilter === "game"}
            on:click={() => setActiveSearchFilter("game")}
          >
            <Icon i="gamepad" wh={20} /> Games
          </button>
          <button
            class="plain"
            data-active={activeSearchFilter === "person"}
            on:click={() => setActiveSearchFilter("person")}
          >
            <Icon i="people-nocircle" wh={20} /> People
          </button>
        </div>
      </div>
      <PosterList>
        {#if searchResults?.length > 0}
          {#each searchResults as w, i (`${i}-${w.media_type}-${w.id}`)}
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
              <Poster media={w} {...getWatchedDependedProps(w.id, w.media_type, wList)} fluidSize />
            {/if}
          {/each}
        {:else if !searchRunning && !contentSearchErr}
          <!-- If search is running or we have an error, no point in showing 'no results' message. -->
          No Search Results!
        {/if}
      </PosterList>

      {#if searchRunning}
        <Spinner />
      {/if}

      {#if contentSearchErr}
        <div style="margin-bottom: 60px;">
          <Error
            pretty="Failed to load results!"
            error={contentSearchErr}
            onRetry={() => {
              contentSearchErr = undefined;
              search(data.slug);
            }}
          />
        </div>
      {/if}
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

    &.search-running {
      button {
        opacity: 0.8;
        pointer-events: none;
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
