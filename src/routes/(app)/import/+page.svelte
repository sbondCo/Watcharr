<!-- 
  /import is for getting the user to select
  the file they want to import and reading
  it. The data is set in a store for
  /import/process to process.
 -->

<script lang="ts">
  import { goto } from "$app/navigation";
  import Icon from "@/lib/Icon.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { notify } from "@/lib/util/notify";
  import { importedList } from "@/store";
  import { onMount } from "svelte";

  let fileInput: HTMLInputElement;
  let dragEnterTarget: EventTarget | null;
  let isDragOver = false;
  let isLoading = false;

  function processFiles(files?: FileList | null) {
    try {
      console.log("processFiles", files);
      if (!files || files?.length <= 0) {
        console.error("processFiles", "No files to process!");
        notify({
          type: "error",
          text: "File not found in dropped items."
        });
        isDragOver = false;
        return;
      }
      isLoading = true;
      if (files.length > 1) {
        notify({
          type: "error",
          text: "Only one file at a time is supported. Continuing with the first.",
          time: 6000
        });
      }
      // Currently only support for importing one file at a time
      const file = files[0];
      if (file.type !== "text/plain" && file.type !== "text/csv") {
        notify({
          type: "error",
          text: "Currently only text and csv (TMDb export) files are supported"
        });
        isLoading = false;
        isDragOver = false;
        return;
      }
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
    } catch (err) {
      isLoading = false;
      notify({ type: "error", text: "Failed to read file!" });
      console.error("import: Failed to read file!", err);
    }
  }

  function importFile() {
    fileInput.click();
  }

  onMount(() => {
    if (!localStorage.getItem("token")) {
      goto("/login");
    }
    if (fileInput) {
      fileInput.addEventListener("change", (ev) => {
        processFiles(fileInput.files);
      });
    }
  });
</script>

<div class="content">
  <div class="inner">
    <span class="header">
      <h2>Import Your Watchlist</h2>
      <h5 class="norm">beta</h5>
    </span>
    <div class="big-btns">
      {#if isLoading}
        <Spinner />
      {:else}
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
      {/if}
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

    .header {
      display: flex;
      gap: 10px;

      h5 {
        margin-top: 3px;
      }
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
