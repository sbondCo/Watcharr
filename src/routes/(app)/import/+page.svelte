<!-- 
  /import is for getting the user to select
  the file they want to import and reading
  it. The data is set in a store for
  /import/process to process.
 -->

<script lang="ts">
  import { goto } from "$app/navigation";
  import DropFileButton from "@/lib/DropFileButton.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { notify } from "@/lib/util/notify";
  import { importedList } from "@/store";
  import { onMount } from "svelte";
  import papa from "papaparse";

  let isDragOver = false;
  let isLoading = false;

  function processFiles(files?: FileList | null) {
    try {
      console.log("processFiles", files);
      if (!files || files?.length <= 0) {
        console.error("processFiles", "No files to process!");
        notify({
          type: "error",
          text: "File not found in dropped items. Please try again or refresh.",
          time: 6000
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

  async function readFile(fr: FileReader, file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const res = () => {
        fr.removeEventListener("load", res);
        fr.removeEventListener("error", rej);
        if (fr.result) {
          resolve(fr.result.toString());
        } else {
          reject("no result");
        }
      };
      const rej = () => {
        fr.removeEventListener("load", res);
        fr.removeEventListener("error", rej);
        reject();
      };
      fr.addEventListener("load", res);
      fr.addEventListener("error", rej);
      fr.readAsText(file);
    });
  }

  interface MovaryExportBase {
    title: string;
    year: string;
    tmdbId: string;
    imdbId: string;
  }

  interface MovaryHistory extends MovaryExportBase {
    watchedAt: string;
    comment: string;
  }

  interface MovaryRatings extends MovaryExportBase {
    userRating: string;
  }

  interface MovaryWatchlist extends MovaryExportBase {
    addedAt: string;
  }

  async function processFilesMovary(files?: FileList | null) {
    try {
      console.log("processFilesMovary", files);
      if (!files || files?.length <= 0) {
        console.error("processFilesMovary", "No files to process!");
        notify({
          type: "error",
          text: "File not found in dropped items. Please try again or refresh.",
          time: 6000
        });
        isDragOver = false;
        return;
      }
      if (files.length !== 3) {
        notify({
          type: "error",
          text: "You must select or drop 3 files: history.csv, ratings.csv and watchlist.csv.",
          time: 6000
        });
        isDragOver = false;
        return;
      }
      isLoading = true;
      // Read file data into strings
      let history: string | undefined;
      let ratings: string | undefined;
      let watchlist: string | undefined;
      const r = new FileReader();
      for (let i = 0; i < files.length; i++) {
        const f = files[i];
        if (f.name === "history.csv") {
          history = await readFile(r, f);
        } else if (f.name === "ratings.csv") {
          ratings = await readFile(r, f);
        } else if (f.name === "watchlist.csv") {
          watchlist = await readFile(r, f);
        }
      }
      if (!history || !ratings || !watchlist) {
        notify({
          type: "error",
          text: "Failed to read history, ratings or watchlist. Ensure you have attached 3 files: history.csv, ratings.csv and watchlist.csv.",
          time: 6000
        });
        isDragOver = false;
        isLoading = false;
        return;
      }
      console.log("loaded all files");
      // Convert csv strings into json
      const historyJson = papa.parse<MovaryHistory>(history.trim(), { header: true });
      const ratingsJson = papa.parse<MovaryRatings>(ratings.trim(), { header: true });
      const watchlistJson = papa.parse<MovaryWatchlist>(watchlist.trim(), { header: true });
      for (let i = 0; i < watchlistJson.data.length; i++) {
        const wl = watchlistJson.data[i];
        const historyEntry = historyJson.data.find(
          (h) => h.title == wl.title && h.tmdbId == wl.tmdbId
        );
        const ratingsEntry = ratingsJson.data.find(
          (r) => r.title == wl.title && r.tmdbId == wl.tmdbId
        );
        console.log(wl, historyEntry, ratingsEntry);
      }
    } catch (err) {
      isLoading = false;
      notify({ type: "error", text: "Failed to read files!" });
      console.error("import: Failed to read files!", err);
    }
  }

  onMount(() => {
    if (!localStorage.getItem("token")) {
      goto("/login");
    }
  });
</script>

<div class="content">
  <div class="inner">
    <div>
      <span class="header">
        <h2>Import Your Watchlist</h2>
        <h5 class="norm">beta</h5>
      </span>
      <!-- <h4 class="norm">Currently txt and csv (TMDb export) files are supported.</h4> -->
    </div>
    <div class="big-btns">
      {#if isLoading}
        <Spinner />
      {:else}
        <DropFileButton
          text=".txt list or .csv TMDb Export"
          filesSelected={(f) => processFiles(f)}
        />

        <DropFileButton
          icon="movary"
          text="Movary Exports"
          filesSelected={(f) => processFilesMovary(f)}
          allowSelectMultipleFiles
        />
      {/if}
    </div>
  </div>
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
      flex-flow: column;
      gap: 20px;
      margin-top: 20px;
    }

    .header {
      display: flex;
      gap: 10px;

      h5 {
        margin-top: 3px;
      }
    }
  }
</style>
