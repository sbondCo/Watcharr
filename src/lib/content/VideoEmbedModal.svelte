<script lang="ts">
  import { onMount } from "svelte";

  export let embed: string;
  export let closed: () => void;

  let modalDiv: HTMLDivElement;

  onMount(() => {
    // For better experience on keyboard.
    // If modal shown by btn click, the btn will still be
    // focused, so clicking `esc` wont trigger until tab
    // is clicked first... which I count as not intuitive.
    if (modalDiv) {
      modalDiv.focus();
    }
  });

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      closed();
    }
  }
</script>

<div
  bind:this={modalDiv}
  class="modal"
  on:click={closed}
  on:keydown={handleKeyDown}
  role="button"
  tabindex="0"
>
  {#if embed}
    <div class="wrapper">
      <iframe title="Video Embed" src={embed} frameborder="0" width="100%" height="100%" />
    </div>
  {/if}
</div>

<style lang="scss">
  .modal {
    display: flex;
    align-items: center;
    justify-content: center;
    position: fixed;
    top: 0;
    left: 0;
    width: 100dvw;
    height: 100dvh;
    background-color: rgba($color: #000000, $alpha: 0.5);
    z-index: 99998;

    .wrapper {
      width: 100%;
      max-width: 800px;
      border-radius: 15px;
      aspect-ratio: 16 / 9;
      overflow: hidden;
      margin: 20px;
    }
  }
</style>
