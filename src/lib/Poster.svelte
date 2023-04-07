<script lang="ts">
  import type { Rating, WatchedStatus } from "@/types";
  import { onMount } from "svelte";
  import Icon from "./Icon.svelte";

  // export let content: Content;
  export let poster: string | undefined;
  export let title: string | undefined;
  export let desc: string | undefined;
  export let rating: Rating = undefined;
  export let status: WatchedStatus | undefined = undefined;
  export let onBtnClicked: (type: WatchedStatus, rating?: Rating) => void = () => {};
  export let onRatingChanged: (rating: Rating) => void = () => {};

  let ratingContainer: HTMLDivElement;

  function handleRating(r: Rating) {
    resetRating();
    rating = r;
    const el = ratingContainer.querySelector(`#s${r}`);
    if (el) {
      el.classList.add("lit");
      onRatingChanged(rating);
    }
  }

  function resetRating() {
    let ratingsContainer = document.getElementById("rating-container");
    const stars = ratingsContainer?.querySelectorAll("button");
    if (!stars) return;
    for (let i = 0; i < stars.length; i++) {
      stars[i].classList.remove("lit");
    }
    rating = undefined;
  }

  function wBtnClicked(type: WatchedStatus) {
    onBtnClicked(type, rating);
  }

  onMount(() => {
    if (typeof rating === "number") handleRating(rating);
  });
</script>

<li>
  <div class="container">
    <img loading="lazy" src={poster} alt="poster" />
    <div class="inner">
      <h2>{title}</h2>
      <span>{desc}</span>

      <div
        id="rating-container"
        class="rating"
        on:dblclick={resetRating}
        bind:this={ratingContainer}
      >
        <button class="plain" id="s5" on:click={() => handleRating(5)}>*</button>
        <button class="plain" id="s4" on:click={() => handleRating(4)}>*</button>
        <button class="plain" id="s3" on:click={() => handleRating(3)}>*</button>
        <button class="plain" id="s2" on:click={() => handleRating(2)}>*</button>
        <button class="plain" id="s1" on:click={() => handleRating(1)}>*</button>
      </div>

      <div class="btn-container">
        <button
          class={status && status !== "PLANNED" ? "not-active" : ""}
          on:click={() => wBtnClicked("PLANNED")}><Icon i="calendar" /></button
        >
        <button
          class={status && status !== "WATCHING" ? "not-active" : ""}
          on:click={() => wBtnClicked("WATCHING")}><Icon i="clock" /></button
        >
        <button
          class={status && status !== "FINISHED" ? "not-active" : ""}
          on:click={() => wBtnClicked("FINISHED")}><Icon i="check" /></button
        >
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
    position: relative;
    // aspect-ratio: 2/3;
    transition: all 150ms ease-in;

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

      .rating {
        display: flex;
        flex-flow: row-reverse;
        align-items: center;
        justify-content: center;
        color: white;
        -webkit-text-stroke: 1.5px white;
        cursor: pointer;
        overflow: hidden;
        margin: 10px 0 10px 0;

        button {
          font-size: 35px;
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
        }
      }

      .btn-container {
        display: flex;
        flex-flow: row;
        gap: 10px;
        margin-top: auto;

        button {
          font-size: 10px;

          &:hover {
            fill: white;
          }
        }
      }
    }

    &:hover {
      transform: scale(1.3);
    }

    &:hover {
      z-index: 99;

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
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
