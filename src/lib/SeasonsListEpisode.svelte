<script lang="ts">
  import type { TMDBSeasonDetailsEpisode } from "@/types";
  import Icon from "./Icon.svelte";
  import { userSettings } from "@/store";

  $: settings = $userSettings;

  export let ep: TMDBSeasonDetailsEpisode;

  let isHidden = false;

  $: {
    if (settings) isHidden = settings.hideSpoilers;
  }
</script>

<li class={isHidden ? "dont-spoil" : ""}>
  {#if ep.still_path}
    <img src={`https://www.themoviedb.org/t/p/w227_and_h127_bestv2/${ep.still_path}`} alt="" />
  {:else}
    <div class="no-still" />
  {/if}
  <span>
    <b>{ep.episode_number}</b>
    <span class="episode-name">{ep.name}</span>
  </span>
  <span
    class="rating"
    title={`TMDB Rating: ${ep.vote_average} out of 10 (based on ${ep.vote_count} votes)`}
  >
    <span>*</span>
    {Math.round(ep.vote_average * 10) / 10}
  </span>
  {#if isHidden}
    <button class="plain spoiler-text" on:click={() => (isHidden = false)}>
      <Icon i="eye-closed" wh={34} />
      <span>Click To Reveal</span>
    </button>
  {/if}
</li>

<style lang="scss">
  li {
    display: flex;
    flex-flow: row;
    gap: 8px;
    position: relative;

    img,
    .no-still {
      width: 227px;
      min-width: 227px;
      height: 127px;
      min-height: 127px;
      border-radius: 10px;
      background-color: rgb(0, 0, 0);
      object-fit: fill;

      @media screen and (max-width: 590px) {
        width: 80%;
        height: auto;
      }

      @media screen and (max-width: 450px) {
        width: 100%;
      }
    }

    span {
      padding: 3px 5px;

      @media screen and (max-width: 590px) {
        text-align: center;
      }
    }

    .rating {
      display: flex;
      align-items: start;
      justify-content: center;
      font-size: 15px;
      color: $rating-color;
      font-weight: bolder;
      overflow: hidden;

      span {
        font-family: "Rampart One";
        -webkit-text-stroke: 1px $rating-color;
        font-size: 25px;
        line-height: 0.7;
        margin-top: 1.5px;
      }
    }

    .spoiler-text {
      display: flex;
      flex-flow: column;
      align-items: center;
      justify-content: center;
      gap: 8px;
      position: absolute;
      width: 100%;
      height: 100%;
      font-weight: bolder;
      font-size: 20px;
      fill: $text-color;
      opacity: 0;
      transition:
        visibility 150ms ease-in,
        opacity 150ms ease-in;
      cursor: pointer;

      span {
        text-shadow: 0 0 6px $bg-color;
      }

      :global(svg) {
        filter: drop-shadow(0 0 8px $bg-color);
      }

      &:hover,
      &:active,
      &:focus {
        opacity: 1;
      }
    }

    .episode-name,
    img {
      transition: filter 150ms ease-out;
    }

    &.dont-spoil {
      .episode-name {
        filter: blur(4px);
      }

      img {
        filter: blur(6px);
      }
      .rating {
        filter:blur(4px)
      }
    }
  }

  @media screen and (max-width: 590px) {
    li {
      align-items: center;
      flex-flow: column;
      width: 100%;
      height: 100%;
    }
  }
</style>
