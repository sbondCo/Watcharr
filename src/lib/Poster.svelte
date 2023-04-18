<script lang="ts">
  import type { WatchedStatus } from "@/types";
  import Icon from "./Icon.svelte";

  export let poster: string | undefined;
  export let title: string | undefined;
  export let desc: string | undefined;
  export let rating: number | undefined = undefined;
  export let status: WatchedStatus | undefined = undefined;
  export let onBtnClicked: (type: WatchedStatus, rating?: number) => void = () => {};
  export let onRatingChanged: (rating: number) => void = () => {};

  function handleStarClick(r: number) {
    if (r == rating) return;
    onRatingChanged(r);
  }

  function wBtnClicked(type: WatchedStatus) {
    onBtnClicked(type, rating);
  }
</script>

<li>
  <div class={`container${!poster ? " details-shown" : ""}`}>
    {#if poster}
      <img loading="lazy" src={poster} alt="poster" />
    {/if}
    <div class="inner">
      <h2>{title}</h2>
      <span>{desc}</span>

      <div class="buttons">
        <button><span>*</span><span>8</span></button>
        <button><Icon i="clock" /></button>
      </div>
    </div>
  </div>
</li>

<style lang="scss">
  .container {
    display: flex;
    flex-flow: column;
    background-color: rgb(48, 45, 45);
    overflow: hidden;
    flex: 1 1;
    border-radius: 5px;
    width: 170px;
    height: 100%;
    min-height: 256.367px;
    position: relative;
    // aspect-ratio: 2/3;
    transition: all 150ms ease-in;

    img {
      width: 100%;
      height: 100%;
    }

    .inner {
      position: absolute;
      visibility: hidden;
      display: flex;
      flex-flow: column;
      top: 0;
      height: 100%;
      width: 100%;
      padding: 10px;
      background-color: transparent;

      h2 {
        font-family: unset;
        font-size: 18px;
      }

      span {
        color: white;
        margin: 5px 0 5px 0;
        font-size: 9px;
        display: -webkit-box;
        -webkit-line-clamp: 5;
        -webkit-box-orient: vertical;
        hyphens: auto;
        overflow: hidden;
      }

      .buttons {
        display: flex;
        flex-flow: row;
        margin-top: auto;
        gap: 10px;
        height: 35px;

        button {
          padding: 3px;

          /** Rating */
          &:first-child {
            font-family: "Rampart One";

            span {
              color: black;
              -webkit-text-stroke: 1.5px black;

              &:first-child {
                font-size: 39px;
                letter-spacing: 10px;
                line-height: 52px;
                height: 42px;
              }

              &:nth-child(2) {
                font-size: 22px;
                height: 35px; // quick fix to make the rating num look centered - text-stroke makes it look not centered
              }
            }

            &:hover span {
              color: gold;
              -webkit-text-stroke: 1.5px gold;
            }
          }

          /** Status */
          &:nth-child(2) {
            width: 40%;
          }
        }
      }
    }

    &:hover {
      transform: scale(1.3);
      z-index: 99;
    }

    &:hover,
    &:global(.details-shown) {
      img {
        filter: blur(4px) grayscale(80%);
        // This makes the background very dark,
        // but atleast the text is visible.. may want to change later.
        mix-blend-mode: multiply;
      }

      .inner {
        color: white;
        visibility: visible;
      }
    }
  }
</style>
