<script lang="ts">
  import PageError from "@/lib/PageError.svelte";
  import PersonPoster from "@/lib/poster/PersonPoster.svelte";
  import Rating from "@/lib/Rating.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import Status from "@/lib/Status.svelte";
  import HorizontalList from "@/lib/HorizontalList.svelte";
  import { contentExistsOnJellyfin, updateWatched } from "@/lib/util/api";
  import { watchedList } from "@/store";
  import type {
    TMDBContentCredits,
    TMDBContentCreditsCrew,
    TMDBMovieDetails,
    WatchedStatus
  } from "@/types";
  import axios from "axios";
  import { getTopCrew } from "@/lib/util/helpers.js";
  import Activity from "@/lib/Activity.svelte";
  import Title from "@/lib/content/Title.svelte";
  import VideoEmbedModal from "@/lib/content/VideoEmbedModal.svelte";
  import ProvidersList from "@/lib/content/ProvidersList.svelte";
  import Icon from "@/lib/Icon.svelte";
  import SimilarContent from "@/lib/content/SimilarContent.svelte";
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import RequestMovie from "@/lib/request/RequestMovie.svelte";

  export let data;

  let trailer: string | undefined;
  let requestModalShown = false;
  let trailerShown = false;
  let jellyfinUrl: string | undefined;

  $: wListItem = $watchedList.find((w) => w.content.tmdbId === data.movieId);

  let movieId: number | undefined;
  let movie: TMDBMovieDetails | undefined;
  let pageError: Error | undefined;

  onMount(() => {
    const unsubscribe = page.subscribe((value) => {
      console.log(value);
      const params = value.params;
      if (params && params.id) {
        movieId = Number(params.id);
      }
    });

    return unsubscribe;
  });

  $: {
    (async () => {
      try {
        movie = undefined;
        pageError = undefined;
        if (!movieId) {
          return;
        }
        const data = (await axios.get(`/content/movie/${movieId}`)).data as TMDBMovieDetails;
        if (data.videos?.results?.length > 0) {
          const t = data.videos.results.find((v) => v.type?.toLowerCase() === "trailer");
          if (t?.key) {
            if (t?.site?.toLowerCase() === "youtube") {
              trailer = `https://www.youtube.com/embed/${t?.key}`;
            }
          }
        }
        contentExistsOnJellyfin("movie", data.title, data.id).then((j) => {
          if (j?.hasContent && j?.url !== "") {
            jellyfinUrl = j.url;
          }
        });
        movie = data;
      } catch (err: any) {
        movie = undefined;
        pageError = err;
      }
    })();
  }

  async function getMovieCredits() {
    const credits = (await axios.get(`/content/movie/${data.movieId}/credits`))
      .data as TMDBContentCredits & { topCrew: TMDBContentCreditsCrew[] };
    if (credits.crew?.length > 0) {
      credits.topCrew = getTopCrew(credits.crew);
    }
    return credits;
  }

  function contentChanged(newStatus?: WatchedStatus, newRating?: number, newThoughts?: string) {
    updateWatched(data.movieId, "movie", newStatus, newRating, newThoughts);
  }
</script>

{#if pageError}
  <PageError pretty="Failed to load movie!" error={pageError} />
{:else if !movie}
  <Spinner />
{:else if Object.keys(movie).length > 0}
  <div>
    <div class="content">
      {#if movie?.backdrop_path}
        <img
          class="backdrop"
          src={"https://www.themoviedb.org/t/p/w1920_and_h800_multi_faces" + movie.backdrop_path}
          alt=""
        />
      {/if}
      <div class="vignette" />

      <div class="details-container">
        <img class="poster" src={"https://image.tmdb.org/t/p/w500" + movie.poster_path} alt="" />

        <div class="details">
          <Title
            title={movie.title}
            homepage={movie.homepage}
            releaseDate={movie.release_date}
            voteAverage={movie.vote_average}
            voteCount={movie.vote_count}
          />

          <span class="quick-info">
            <span>{movie.runtime}m</span>

            <div>
              {#each movie.genres as g, i}
                <span>{g.name}{i !== movie.genres.length - 1 ? ", " : ""}</span>
              {/each}
            </div>
          </span>

          <!-- <span>{movie.tagline}</span> -->

          <!-- {movie.status} -->

          <span style="font-weight: bold; font-size: 14px;">Overview</span>
          <p>{movie.overview}</p>

          <div class="btns">
            {#if trailer}
              <button on:click={() => (trailerShown = !trailerShown)}>View Trailer</button>
              {#if trailerShown}
                <VideoEmbedModal embed={trailer} closed={() => (trailerShown = false)} />
              {/if}
            {/if}
            {#if jellyfinUrl}
              <a class="btn" href={jellyfinUrl} target="_blank">
                <Icon i="jellyfin" wh={14} />Play On Jellyfin
              </a>
            {/if}
            <button on:click={() => (requestModalShown = !requestModalShown)}>Request</button>
          </div>

          <ProvidersList providers={movie["watch/providers"]} />
        </div>
      </div>
    </div>

    {#if requestModalShown}
      <RequestMovie content={movie} onClose={() => (requestModalShown = false)} />
    {/if}

    <div class="page">
      <div class="review">
        <!-- <span>What did you think?</span> -->
        <Rating rating={wListItem?.rating} onChange={(n) => contentChanged(undefined, n)} />
        <Status status={wListItem?.status} onChange={(n) => contentChanged(n)} />
        {#if wListItem}
          <textarea
            name="Thoughts"
            rows="3"
            placeholder={`My thoughts on ${movie.title}`}
            value={wListItem?.thoughts}
            on:blur={(e) => {
              console.log(e.currentTarget?.value);
              contentChanged(undefined, undefined, e.currentTarget?.value);
            }}
          />
        {/if}
      </div>

      {#await getMovieCredits()}
        <Spinner />
      {:then credits}
        {#if credits.topCrew?.length > 0}
          <div class="creators">
            {#each credits.topCrew as crew}
              <div>
                <span>{crew.name}</span>
                <span>{crew.job}</span>
              </div>
            {/each}
          </div>
        {/if}

        {#if credits.cast?.length > 0}
          <HorizontalList title="Cast">
            {#each credits.cast?.slice(0, 50) as cast}
              <PersonPoster
                id={cast.id}
                name={cast.name}
                path={cast.profile_path}
                role={cast.character}
                zoomOnHover={false}
              />
            {/each}
          </HorizontalList>
        {/if}
      {:catch err}
        <Error error={err} pretty="Failed to load cast!" />
      {/await}

      <SimilarContent type="movie" similar={movie.similar} />

      {#if wListItem}
        <Activity activity={wListItem?.activity} />
      {/if}
    </div>
  </div>
{:else}
  Movie not found
{/if}

<style lang="scss">
  .content {
    position: relative;
    color: white;

    img.backdrop {
      position: absolute;
      left: 0;
      top: 0;
      z-index: -2;
      width: 100%;
      height: 100%;
      object-fit: cover;
      filter: $backdrop-filter;
      mix-blend-mode: $backdrop-mix-blend-mode;
      mask-image: $backdrop-mask-image;
    }

    .vignette {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background-color: rgba($color: #000000, $alpha: 0.7);
      z-index: -1;
      mask-image: $backdrop-mask-image;
    }

    .details-container {
      display: flex;
      flex-flow: row;
      gap: 35px;
      max-width: 1100px;
      padding: 40px 80px;
      margin-left: auto;
      margin-right: auto;

      img.poster {
        width: 235px;
        height: 100%;
        box-shadow: 0px 0px 14px -4px #9c8080;
        border-radius: 12px;
      }

      .details {
        display: flex;
        flex-flow: column;
        gap: 5px;

        .quick-info {
          display: flex;
          gap: 10px;
          margin-bottom: 8px;
        }

        p {
          font-size: 14px;
        }

        .btns {
          display: flex;
          flex-flow: row;
          gap: 8px;
          margin-top: 18px;

          a.btn,
          button {
            max-width: fit-content;
            overflow: hidden;
            animation: 50ms cubic-bezier(0.86, 0, 0.07, 1) forwards otherbtn;
            white-space: nowrap;
            gap: 6px;
            justify-content: flex-start;
            font-size: 14px;

            @keyframes otherbtn {
              from {
                width: 0px;
              }
              to {
                width: 100%;
              }
            }
          }
        }
      }

      @media screen and (max-width: 700px) {
        padding: 40px;
      }

      @media screen and (max-width: 590px) {
        flex-flow: column;
        align-items: center;
      }
    }
  }

  .page {
    display: flex;
    flex-flow: column;
    align-items: center;
    margin-left: auto;
    margin-right: auto;
    gap: 30px;
    padding: 20px 50px;
    max-width: 1200px;

    @media screen and (max-width: 500px) {
      padding: 20px;
    }
  }

  .review {
    display: flex;
    flex-flow: column;
    gap: 10px;
    max-width: 380px;
  }

  .creators {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 35px;
    margin: 10px 60px;

    div {
      display: flex;
      flex-flow: column;
      min-width: 150px;

      span:first-child {
        font-weight: bold;
      }
    }
  }
</style>
