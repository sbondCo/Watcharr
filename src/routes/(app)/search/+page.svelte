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
  import Icon from "@/lib/Icon.svelte";
  import { afterNavigate, goto } from "$app/navigation";
  import { page } from "$app/stores";

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
  let reqController = new AbortController();

  $: query = data?.query;
  $: searchQ = $searchQuery;
  $: wList = $watchedList;

  async function searchMovies(query: string, page: number) {
    try {
      const movies = await axios.get<MoviesSearchResponse>(`/content/search/movie`, {
        params: {
          q: encodeURIComponent(query),
          page
        },
        signal: reqController.signal
      });
      return movies;
    } catch (err) {
      console.error("Movies search failed!", err);
      throw err;
    }
  }

  async function searchTv(query: string, page: number) {
    try {
      const shows = await axios.get<ShowsSearchResponse>(`/content/search/tv`, {
        params: {
          q: encodeURIComponent(query),
          page
        },
        signal: reqController.signal
      });
      return shows;
    } catch (err) {
      console.error("Tv search failed!", err);
      throw err;
    }
  }

  async function searchPeople(query: string, page: number) {
    try {
      const people = await axios.get<PeopleSearchResponse>(`/content/search/person`, {
        params: {
          q: encodeURIComponent(query),
          page
        },
        signal: reqController.signal
      });
      return people;
    } catch (err) {
      console.error("People search failed!", err);
      throw err;
    }
  }

  async function searchMulti(query: string, page: number) {
    try {
      return await axios.get<ContentSearch>(`/content/search/multi`, {
        params: {
          q: encodeURIComponent(query),
          page
        },
        signal: reqController.signal
      });
    } catch (err) {
      console.error(`Movies/Tv search failed! (${query})`, err);
      throw err;
    }
  }

  async function searchGames(query: string, page: number) {
    try {
      // Doesn't support pagination, so return if a page higher than 1
      // is requested.
      if (page > 1) {
        return;
      }
      const f = get(serverFeatures);
      if (!f.games) {
        console.debug("game search is not enabled on this server");
        return { data: [] };
      }
      const games = await axios.get<GameSearch[]>(`/game/search`, {
        params: {
          q: encodeURIComponent(query),
          page
        },
        signal: reqController.signal
      });
      return {
        data: games?.data?.map((g) => ({
          ...g,
          media_type: "game"
        })) as GameWithMediaType[]
      };
    } catch (err) {
      console.error(`Game search failed! (${query})`, err);
      throw err;
    }
  }

  async function searchExternalId(id: string, provider: string) {
    try {
      return await axios.get<ContentSearch>(`/content/search/ext/${id}/${provider}`, {
        signal: reqController.signal
      });
    } catch (err) {
      console.error(`External id search failed!`, id, provider, err);
      throw err;
    }
  }

  function checkForExternalIdSearch(query: string) {
    const spl = query.split(":");
    if (spl.length !== 2) {
      // Only ever accept one ':' in query.
      console.debug(
        "checkForExternalIdSearch: One lonely separator not found.. stopping check here."
      );
      return;
    }
    if (!spl[1]) {
      console.info("checkForExternalIdSearch: No id found.");
      return;
    }
    // Check if first part of query is a supported provider.
    let p = spl[0]?.toLowerCase();
    switch (p) {
      // Default names that are supported right out of the box
      case "imdb":
      case "tvdb":
      case "youtube":
      case "wikidata":
      case "facebook":
      case "instagram":
      case "twitter":
      case "tiktok":
        break;
      // Any aliases we want to support
      case "i":
      case "imd":
        p = "imdb";
        break;
      case "wd":
      case "wdt":
        p = "wikidata";
        break;
      case "yt":
        p = "youtube";
        break;
      case "thetvdb":
        p = "tvdb";
        break;
      // If none match, then is invalid provider.
      default:
        console.info("checkForExternalIdSearch: Invalid provider found:", p);
        return;
    }
    console.debug("checkForExternalIdSearch: Found required params:", p, spl[1]);
    return {
      provider: p,
      id: spl[1]
    };
  }

  async function search(query: string) {
    console.debug("search: query:", query);
    if (searchRunning) {
      console.debug("search: already running");
      return;
    }
    if (curPage >= maxContentPage) {
      console.debug("search: max page reached");
      return;
    }
    searchRunning = true;
    const extProvider = checkForExternalIdSearch(query);
    const isExtSearch = !!extProvider;
    reqController = new AbortController();
    try {
      if (isExtSearch) {
        console.log("Search: Performing external id search.");
        const resp = await searchExternalId(extProvider.id, extProvider.provider);
        const data = resp.data;
        if (!data || !data.results) {
          console.warn("Search: No results from external id search.");
          return;
        }
        if (data.results.length === 1 && data.results[0].media_type && data.results[0].id) {
          console.info(
            "Search: Only one result from external id search. Redirecting..",
            data.results[0]
          );
          goto(`/${data.results[0].media_type}/${data.results[0].id}`);
          return;
        }
        console.info("Search: Multiple results from external id search.");
        allSearchResults.push(...data.results);
        searchResults = allSearchResults;
        curPage++;
      } else if (activeSearchFilter) {
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
          const gdata = await searchGames(query, curPage + 1);
          if (gdata) {
            allSearchResults.push(...gdata.data);
          } else {
            console.log("no gdata");
          }
        } else {
          console.error("Active search filter is invalid:", activeSearchFilter);
        }
        maxContentPage = cdata ? cdata.total_pages ?? 1 : 1;
        searchResults = allSearchResults;
        curPage++;
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
          // We only want the current page to increment if multi results
          // request was fulfilled, otherwise the 'try again' button
          // won't work since curPage === maxContentPage (search() wont run).
          curPage++;
        }
        if (r[1].status == "fulfilled" && r[1].value) {
          allSearchResults.push(...r[1].value.data);
        }
        searchResults = allSearchResults;
        // Check this after setting searchResults, so if only game search fails,
        // we can still show the multi results if that succeeded (and vice versa).
        if (r[0].status === "rejected") {
          throw r[0].reason;
        } else if (r[1].status === "rejected") {
          throw r[1].reason;
        }
      }
      console.debug("allSearchResults:", allSearchResults);

      searchRunning = false;
      // If results don't fill the page enough to enable scrolling,
      // the user could be stuck and not be able to get more results
      // to show, run `infiniteScroll` to load more if we can.
      // Smol timeout to give ui time to render so end of page calc
      // can be accurate.
      setTimeout(() => {
        const p = get(page);
        // Quick fix, if user navigates away from search page while response is loading,
        // we don't want to call infiniteScroll or we could end up loading all pages
        // in the background.
        if (p.url?.pathname?.toLowerCase()?.startsWith("/search")) {
          infiniteScroll();
        } else {
          console.debug("No longer on search page, not calling infiniteScroll.");
        }
      }, 250);
    } catch (err: any) {
      if (err?.code === "ERR_CANCELED") {
        console.warn("search was cancelled, not showing error.");
      } else {
        console.error("search failed!", err);
        contentSearchErr = err;
        searchRunning = false;
      }
    }
  }

  async function doCleanSearch() {
    if (!query) {
      console.error("doCleanSearch: No query to use.");
      return;
    }
    console.debug("doCleanSearch()");
    curPage = 0;
    allSearchResults = [];
    searchResults = [];
    search(query);
  }

  async function searchUsers(query: string) {
    return (await axios.get(`/user/search`, { params: { q: query } })).data as PublicUser[];
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
      if (query) await search(query);
      window.addEventListener("scroll", infiniteScroll);
      console.debug(`Page: ${curPage} / ${maxContentPage}`);
    }
  }

  onMount(() => {
    if (!searchQ && query) {
      searchQuery.set(query);
    }
    doCleanSearch();

    window.addEventListener("scroll", infiniteScroll);
    window.addEventListener("resize", infiniteScroll);

    return () => {
      window.removeEventListener("scroll", infiniteScroll);
      window.removeEventListener("resize", infiniteScroll);
    };
  });

  afterNavigate((e) => {
    if (!e.from?.route?.id?.toLowerCase()?.includes("/search")) {
      // AfterNavigate will also be called when this page is mounted,
      // but that won't work for us since the OnMount hook also runs
      // a clean search, which can cause errors when both ran at same
      // time. We can't remove the OnMount hook since it's the only
      // hook to be ran if watcharr is first loaded at a search url.
      // `e.type` is always `goto` (that's how we search) so we can't
      // use that. The only alternative to only run this hook after a
      // navigation on the search page (query change), seems to be
      // checking the `from` property an making sure it's from the
      // `/search` route already.
      return;
    }
    console.log("Query changed (or just loaded first query), performing search");
    reqController.abort("navigated away");
    searchRunning = false;
    doCleanSearch();
  });

  onDestroy(() => {
    console.debug("SEARCH PAGE DESTROYED");
    searchQuery.set("");
    reqController.abort("page destroyed");
  });
</script>

<svelte:head>
  <title>Search Results{query ? ` for '${query}'` : ""}</title>
</svelte:head>

<!-- <span style="position: sticky;top: 70px;">{curPage} / {maxContentPage}</span> -->
<div class="content">
  <div class="inner">
    {#if query}
      {#await searchUsers(query) then results}
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
          {#if $serverFeatures.games}
            <button
              class="plain"
              data-active={activeSearchFilter === "game"}
              on:click={() => setActiveSearchFilter("game")}
            >
              <Icon i="gamepad" wh={20} /> Games
            </button>
          {/if}
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
        <div style="margin-bottom: 60px;">
          <Spinner />
        </div>
      {/if}

      {#if contentSearchErr}
        <div style="margin-bottom: 60px;">
          <Error
            pretty="Failed to load results!"
            error={contentSearchErr}
            onRetry={() => {
              contentSearchErr = undefined;
              search(query);
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
