<script lang="ts">
  import { watchedList } from "@/store";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import axios from "axios";
  import type {
    TMDBDiscoverMovies,
    TMDBDiscoverShows,
    TMDBTrendingAll,
    TMDBUpcomingMovies,
    TMDBUpcomingShows
  } from "@/types";
  import Poster from "@/lib/Poster.svelte";
  import { getWatchedDependedProps } from "@/lib/util/helpers";
  import PosterList from "@/lib/PosterList.svelte";

  $: wList = $watchedList;

  async function allTrending() {
    return (await axios.get(`/content/trending`)).data as TMDBTrendingAll;
  }

  async function trendingMovies() {
    return (await axios.get(`/content/discover/movies`)).data as TMDBDiscoverMovies;
  }

  async function trendingShows() {
    return (await axios.get(`/content/discover/tv`)).data as TMDBDiscoverShows;
  }

  async function upcomingMovies() {
    return (await axios.get(`/content/upcoming/movies`)).data as TMDBUpcomingMovies;
  }

  async function upcomingShows() {
    return (await axios.get(`/content/upcoming/tv`)).data as TMDBUpcomingShows;
  }
</script>

<svelte:head>
  <title>Discover Content</title>
</svelte:head>

<div class="page">
  <h1>Discover</h1>

  <h2 class="norm">Trending Today</h2>
  {#await allTrending()}
    <Spinner />
  {:then trending}
    <PosterList type="vertical">
      {#each trending.results as trend}
        <!-- Possible a person gets returned, but i don't think anyone cares for them -->
        {#if trend.media_type === "movie" || trend.media_type === "tv"}
          <Poster
            media={{ ...trend, media_type: trend.media_type }}
            {...getWatchedDependedProps(trend.id, trend.media_type, wList)}
            small={true}
          />
        {/if}
      {/each}
    </PosterList>
  {:catch err}
    <PageError pretty="Failed to load currently trending!" error={err} />
  {/await}

  <h2 class="norm">Trending Movies</h2>
  {#await trendingMovies()}
    <Spinner />
  {:then movies}
    <PosterList type="vertical">
      {#each movies.results as movie}
        <Poster
          media={{ ...movie, media_type: "movie" }}
          {...getWatchedDependedProps(movie.id, "movie", wList)}
          small={true}
        />
      {/each}
    </PosterList>
  {:catch err}
    <PageError pretty="Failed to load discovered movies!" error={err} />
  {/await}

  <h2 class="norm">Trending Shows</h2>
  {#await trendingShows()}
    <Spinner />
  {:then shows}
    <PosterList type="vertical">
      {#each shows.results as tv}
        <Poster
          media={{ ...tv, media_type: "tv" }}
          {...getWatchedDependedProps(tv.id, "tv", wList)}
          small={true}
        />
      {/each}
    </PosterList>
  {:catch err}
    <PageError pretty="Failed to load discovered shows!" error={err} />
  {/await}

  <h2 class="norm">Upcoming Movies</h2>
  {#await upcomingMovies()}
    <Spinner />
  {:then shows}
    <PosterList type="vertical">
      {#each shows.results as tv}
        <Poster
          media={{ ...tv, media_type: "movie" }}
          {...getWatchedDependedProps(tv.id, "movie", wList)}
          small={true}
        />
      {/each}
    </PosterList>
  {:catch err}
    <PageError pretty="Failed to load upcoming movies!" error={err} />
  {/await}

  <h2 class="norm">Upcoming Shows</h2>
  {#await upcomingShows()}
    <Spinner />
  {:then shows}
    <PosterList type="vertical">
      {#each shows.results as tv}
        <Poster
          media={{ ...tv, media_type: "tv" }}
          {...getWatchedDependedProps(tv.id, "tv", wList)}
          small={true}
        />
      {/each}
    </PosterList>
  {:catch err}
    <PageError pretty="Failed to load upcoming shows!" error={err} />
  {/await}
</div>

<style lang="scss">
  .page {
    display: flex;
    flex-flow: column;
    margin-left: auto;
    margin-right: auto;
    padding: 20px 50px;
    max-width: 1200px;

    h1 {
      margin-bottom: 15px;
    }

    h2 {
      font-variant: small-caps;
    }

    @media screen and (max-width: 500px) {
      padding: 20px;
    }
  }
</style>
