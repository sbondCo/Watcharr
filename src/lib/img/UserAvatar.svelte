<script lang="ts">
  import { decode } from "blurhash";
  import Icon from "../Icon.svelte";
  import { baseURL } from "../util/api";
  import { onMount } from "svelte";
  import type { Image } from "@/types";

  /**
   * Users avatar.
   */
  export let img: Image | undefined;
  export let avatarDropped: ((ev: Event) => void) | undefined = undefined;

  let bhCanvas: HTMLCanvasElement;
  let avatarInput: HTMLInputElement;

  function avatarLoaded() {
    console.log("avatar loaded.. removing canvas");
    bhCanvas?.remove();
  }

  $: {
    if (img?.path && img?.blurHash && bhCanvas) {
      const pixels = decode(img.blurHash, 80, 80);
      const ctx = bhCanvas.getContext("2d");
      if (ctx) {
        const imageData = ctx.createImageData(80, 80);
        imageData.data.set(pixels);
        ctx.putImageData(imageData, 0, 0);
      }
    }
  }

  onMount(() => {
    // Ignore rest if avatarDropped not defined
    if (typeof avatarDropped !== "function") return;

    avatarInput?.addEventListener("input", avatarDropped);

    return () => {
      avatarInput?.removeEventListener("input", avatarDropped!);
    };
  });
</script>

<div class={["img-ctr", typeof avatarDropped === "function" ? "" : "no-click"].join(" ")}>
  {#if img?.path}
    <canvas bind:this={bhCanvas} />
    <img src={`${baseURL}/${img.path}`} alt="" on:load={avatarLoaded} />
  {:else}
    <Icon i="person" wh="100%" />
  {/if}
  {#if typeof avatarDropped === "function"}
    <input bind:this={avatarInput} type="file" title="" accept=".jpg,.png,.gif,.webp" />
  {/if}
</div>

<style lang="scss">
  .img-ctr {
    width: 80px;
    min-width: 80px;
    height: 80px;
    min-height: 80px;
    border-radius: 50%;
    position: relative;
    overflow: hidden;

    img {
      width: 80px;
      min-width: 80px;
      height: 80px;
      min-height: 80px;
      object-fit: cover;
    }

    canvas {
      position: absolute;
      left: 0;
    }

    &:not(.no-click):hover {
      opacity: 0.8;
    }

    input[type="file"] {
      opacity: 0;
      width: 100%;
      height: 100%;
      left: 0;
      position: absolute;
      cursor: pointer;
    }
  }
</style>
