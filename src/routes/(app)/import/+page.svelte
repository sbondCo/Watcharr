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
  import type {
    ImportedList,
    MovaryHistory,
    MovaryRatings,
    MovaryWatchlist,
    Watched
  } from "@/types";
  import { json } from "@sveltejs/kit";

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
      let type: "text-list" | "tmdb" = "text-list";
      if (file.type === "text/csv") type = "tmdb";
      r.addEventListener(
        "load",
        () => {
          if (r.result) {
            importedList.set({
              data: r.result.toString(),
              type
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

  /**
   * Process movary import files.
   *
   * Movary exports 3 different files:
   *
   *  - watchlist.csv = Planned movies.
   *  - history.csv   = Watched movies, movie can be watched multiple times.
   *  - ratings.csv   = Ratings for movies. One rating per movie.
   *
   * Export types explained better here:
   * https://github.com/sbondCo/Watcharr/issues/332#issuecomment-1920662244
   */
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
      // Build toImport array
      const toImport: ImportedList[] = [];
      // Add all history movies (watched). There can be multiple entries for each movie.
      for (let i = 0; i < historyJson.data.length; i++) {
        const h = historyJson.data[i];
        // Skip if no tmdb id.
        if (!h.tmdbId) {
          continue;
        }
        // Skip if already added. The first time it is added we get all info needed from other entries.
        if (toImport.filter((ti) => ti.tmdbId == Number(h.tmdbId)).length > 0) {
          continue;
        }
        const ratingsEntry = ratingsJson.data.find((r) => r.tmdbId == h.tmdbId);
        const t: ImportedList = {
          name: h.title,
          tmdbId: Number(h.tmdbId),
          status: "FINISHED",
          type: "movie", // movary only supports movies
          datesWatched: [],
          thoughts: ""
        };
        // Movie can be watched more than once, get all entries to store all watch dates.
        const allEntries = historyJson.data.filter((he) => he.tmdbId === h.tmdbId);
        for (let i = 0; i < allEntries.length; i++) {
          const e = allEntries[i];
          if (e.watchedAt) {
            t.datesWatched?.push(new Date(e.watchedAt));
          }
          if (e.comment) {
            t.thoughts += e.comment + "\n";
          }
        }
        if (h.year) {
          t.year = h.year;
        }
        if (ratingsEntry && ratingsEntry?.userRating) {
          t.rating = Number(ratingsEntry.userRating);
        }
        toImport.push(t);
      }
      // Add all watchlist movies (planned).
      for (let i = 0; i < watchlistJson.data.length; i++) {
        const wl = watchlistJson.data[i];
        const existing = toImport.find((ti) => ti.tmdbId == Number(wl.tmdbId));
        // If already exists in toImport, simply update status to PLANNED.
        // The movie must have been completed in past, but added back to
        // the users movary watch list as they are planning to watch it again.
        if (existing) {
          existing.status = "PLANNED";
          continue;
        }
        toImport.push({
          name: wl.title,
          tmdbId: Number(wl.tmdbId),
          status: "PLANNED",
          type: "movie" // movary only supports movies
        });
      }
      console.log("toImport:", toImport);
      importedList.set({
        data: JSON.stringify(toImport),
        type: "movary"
      });
      goto("/import/process");
    } catch (err) {
      isLoading = false;
      notify({ type: "error", text: "Failed to read files!" });
      console.error("import: Failed to read files!", err);
    }
  }

  async function processWatcharrFile(files?: FileList | null) {
    try {
      console.log("processWatcharrFile", files);
      if (!files || files?.length <= 0) {
        console.error("processWatcharrFile", "No files to process!");
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
      if (file.type !== "application/json") {
        notify({
          type: "error",
          text: "Must be a Watcharr JSON export file"
        });
        isLoading = false;
        isDragOver = false;
        return;
      }
      // Build toImport array
      const toImport: ImportedList[] = [];
      const fileText = await readFile(new FileReader(), file);
      const jsonData = JSON.parse(fileText) as Watched[];
      for (const v of jsonData) {
        if (!v.content || !v.content.title) {
          notify({
            type: "error",
            text: "Item in export has no content or a missing title! Look in console for more details."
          });
          console.error(
            "Can't add export item to import table! It has no content or a missing content.title! Item:",
            v
          );
          continue;
        }
        const t: ImportedList = {
          tmdbId: v.content.tmdbId,
          name: v.content.title,
          year: new Date(v.content.release_date)?.getFullYear()?.toString(),
          type: v.content.type,
          rating: v.rating,
          status: v.status,
          thoughts: v.thoughts,
          // datesWatched: [new Date(v.createdAt)], // Shouldn't need this, all activity will be imported, including ADDED_WATCHED activity
          activity: v.activity
        };
        toImport.push(t);
      }
      console.log("toImport:", toImport);
      importedList.set({
        data: JSON.stringify(toImport),
        type: "watcharr"
      });
      goto("/import/process");
    } catch (err) {
      isLoading = false;
      notify({ type: "error", text: "Failed to read file!" });
      console.error("import: Failed to read file!", err);
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

        <DropFileButton text="Watcharr Exports" filesSelected={(f) => processWatcharrFile(f)} />
      {/if}
    </div>
  </div>
</div>

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;
    padding: 0 30px 30px 30px;

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
