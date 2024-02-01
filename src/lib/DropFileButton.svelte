<script lang="ts">
  import { onMount } from "svelte";
  import type { Icon as IconT } from "@/types";
  import Icon from "./Icon.svelte";

  export let text: string;
  export let icon: IconT = "document";
  export let filesSelected: (f?: FileList | null) => void;
  export let allowSelectMultipleFiles = false;

  let fileInput: HTMLInputElement;
  let dragEnterTarget: EventTarget | null;
  let isDragOver = false;

  function importFile() {
    fileInput.click();
  }

  onMount(() => {
    if (fileInput) {
      fileInput.addEventListener("change", () => {
        filesSelected(fileInput.files);
      });
    }
  });
</script>

<div class="drop-file-btn">
  <button
    on:click={importFile}
    on:dragover={(ev) => {
      ev.preventDefault();
      ev.stopPropagation();
    }}
    on:dragenter={(ev) => {
      ev.preventDefault();
      ev.stopPropagation();
      dragEnterTarget = ev.target;
      console.log("enter");
      isDragOver = true;
    }}
    on:dragleave={(ev) => {
      ev.preventDefault();
      ev.stopPropagation();
      if (dragEnterTarget === ev.target) {
        console.log("leave");
        isDragOver = false;
      }
    }}
    on:drop={(ev) => {
      ev.preventDefault();
      ev.stopPropagation();
      filesSelected(ev.dataTransfer?.files);
    }}
    class={isDragOver ? "dragging-over" : ""}
  >
    <Icon i={isDragOver ? "add" : icon} wh="100%" />
    <div>
      <h4 class="norm">
        {#if isDragOver}
          Import {text}
        {:else}
          Browse For {text}
        {/if}
      </h4>
    </div>
  </button>
  <input type="file" multiple={allowSelectMultipleFiles} bind:this={fileInput} />
</div>

<style lang="scss">
  .drop-file-btn {
    button {
      display: flex;
      flex-flow: column;
      justify-content: center;
      align-items: center;
      gap: 10px;
      height: 180px;
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

      &:hover,
      &.dragging-over {
        color: $bg-color;
        background-color: $accent-color-hover;

        :global(#reel path:last-of-type) {
          fill: $bg-color;
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
