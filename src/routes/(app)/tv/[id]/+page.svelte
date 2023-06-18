<script lang="ts">
  import Error from "@/lib/Error.svelte";
  import HorizontalList from "@/lib/HorizontalList.svelte";
  import PageError from "@/lib/PageError.svelte";
  import PersonPoster from "@/lib/PersonPoster.svelte";
  import Rating from "@/lib/Rating.svelte";
  import SeasonsList from "@/lib/SeasonsList.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import Status from "@/lib/Status.svelte";
  import { updateWatched } from "@/lib/util/api";
  import { getTopCrew } from "@/lib/util/helpers.js";
  import { watchedList } from "@/store";
  import type {
    TMDBContentCredits,
    TMDBContentCreditsCrew,
    TMDBShowDetails,
    WatchedStatus
  } from "@/types";
  import axios from "axios";

  export let data;

  $: wListItem = $watchedList.find((w) => w.content.tmdbId === data.tvId);

  async function getShow() {
    return (await axios.get(`/content/tv/${data.tvId}`)).data as TMDBShowDetails;
  }

  async function getTvCredits() {
    const credits = (await axios.get(`/content/tv/${data.tvId}/credits`))
      .data as TMDBContentCredits & { topCrew: TMDBContentCreditsCrew[] };
    if (credits.crew?.length > 0) {
      credits.topCrew = getTopCrew(credits.crew);
    }
    return credits;
  }

  function contentChanged(newStatus?: WatchedStatus, newRating?: number) {
    updateWatched(data.tvId, "tv", newStatus, newRating);
  }
</script>

{#await getShow()}
  <Spinner />
{:then show}
  {#if Object.keys(show).length > 0}
    <div>
      <div class="content">
        <img
          class="backdrop"
          src={"https://www.themoviedb.org/t/p/w1920_and_h800_multi_faces" + show.backdrop_path}
          alt=""
        />
        <div class="vignette" />

        <div class="details-container">
          <img class="poster" src={"https://image.tmdb.org/t/p/w500" + show.poster_path} alt="" />

          <div class="details">
            <span class="title-container">
              <a href={show.homepage} target="_blank">{show.name}</a>
              <span>{new Date(Date.parse(show.first_air_date)).getFullYear()}</span>
            </span>

            <span class="quick-info">
              <span>{show.episode_run_time}m</span>

              <div>
                {#each show.genres as g, i}
                  <span>{g.name}{i !== show.genres.length - 1 ? ", " : ""}</span>
                {/each}
              </div>
            </span>

            <span style="font-weight: bold; font-size: 14px;">Overview</span>
            <p>{show.overview}</p>
          </div>
        </div>
      </div>

      <div class="page">
        <div class="review">
          <!-- <span>What did you think?</span> -->
          <Rating rating={wListItem?.rating} onChange={(n) => contentChanged(undefined, n)} />
          <Status status={wListItem?.status} onChange={(n) => contentChanged(n)} />
        </div>

        {#await getTvCredits()}
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
        {:catch err}
          <Error error={err} pretty="Failed to load cast!" />
        {/await}

        <SeasonsList tvId={data.tvId} seasons={show.seasons} />
      </div>
    </div>
  {:else}
    Show not found
  {/if}
{:catch err}
  <PageError pretty="Failed to load tv show!" error={err} />
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
      filter: blur(4px) grayscale(80%);
      mix-blend-mode: multiply;
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

        .title-container {
          a {
            color: white;
            text-decoration: none;
            font-size: 30px;
            font-weight: bold;
            padding-right: 3px;
          }

          span {
            font-size: 20px;
            color: rgba($color: #fff, $alpha: 0.7);
          }
        }

        .quick-info {
          display: flex;
          gap: 10px;
          margin-bottom: 8px;
        }

        p {
          font-size: 14px;
        }
      }

      @media screen and (max-width: 700px) {
        padding: 40px;
      }

      @media screen and (max-width: 570px) {
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
