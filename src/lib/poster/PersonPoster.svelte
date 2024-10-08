<script lang="ts">
  import { goto } from "$app/navigation";
  import { addClassToParent, calculateTransformOrigin } from "@/lib/util/helpers";

  export let id: number | undefined;
  export let name: string | undefined;
  export let path: string | undefined;
  export let role: string | undefined = undefined;
  export let zoomOnHover: boolean = true;

  const poster = path ? `https://image.tmdb.org/t/p/w300_and_h450_bestv2${path}` : undefined;
  const link = id ? `/person/${id}` : undefined;
</script>

<!-- Quick fix to ignore error, should be fixed -->
<!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
<li
  on:mouseenter={(e) => calculateTransformOrigin(e)}
  on:focusin={(e) => calculateTransformOrigin(e)}
  on:click={() => {
    if (link) goto(link);
  }}
  on:keypress={() => console.log("on kpress")}
>
  <div class={`container${!poster ? " details-shown" : ""}${!zoomOnHover ? " no-zoom" : ""}`}>
    {#if poster}
      <div class="img-loader" />
      <img
        loading="lazy"
        src={poster}
        alt=""
        on:load={(e) => {
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
      {#if role}
        <h3>{role}</h3>
      {/if}
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
    transition:
      transform 150ms ease,
      outline 50ms ease;
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

    .inner {
      position: absolute;
      display: flex;
      flex-flow: column;
      top: 0;
      height: 100%;
      width: 100%;
      padding: 10px;
      background-color: transparent;

      h2,
      h3 {
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
      }

      h2 {
        margin-top: auto;
        text-shadow: 1px 1px 3px black;
      }

      h3 {
        font-size: 12px;
        text-shadow: 0px 1px 1px black;
      }
    }

    &:not(.no-zoom) {
      &:hover,
      &:has(:focus-visible) {
        transform: scale(1.3);
        z-index: 99;
      }
    }

    &:not(:not(.no-zoom)) {
      &:hover,
      &:has(:focus-visible) {
        outline: 3px solid $text-color;
      }
    }

    &:hover,
    &:has(:focus-visible),
    &:global(.details-shown) {
      .inner {
        color: white;
      }
    }
  }
</style>
