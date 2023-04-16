<script lang="ts">
  import type { TMDBMovieDetails } from "@/types";
  import { onMount } from "svelte";

  export let data: TMDBMovieDetails;

  let releaseDate = new Date(Date.parse(data.release_date));

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

        <!-- <span>{data.tagline}</span> -->

        <span class="quick-info">
          <span>{data.runtime}m</span>

          <!-- <span>|</span> -->

          <div>
            {#each data.genres as g, i}
              <span>{g.name}{i !== data.genres.length - 1 ? ", " : ""}</span>
            {/each}
          </div>
        </span>

        <!-- {data.status} -->

        <span style="font-weight: bold; font-size: 14px;">Overview</span>
        <p>{data.overview}</p>
      </div>
    </div>

    <div class="page">
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

      <div class="review">
        <span>What did you think?</span>

        <div id="rating-container" class="rating">
          <button class="plain">*</button>
          <button class="plain">*</button>
          <button class="plain">*</button>
          <button class="plain">*</button>
          <button class="plain">*</button>
          <button class="plain">*</button>
          <button class="plain">*</button>
          <button class="plain">*</button>
          <button class="plain">*</button>
          <button class="plain">*</button>
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
    gap: 30px;
    padding: 20px 50px;

    @media screen and (max-width: 500px) {
      padding: 20px 30px;
    }
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

  .rating {
    display: flex;
    flex-flow: row-reverse;
    align-items: center;
    justify-content: center;
    color: black;
    -webkit-text-stroke: 1.5px black;
    cursor: pointer;
    overflow: hidden;
    margin: 10px 0 10px 0;

    button {
      font-size: 55px;
      font-family: "Rampart One";
      letter-spacing: 10px;
      line-height: 52px;
      height: 38px;

      &:hover,
      &:hover ~ button,
      &:global(.lit),
      &:global(.lit ~ button) {
        color: gold;
        -webkit-text-stroke: 1.5px gold;
      }

      @media screen and (max-width: 500px) {
        font-size: 40px;
      }
    }
  }
</style>
