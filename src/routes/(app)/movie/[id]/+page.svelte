<script lang="ts">
  import Error from "@/lib/Error.svelte";
  import PageError from "@/lib/PageError.svelte";
  import PersonPoster from "@/lib/PersonPoster.svelte";
  import Rating from "@/lib/Rating.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import Status from "@/lib/Status.svelte";
  import HorizontalList from "@/lib/HorizontalList.svelte";
  import { updateWatched } from "@/lib/util/api";
  import { watchedList } from "@/store";
  import type {
    TMDBContentCredits,
    TMDBContentCreditsCrew,
    TMDBMovieDetails,
    TMDBWatchProvider,
    WatchedStatus
  } from "@/types";
  import axios from "axios";
  import { getTopCrew } from "@/lib/util/helpers.js";
  import Activity from "@/lib/Activity.svelte";
  import Title from "@/lib/content/Title.svelte";
  import VideoEmbedModal from "@/lib/content/VideoEmbedModal.svelte";
  import Icon from "@/lib/Icon.svelte";

  export let data;

  let trailer: string | undefined;
  let trailerShown = false;
	let providers: TMDBWatchProvider[] | undefined;

  $: wListItem = $watchedList.find((w) => w.content.tmdbId === data.movieId);

  async function getMovie() {
    const movie = (await axios.get(`/content/movie/${data.movieId}`)).data as TMDBMovieDetails;
    if (movie.videos?.results?.length > 0) {
      const t = movie.videos.results.find((v) => v.type?.toLowerCase() === "trailer");
      if (t?.key) {
        if (t?.site?.toLowerCase() === "youtube") {
          trailer = `https://www.youtube.com/embed/${t?.key}`;
        }
      }
    }

		// TODO: Move to server?
		if (movie["watch/providers"]?.results?.GB?.flatrate) {
			providers = movie["watch/providers"]?.results?.GB?.flatrate
		}

    return movie;
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

{#await getMovie()}
  <Spinner />
{:then movie}
  {#if Object.keys(movie).length > 0}
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
            </div>

            {#if providers}
            <div class="streaming-providers">
              {#each providers as provider}
                <Icon i={provider.provider_name} wh={50}/>
              {/each}
            </div>
            {/if}
          </div>
        </div>
      </div>

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
            <div class="cast">
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
            </div>
          {/if}
        {:catch err}
          <Error error={err} pretty="Failed to load cast!" />
        {/await}

        {#if wListItem}
          <Activity activity={wListItem?.activity} />
        {/if}
      </div>
    </div>
  {:else}
    Movie not found
  {/if}
{:catch err}
  <PageError pretty="Failed to load movie!" error={err} />
{/await}

<style lang="scss">
  .content {
    position: relative;
    color: white;
    margin-bottom: 15px;

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
    }

    .vignette {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background-color: rgba($color: #000000, $alpha: 0.7);
      z-index: -1;
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

				.streaming-providers {
          display: flex;
          gap: 15px;
					margin-top: 8px;
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

          button {
            width: fit-content;
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

  .cast {
    width: 100%;
    overflow-x: auto;
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
