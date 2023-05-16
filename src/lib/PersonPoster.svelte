<script lang="ts">
  import type { MediaType, WatchedStatus } from "@/types";
  import Icon from "./Icon.svelte";
  import { isTouch, watchedStatuses } from "./helpers";
  import { goto } from "$app/navigation";
  import tooltip from "./actions/tooltip";

  export let id: number | undefined;
  export let name: string | undefined;
  export let path: string | undefined;

  const poster = path ? `https://image.tmdb.org/t/p/w300_and_h450_bestv2${path}` : undefined;
  const link = id ? `/person/${id}` : undefined;

  function addClassToParent(
    e: Event & {
      currentTarget: EventTarget & Element;
    },
    c: string
  ) {
    (e.currentTarget?.parentNode as HTMLDivElement)?.classList.add(c);
  }
</script>

<li
  on:click={() => {
    if (link) goto(link);
  }}
  on:keypress={() => console.log("on kpress")}
>
  <div class={`container${!poster ? " details-shown" : ""}`}>
    {#if poster}
      <div class="img-loader" />
      <img
        loading="lazy"
        src={poster}
        alt=""
        on:load={(e) => {
          console.log("on lod");
          addClassToParent(e, "img-loaded");
        }}
        on:error={(e) => {
          console.log("on err");
          addClassToParent(e, "details-shown");
        }}
      />
    {/if}
    <div class="inner">
      <h2>
        {#if link}
          <a data-sveltekit-preload-data="tap" href={link}>
            {name}
          </a>
        {:else}
          {name}
        {/if}
      </h2>
    </div>
  </div>
</li>

<style lang="scss">
  li.active {
    cursor: pointer;
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
    transition: transform 150ms ease;
    cursor: pointer;

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
      background-color: gray;
      background: linear-gradient(359deg, #5c5c5c, #2c2929, #2c2424);
      background-size: 400% 400%;
      animation: imgloader 4s ease infinite;

      @-webkit-keyframes imgloader {
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

    .inner {
      position: absolute;
      display: flex;
      flex-flow: column;
      top: 0;
      height: 100%;
      width: 100%;
      padding: 10px;
      background-color: transparent;

      h2 {
        font-family: sans-serif, system-ui, -apple-system, BlinkMacSystemFont;
        font-size: 18px;
        color: white;
        margin-top: auto;
      }
    }

    &:hover,
    &:focus-within {
      transform: scale(1.3);
      z-index: 99;
    }

    &:hover,
    &:focus-within,
    &:global(.details-shown) {
      .inner {
        color: white;
      }
    }
  }
</style>
