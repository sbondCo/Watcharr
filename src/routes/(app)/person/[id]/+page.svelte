<script lang="ts">
  import Error from "@/lib/Error.svelte";
  import PageError from "@/lib/PageError.svelte";
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { removeWatched, updateWatched } from "@/lib/util/api";
  import { getWatchedDependedProps } from "@/lib/util/helpers";
  import { watchedList } from "@/store";
  import type { TMDBPersonCombinedCredits, TMDBPersonDetails } from "@/types";
  import axios from "axios";

  export let data;

  $: wList = $watchedList;

  async function getPerson() {
    return (await axios.get(`/content/person/${data.personId}`)).data as TMDBPersonDetails;
  }

  async function getPersonCredits() {
    const credits = (await axios.get(`/content/person/${data.personId}/credits`))
      .data as TMDBPersonCombinedCredits;
    credits.cast = credits.cast?.sort((a, b) => b.vote_count - a.vote_count);
    return credits;
  }
</script>

<div>
  {#await getPerson()}
    <Spinner />
  {:then person}
    {#if Object.keys(person).length > 0}
      <div class="content">
        <img
          class="backdrop"
          src={"https://www.themoviedb.org/t/p/w1920_and_h800_multi_faces" + person.profile_path}
          alt=""
        />
        <div class="vignette" />

        <div class="details-container">
          <img
            class="poster"
            src={"https://image.tmdb.org/t/p/w500" + person.profile_path}
            alt=""
          />

          <div class="details">
            <span class="title-container">
              <a href={person.homepage} target="_blank">{person.name}</a>
              <span />
            </span>

            {#if person.biography}
              <span style="font-weight: bold; font-size: 14px;">Biography</span>
              <!-- Show just the first paragraph -->
              <p>{person.biography?.split("\n")[0]}</p>
            {/if}

            <div class="detail-info">
              {#if person.known_for_department}
                <div>
                  <span>Department</span>
                  <span>{person.known_for_department}</span>
                </div>
              {/if}
              {#if person.place_of_birth}
                <div>
                  <span>Born In</span>
                  <span>{person.place_of_birth}</span>
                </div>
              {/if}
              {#if person.birthday}
                <div>
                  <span>Born On</span>
                  <span>{new Date(Date.parse(person.birthday)).toLocaleDateString()}</span>
                </div>
              {/if}
              {#if person.deathday}
                <div>
                  <span>Died On</span>
                  <span>{new Date(Date.parse(person.deathday)).toLocaleDateString()}</span>
                </div>
              {/if}
            </div>
          </div>
        </div>
      </div>
    {:else}
      person not found
    {/if}
  {:catch err}
    <PageError pretty="Failed to load tv person!" error={err} />
  {/await}

  {#await getPersonCredits()}
    <Spinner />
  {:then credits}
    <div class="page">
      <PosterList>
        {#each credits?.cast as c}
          <Poster
            media={c}
            onStatusChanged={(t) => updateWatched(c.id, c.media_type, t)}
            onRatingChanged={(r) => updateWatched(c.id, c.media_type, undefined, r)}
            onDeleteClicked={() => removeWatched(c.id)}
            {...getWatchedDependedProps(c.id, c.media_type, wList)}
          />
        {/each}
      </PosterList>
    </div>
  {:catch err}
    <Error pretty="Failed to load credits!" error={err} />
  {/await}
</div>

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
      /* mix-blend-mode: multiply; */
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

        p {
          font-size: 14px;
          max-height: 200px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: pre-line;
        }

        .detail-info {
          display: flex;
          flex-wrap: wrap;
          gap: 35px;
          margin-top: 10px;
          font-size: 14px;

          div {
            display: flex;
            flex-flow: column;

            span:first-child {
              font-weight: bold;
            }
          }
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
    gap: 30px;
    padding: 10px 0px;
  }
</style>
