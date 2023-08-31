<script lang="ts">
  import Icon from "@/lib/Icon.svelte";
  import { onMount } from "svelte";

  let fileInput: HTMLInputElement;

  function fileInputChange(e: Event) {
    // const ev = e as InputEvent;
    console.log(fileInput.files);
  }

  function importText() {
    fileInput.click();
  }

  onMount(() => {
    if (fileInput) {
      fileInput.addEventListener("change", fileInputChange);
    }
  });
</script>

<div class="content">
  <h2>Import Your Watchlist</h2>
  <div class="big-btns">
    <button on:click={importText}>
      <Icon i="document" />
      <h4 class="norm">Text File</h4>
    </button>
    <button>
      <Icon i="reel" wh="100%" />
      <h4 class="norm">Watcharr</h4>
    </button>
  </div>
  <input type="file" bind:this={fileInput} />
</div>

<style lang="scss">
  .content {
    display: flex;
    flex-flow: column;
    width: 100%;
    justify-content: center;
    padding: 0 30px 0 30px;

    .big-btns {
      display: flex;
      flex-flow: row;
      gap: 20px;

      button {
        display: flex;
        flex-flow: column;
        justify-content: center;
        align-items: center;
        gap: 10px;
        width: 150px;
        padding: 20px;
        background-color: $accent-color;
        border: unset;
        border-radius: 10px;
        user-select: none;
        transition: 180ms ease-in-out;

        :global {
          #reel path {
            transition: 180ms ease-in-out;

            &:first-of-type {
              fill: transparent;
            }

            &:last-of-type {
              fill: $text-color;
            }
          }
        }

        &:hover {
          color: $bg-color;
          background-color: $accent-color-hover;

          :global(#reel path:last-of-type) {
            fill: $bg-color;
          }
        }
      }
    }

    input[type="file"] {
      width: 0px;
      overflow: hidden;
      border: unset;
      background-color: transparent;
      position: absolute;
      top: -500px;
    }
  }
</style>
