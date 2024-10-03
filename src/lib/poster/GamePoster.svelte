<script lang="ts">
  import type { PosterExtraDetails, Image, WatchedStatus } from "@/types";
  import {
    addClassToParent,
    calculateTransformOrigin,
    getOrdinalSuffix,
    isTouch,
    monthsShort
  } from "@/lib/util/helpers";
  import { goto } from "$app/navigation";
  import { baseURL, removeWatched, updatePlayed } from "../util/api";
  import { notify } from "../util/notify";
  import { onMount } from "svelte";
  import PosterStatus from "./PosterStatus.svelte";
  import PosterRating from "./PosterRating.svelte";
  import { wlDetailedView } from "@/store";
  import { decode } from "blurhash";
  import ExtraDetails from "./ExtraDetails.svelte";

  export let id: number | undefined = undefined; // Watched list id
  export let media: {
    id: number;
    coverId: string;
    firstReleaseDate?: string | number;
    name: string;
    summary?: string;
    poster?: Image;
  };
  export let rating: number | undefined = undefined;
  export let status: WatchedStatus | undefined = undefined;
  export let small = false;
  export let disableInteraction = false;
  export let hideButtons = false;
  export let extraDetails: PosterExtraDetails | undefined = undefined;
  export let fluidSize = false;
  export let pinned = false;
  // When provided, default click handlers will instead run this callback.
  export let onClick: (() => void) | undefined = undefined;

  $: dve = $wlDetailedView;

  // If poster is active (scaled up)
  let posterActive = false;

  let containerEl: HTMLDivElement;
  let bhCanvas: HTMLCanvasElement;

  const title = `${media.name}`;
  const poster = media.poster?.path
    ? `${baseURL}/${media.poster.path}`
    : `https://images.igdb.com/igdb/image/upload/t_cover_big/${media.coverId}.jpg`;
  const link = media.id ? `/game/${media.id}` : undefined;
  const dateStr = media.firstReleaseDate;
  const year = dateStr ? new Date(dateStr).getFullYear() : undefined;

  function handleStarClick(r: number) {
    if (r == rating) return;
    updatePlayed(media.id, undefined, r);
  }

  function handleStatusClick(type: WatchedStatus | "DELETE") {
    if (type === "DELETE") {
      if (!id) {
        notify({ text: "Content has no watched list id, can't delete.", type: "error" });
        return;
      }
      removeWatched(id);
      return;
    }
    if (type == status) return;
    updatePlayed(media.id, type);
  }

  function handleInnerKeyUp(e: KeyboardEvent) {
    console.log(e.target);
    if (e.key === "Enter" && (e.target as HTMLElement)?.id === "ilikemoviessueme") {
      if (typeof onClick !== "undefined") {
        onClick();
        return;
      }
      if (link) {
        goto(link);
      }
    }
  }

  function formatDate(e: number) {
    if (!e) {
      return "Unknown";
    }
    const d = new Date(e);
    return `${d.getDate()}${getOrdinalSuffix(d.getDate())} ${monthsShort[d.getMonth()]} '${String(
      d.getFullYear()
    ).substring(2, 4)}`;
  }

  $: {
    if (media.poster?.path && media.poster?.blurHash && bhCanvas) {
      const pixels = decode(media.poster.blurHash, 170, 256);
      const ctx = bhCanvas.getContext("2d");
      if (ctx) {
        const imageData = ctx.createImageData(170, 256);
        imageData.data.set(pixels);
        ctx.putImageData(imageData, 0, 0);
      }
    }
  }

  onMount(() => {
    if (containerEl) {
      if (small) {
        containerEl.classList.add("small");
      }
      if (fluidSize) {
        containerEl.classList.add("fluid-size");
      }
    }
  });
</script>

<!-- HACK: disabled this issue for now, it should probably be fixed properly -->
<!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
<li
  on:mouseenter={(e) => {
    if (!posterActive) calculateTransformOrigin(e);
    if (!isTouch()) {
      posterActive = true;
    }
  }}
  on:focusin={(e) => {
    if (!posterActive) calculateTransformOrigin(e);
    if (!isTouch()) {
      posterActive = true;
    }
  }}
  on:focusout={() => {
    if (!isTouch()) {
      // Only on !isTouch (to match focusin) to avoid breaking a tap and hold on link on mobile.
      posterActive = false;
    }
  }}
  on:mouseleave={() => {
    posterActive = false;
    const ae = document.activeElement;
    if (
      ae &&
      ae instanceof HTMLElement &&
      (ae.parentElement?.id === "ilikemoviessueme" ||
        ae.parentElement?.parentElement?.id === "ilikemoviessueme")
    ) {
      // Stops the poster being re-focused after the browser window
      // loses focus, then regains it (ex: you middle click the poster,
      // go to the opened tab (or lose browser window focus, then when
      // you come back the poster is sent `focusin` and stuck activated
      // until mouseleave again).
      ae.blur();
    }
  }}
  on:click={() => (posterActive = true)}
  on:keyup={(e) => {
    if (e.key === "Tab") {
      e.currentTarget.scrollIntoView({ block: "center" });
    }
  }}
  on:keypress={() => console.log("on kpress")}
  class={`${posterActive ? "active " : ""}${pinned ? "pinned " : ""}`}
>
  <div
    class={`container${!poster || (!media.coverId && !media.poster?.path) ? " details-shown" : ""}`}
    bind:this={containerEl}
  >
    {#if poster && (media.coverId || media.poster?.path)}
      {#if media?.poster?.blurHash}
        <canvas width="170" height="256" bind:this={bhCanvas} class="img-loader" />
      {:else}
        <div class="img-loader"></div>
      {/if}
      <img
        loading="lazy"
        src={poster}
        alt=""
        on:load={(e) => {
          addClassToParent(e, "img-loaded");
        }}
        on:error={(e) => {
          addClassToParent(e, "details-shown");
        }}
      />
    {/if}
    {#if id && !posterActive}
      <!-- Must be on watched list, and poster not hovered -->
      <ExtraDetails details={extraDetails} {status} {rating} />
    {/if}
    <div
      on:click={(e) => {
        if (typeof onClick !== "undefined") {
          onClick();
          // Prevent the link inside this div from being clicked in this case.
          e.preventDefault();
          return;
        }
        if (posterActive && link) goto(link);
      }}
      on:keyup={handleInnerKeyUp}
      id="ilikemoviessueme"
      class="inner"
      role="button"
      tabindex="-1"
    >
      <a data-sveltekit-preload-data="tap" href={link}>
        <h2>
          {title}
          {#if year}
            <time>{year}</time>
          {/if}
        </h2>
        <span>{media.summary}</span>
      </a>

      {#if !hideButtons}
        <div class="buttons">
          <PosterRating {rating} {handleStarClick} {disableInteraction} />
          <PosterStatus {status} {handleStatusClick} {disableInteraction} isForGame={true} />
        </div>
      {/if}
    </div>
  </div>
</li>

<style lang="scss">
  li.active {
    cursor: pointer;
  }

  li.pinned:not(.active) .container {
    outline: 3px solid gold;
  }

  li {
    &:not(.active) {
      .container .inner,
      .container .inner .buttons {
        pointer-events: none !important;
      }
    }
  }

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
    transition: transform 150ms ease;

    &.fluid-size {
      height: 100%;
      width: 100%;
    }

    img {
      width: 100%;
      height: 100%;
    }

    &.img-loaded .img-loader,
    &.details-shown .img-loader {
      display: none;
    }

    .img-loader {
      position: absolute;
      width: 100%;
      height: 100%;
    }

    div.img-loader {
      background-color: gray;
      background: linear-gradient(359deg, #5c5c5c, #2c2929, #2c2424);
      background-size: 400% 400%;
      animation: imgloader 4s ease infinite;

      @keyframes imgloader {
        0% {
          background-position: 50% 0%;
        }
        50% {
          background-position: 50% 100%;
        }
        100% {
          background-position: 50% 0%;
        }
      }
    }

    canvas.img-loader {
      animation: cimgloader 4s ease infinite;

      @keyframes cimgloader {
        0% {
          opacity: 1;
        }
        50% {
          opacity: 0.7;
        }
        100% {
          opacity: 1;
        }
      }
    }

    .inner {
      position: absolute;
      opacity: 0;
      display: flex;
      flex-flow: column;
      top: 0;
      height: 100%;
      width: 100%;
      padding: 10px;
      background-color: transparent;
      transition: opacity 150ms cubic-bezier(0.19, 1, 0.22, 1);

      & > a {
        height: 100%;
      }

      h2 {
        font-family:
          sans-serif,
          system-ui,
          -apple-system,
          BlinkMacSystemFont;
        font-size: 18px;
        color: white;
        word-wrap: break-word;

        a {
          color: white;
        }

        time {
          font-size: 14px;
          font-weight: 400;
          color: rgba(255, 255, 255, 0.7);
        }
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
      }
    }

    &.small .inner span {
      font-size: 11px;
    }

    .active & {
      transform: scale(1.3);
      z-index: 99;
    }

    .active &.small {
      transform: scale(1.1);
    }

    .active &,
    &:global(.details-shown) {
      img {
        filter: blur(4px) grayscale(80%);
        // This makes the background very dark,
        // but atleast the text is visible.. may want to change later.
        mix-blend-mode: multiply;
      }

      .inner {
        color: white;
        opacity: 1;
      }
    }
  }
</style>
