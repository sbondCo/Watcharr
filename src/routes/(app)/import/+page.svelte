<!-- 
  /import is for getting the user to select
  the file they want to import and reading
  it. The data is set in a store for
  /import/process to process.
 -->

<script lang="ts">
  import { goto } from "$app/navigation";
  import Icon from "@/lib/Icon.svelte";
  import { importedList } from "@/store";
  import { onMount } from "svelte";

  let fileInput: HTMLInputElement;
  let dragEnterTarget: EventTarget | null;
  let isDragOver = false;

  function processFiles(files?: FileList | null) {
    console.log("processFiles", files);
    if (!files) {
      console.error("processFiles", "No files to process!");
      return;
    }
    // Currently only support for importing one file at a time
    const file = files[0];
    const r = new FileReader();
    r.addEventListener(
      "load",
      () => {
        if (r.result) {
          importedList.set({
            file,
            data: r.result.toString()
          });
          goto("/import/process");
        }
      },
      false
    );
    r.readAsText(file);
  }

  function importFile() {
    fileInput.click();
  }

  onMount(() => {
    if (fileInput) {
      fileInput.addEventListener("change", (ev) => {
        processFiles(fileInput.files);
      });
    }
  });
</script>

<div class="content">
  <div class="inner">
    <h2>Import Your Watchlist</h2>
    <div class="big-btns">
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
          processFiles(ev.dataTransfer?.files);
        }}
        class={isDragOver ? "dragging-over" : ""}
      >
        <Icon i={isDragOver ? "add" : "document"} wh="100%" />
        <div>
          <h4 class="norm">
            {#if isDragOver}
              Import
            {:else}
              Browse
            {/if}
          </h4>
          <!-- <h5 class="norm">Or Drag And Drop</h5> -->
        </div>
      </button>
    </div>
  </div>
  <input type="file" bind:this={fileInput} />
</div>

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;
    padding: 0 30px 0 30px;

    .inner {
      display: flex;
      flex-flow: column;
      min-width: 400px;
      max-width: 400px;
      overflow: hidden;
    }

    .big-btns {
      display: flex;
      justify-content: center;
      flex-flow: row;
      gap: 20px;
      margin-top: 20px;

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

        &:hover,
        &.dragging-over {
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