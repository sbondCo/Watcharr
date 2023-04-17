<script lang="ts">
  import Rating from "@/lib/Rating.svelte";
  import Status from "@/lib/Status.svelte";
  import { updateWatched } from "@/lib/api";
  import { watchedList } from "@/store";
  import type { TMDBMovieDetails, WatchedStatus } from "@/types";

  export let data: TMDBMovieDetails;
  let releaseDate = new Date(Date.parse(data.release_date));

  $: wListItem = $watchedList.find((w) => w.content.id === data.id);

  function statusChanged(newStatus: WatchedStatus) {
    if (wListItem?.id)
      updateWatched(wListItem.id, newStatus)
        ?.then(() => {
          wListItem!.status = newStatus;
          $watchedList = $watchedList;
        })
        .catch((err) => {
          console.error(err);
        });
  }

  console.log(data);
</script>

{#if Object.keys(data).length > 0}
  <div>
    <div class="content">
      <img
        class="backdrop"
        src={"https://www.themoviedb.org/t/p/w1920_and_h800_multi_faces" + data.backdrop_path}
        alt=""
      />
      <div class="vignette" />

      <img src={"https://image.tmdb.org/t/p/w500" + data.poster_path} alt="" />

      <div class="details">
        <span class="title-container">
          <a href={data.homepage} target="_blank">{data.title}</a>
          <span>{releaseDate.getFullYear()}</span>
        </span>

        <span class="quick-info">
          <span>{data.runtime}m</span>

          <div>
            {#each data.genres as g, i}
              <span>{g.name}{i !== data.genres.length - 1 ? ", " : ""}</span>
            {/each}
          </div>
        </span>

        <!-- <span>{data.tagline}</span> -->

        <!-- {data.status} -->

        <span style="font-weight: bold; font-size: 14px;">Overview</span>
        <p>{data.overview}</p>
      </div>
    </div>

    <div class="page">
      <div class="review">
        <!-- <span>What did you think?</span> -->
        <Rating rating={wListItem?.rating} />
        <Status status={wListItem?.status} onChange={statusChanged} />
      </div>

      <div class="creators">
        <div>
          <span>Mr Boombastic</span>
          <span>Director</span>
        </div>
        <div>
          <span>Mr Boombastic</span>
          <span>Writer</span>
        </div>
        <div>
          <span>Mr Boombastic</span>
          <span>Producer</span>
        </div>
        <div>
          <span>Mr Boombastic</span>
          <span>Producer</span>
        </div>
      </div>
    </div>
  </div>
{:else}
  Movie not found
{/if}

<style lang="scss">
  .content {
    position: relative;
    display: flex;
    flex-flow: row;
    gap: 15px;
    padding: 20px;
    color: white;

    img {
      width: 170px;
      height: 100%;
    }

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

      // ul {
      //   display: flex;
      //   flex-flow: row;
      //   gap: 10px;
      //   list-style: none;

      //   li {
      //     padding: 5px 8px;
      //   }
      // }
    }

    @media screen and (max-width: 450px) {
      flex-flow: column;
      align-items: center;
    }
  }

  .page {
    display: flex;
    flex-flow: column;
    align-items: center;
    gap: 30px;
    padding: 20px 50px;

    @media screen and (max-width: 500px) {
      padding: 20px 30px;
    }
  }

  .review {
    display: flex;
    flex-flow: column;
    gap: 10px;
  }

  .creators {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 35px;

    div {
      display: flex;
      flex-flow: column;

      span:first-child {
        font-weight: bold;
      }
    }
  }
</style>
