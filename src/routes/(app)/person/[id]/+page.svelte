<script lang="ts">
  import { page } from "$app/stores";
  import Error from "@/lib/Error.svelte";
  import PageError from "@/lib/PageError.svelte";
  import Poster from "@/lib/poster/Poster.svelte";
  import PosterList from "@/lib/poster/PosterList.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import DropDown from "@/lib/DropDown.svelte";
  import { getWatchedDependedProps } from "@/lib/util/helpers";
  import { watchedList } from "@/store";
  import type { TMDBPersonCombinedCredits, TMDBPersonDetails } from "@/types";
  import axios from "axios";
  import { onMount } from "svelte";
  import Checkbox from "@/lib/Checkbox.svelte";

  export let data;

  $: wList = $watchedList;

  let personId: number | undefined;
  let person: TMDBPersonDetails | undefined;
  let pageError: Error | undefined;
  let sortOption = "Vote count";
  let credits: TMDBPersonCombinedCredits | undefined;
  let onMyListFilter = false;

  onMount(() => {
    const unsubscribe = page.subscribe((value) => {
      const params = value.params;
      if (params && params.id) {
        personId = Number(params.id);
      }
    });

    return unsubscribe;
  });

  $: if (personId) {
    fetchPersonData();
  }

  $: if (sortOption && credits) {
    sortCredits(sortOption);
  }

  async function fetchPersonData() {
    try {
      person = undefined;
      pageError = undefined;
      if (!personId) {
        return;
      }
      person = await getPerson(personId);
      await updatePersonCredits();
      sortCredits(sortOption);
    } catch (err: any) {
      person = undefined;
      pageError = err;
    }
  }

  async function getPerson(id: number) {
    return (await axios.get(`/content/person/${id}`)).data as TMDBPersonDetails;
  }

  async function updatePersonCredits() {
    credits = (await axios.get(`/content/person/${data.personId}/credits`))
      .data as TMDBPersonCombinedCredits;
    credits.cast = credits.cast.filter((v, i, a) => a.findIndex((t) => t.id === v.id) === i); // remove duplicate entries. If an actor has multiple roles in a single movie, it would otherwise show up multiple times
  }

  function sortCredits(sortOption: string) {
    if (!credits || !credits.cast) return;

    switch (sortOption) {
      case "Vote count":
        credits.cast.sort((a, b) => b.vote_count - a.vote_count);
        break;
      case "Newest":
        credits.cast.sort((a, b) => {
          const dateA = new Date(a.release_date || a.first_air_date).valueOf();
          const dateB = new Date(b.release_date || b.first_air_date).valueOf();

          // if a date is missing it should be sorted first since it's probably a future release
          if (!a.release_date && !a.first_air_date) return -1;
          if (!b.release_date && !b.first_air_date) return 1;

          return dateB - dateA;
        });
        break;
      case "Oldest":
        credits.cast.sort((a, b) => {
          const dateA = new Date(a.release_date || a.first_air_date).valueOf();
          const dateB = new Date(b.release_date || b.first_air_date).valueOf();

          // if a date is missing it should be sorted last since it's probably a future release
          if (!a.release_date && !a.first_air_date) return 1;
          if (!b.release_date && !b.first_air_date) return -1;

          return dateA - dateB;
        });
        break;
    }
    credits.cast = credits.cast;
  }
</script>

<svelte:head>
  <title>{person?.name ? `${person.name} - ` : ""}Person</title>
</svelte:head>

<div>
  {#if pageError}
    <PageError pretty="Failed to load person!" error={pageError} />
  {:else if !person}
    <Spinner />
  {:else if Object.keys(person).length > 0}
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
      <div class="filters">
        <div class="listFilter">
          <span>On my list</span>
          <Checkbox name="On my list" bind:value={onMyListFilter} />
        </div>
        <DropDown
          bind:active={sortOption}
          placeholder="Vote count"
          options={["Vote count", "Newest", "Oldest"]}
          isDropDownItem={false}
          showActiveElementsInOptions={true}
        />
      </div>
      {#if credits}
        <div class="page">
          <PosterList>
            {#each credits.cast as c (c.id)}
              <Poster
                media={c}
                {...getWatchedDependedProps(c.id, c.media_type, wList)}
                fluidSize
                hideIfNotOnList={onMyListFilter}
              />
            {/each}
          </PosterList>
        </div>
      {:else}
        <Spinner />
      {/if}
    {:else}
      person not found
    {/if}
  {:else}
    <Error error="Person not found" pretty="Person not found" />
  {/if}
</div>

<style lang="scss">
  .filters {
    align-items: center;
    display: flex;
    justify-content: flex-end;
    gap: 30px;
    margin: 0 auto;
    padding-left: 20px;
    padding-right: 20px;
    width: 100%;
    /* Same as in PosterList */
    max-width: 1200px;

    .listFilter {
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }

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
