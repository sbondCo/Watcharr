<script lang="ts">
  import { watchedList } from "@/store";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import axios from "axios";
  import type { TMDBDiscoverMovies, TMDBDiscoverShows } from "@/types";
  import Poster from "@/lib/Poster.svelte";
  import { getWatchedDependedProps } from "@/lib/util/helpers";
  import PosterList from "@/lib/PosterList.svelte";

  $: wList = $watchedList;

  async function trendingMovies() {
    return (await axios.get(`/content/discover/movies`)).data as TMDBDiscoverMovies;
  }

  async function trendingShows() {
    return (await axios.get(`/content/discover/tv`)).data as TMDBDiscoverShows;
  }
</script>

<svelte:head>
  <title>Discover Content</title>
</svelte:head>

<div class="page">
  <h1>Discover</h1>

  <h2>Trending Movies</h2>
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

  <h2>Trending Shows</h2>
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
</div>

<style>
  .page {
    display: flex;
    flex-flow: column;
    margin-left: auto;
    margin-right: auto;
    /* gap: 30px; */
    padding: 20px 50px;
    max-width: 1200px;

    @media screen and (max-width: 500px) {
      padding: 20px;
    }
  }
</style>
