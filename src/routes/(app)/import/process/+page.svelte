<!-- 
  /import/process is for processing the
  selected files data. Here it will be
  displayed and imported.
 -->

<script lang="ts">
  import { goto } from "$app/navigation";
  import DropDown from "@/lib/DropDown.svelte";
  import Error from "@/lib/Error.svelte";
  import Icon from "@/lib/Icon.svelte";
  import Modal from "@/lib/Modal.svelte";
  import Poster from "@/lib/poster/Poster.svelte";
  import PosterList from "@/lib/poster/PosterList.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
  import { sleep } from "@/lib/util/helpers";
  import { notify } from "@/lib/util/notify";
  import { importedList, parsedImportedList, watchedList } from "@/store";
  import {
    ImportResponseType,
    type ImportResponse,
    type ContentSearchTv,
    type ContentSearchMovie,
    type ImportedList
  } from "@/types";
  import axios from "axios";
  import { onDestroy } from "svelte";
  import { get } from "svelte/store";
  import papa from "papaparse";

  const wList = get(watchedList);

  interface ImportedListItemMultiProblem {
    original: ImportedList;
    results: (ContentSearchMovie | ContentSearchTv)[];
    callback: (err: Error | string | undefined) => void;
  }

  let rList: ImportedList[] = [];
  let isImporting = false;
  let cancelled = false;
  let importText = "";

  onDestroy(() => {
    cancelled = true;
  });

  // Set when current item being imported gets an IMPORT_MULTI
  // response, which then shows the modal for user to pick correct item.
  let importMultiItem: ImportedListItemMultiProblem | undefined;

  async function getList() {
    const list = get(importedList);
    if (!list) {
      console.log("import/process, no list, returning to /import");
      goto("/import");
      return;
    }
    console.log("getList", list);
    if (list?.type === "text-list") {
      importText = "Text List";
      // Regex to match a year in between brackets,
      // which we assume is the release year of content.
      const yearRegex = new RegExp(/\([0-9]{4}\)/);
      const s = list.data.split("\n");
      for (let i = 0; i < s.length; i++) {
        const el = s[i]?.trim();
        if (el) {
          const l: ImportedList = { name: el };
          const year = el.match(yearRegex);
          if (year && year.length > 0) {
            l.year = year[0].replaceAll(/\(|\)/g, "");
            l.name = l.name.replace(yearRegex, "").trim();
          }
          rList.push(l);
        }
      }
    } else if (list?.type === "tmdb") {
      importText = "TMDB";
      const s = papa.parse(list.data.trim(), { header: true });
      console.debug("parsed csv", s);
      for (let i = 0; i < s.data.length; i++) {
        try {
          const el = s.data[i] as any;
          if (el) {
            // Skip if no name or tmdb id
            if (!el.Name && !el["TMDb ID"]) {
              console.warn("Skipping item with no name or tmdb id", el);
              return;
            }
            const l: ImportedList = { name: el.Name };
            const year = el["Release Date"] ? new Date(el["Release Date"]) : undefined;
            if (year) {
              l.year = String(year.getFullYear());
            }
            if (el.Type === "movie" || el.Type === "tv") {
              l.type = el.Type;
            }
            if (el["TMDb ID"]) {
              l.tmdbId = Number(el["TMDb ID"]);
            }
            if (el["Your Rating"]) {
              l.rating = Math.floor(Number(el["Your Rating"]));
            }
            if (el["Date Rated"]) {
              l.ratingCustomDate = new Date(el["Date Rated"]);
            }
            rList.push(l);
          }
        } catch (err) {
          console.error("Failed to process an item!", err);
          notify({
            type: "error",
            text: "Failed to process an item!"
          });
        }
      }
    } else if (list?.type === "movary") {
      importText = "Movary";
      try {
        const s = JSON.parse(list.data);
        // Builds imported list in previous step for ease.
        rList = s;
      } catch (err) {
        console.error("Movary import processing failed!", err);
        notify({
          type: "error",
          text: "Processing failed!. Please report this issue if it persists."
        });
      }
    }
    // TODO: remove duplicate names in list
    return list;
  }

  function addRow(ev: FocusEvent & { currentTarget: EventTarget & HTMLInputElement }) {
    if (!ev.currentTarget.value) {
      return;
    }
    const lo = { name: ev.currentTarget.value } as ImportedList;
    const yearEl = document.getElementById("addYear") as HTMLInputElement;
    if (yearEl?.value) {
      lo.year = yearEl.value;
    }
    rList.push(lo);
    rList = rList;
    ev.currentTarget.value = "";
    yearEl.value = "";
  }

  function removeRow(l: ImportedList) {
    rList = rList.filter((r) => r.name !== l.name);
    rList = rList;
  }

  async function startImport() {
    console.log(rList);
    isImporting = true;
    window.scrollTo(0, 0);
    for (let i = 0; i < rList.length; i++) {
      if (cancelled) {
        notify({ type: "error", text: "Importing Cancelled" });
        return;
      }
      const li = rList[i];
      try {
        console.log("Importing", li);
        await doImport(li);
      } catch (err) {
        console.error("Failed to import item:", li, "reason:", err);
        notify({
          type: "error",
          text: "Failed to import an item! Check console for more info.",
          time: Infinity
        });
      }
      await sleep(1500);
    }
    importedList.set(undefined);
    if (
      rList.some(
        (i) =>
          i.state == ImportResponseType.IMPORT_FAILED ||
          i.state == ImportResponseType.IMPORT_NOTFOUND
      )
    ) {
      // Some items failed.. go to some-failed
      parsedImportedList.set(rList);
      goto("/import/some-failed");
    } else {
      notify({ type: "success", text: "All content successfully imported!" });
      goto("/");
    }
  }

  async function doImport(item: ImportedList) {
    if (!item.name?.trim()) {
      item.state = ImportResponseType.IMPORT_NOTFOUND;
      rList = rList;
      return;
    }
    const resp = await axios.post<ImportResponse>("/import", item);
    return new Promise((res, rej) => {
      if (resp.data.type === ImportResponseType.IMPORT_MULTI) {
        console.log("Import found multiple responses for content", resp.data);
        let results = resp.data.results;
        if (item.year) {
          results = results.sort((a, b) => {
            try {
              const ar = a.media_type === "movie" ? a.release_date : a.first_air_date;
              const ay = ar ? new Date(Date.parse(ar)).getFullYear() : undefined;
              const br = b.media_type === "movie" ? b.release_date : b.first_air_date;
              const by = br ? new Date(Date.parse(br)).getFullYear() : undefined;
              if (ay == item.year) return -1;
              else if (by == item.year) return 1;
            } catch (err) {
              console.error("doImport: results sort failed", err);
            }
            return 0;
          });
        }
        importMultiItem = {
          original: item,
          results: resp.data.results,
          callback: (err) => {
            if (err) {
              item.state = ImportResponseType.IMPORT_NOTFOUND;
              rList = rList;
              rej(err);
            } else {
              res(0);
            }
          }
        };
      } else if (resp.data.type === ImportResponseType.IMPORT_SUCCESS) {
        item.state = ImportResponseType.IMPORT_SUCCESS;
        const w = resp.data.watchedEntry;
        if (w) {
          const release =
            w.content.type === "movie" ? w.content.release_date : w.content.first_air_date;
          if (release) item.year = String(new Date(Date.parse(release)).getFullYear());
          item.type = w.content.type;
          wList.push(w);
          watchedList.update(() => wList);
        }
        rList = rList;
        res(0);
      } else {
        item.state = resp.data.type;
        rList = rList;
        res(0);
      }
    });
  }
</script>

{#await getList()}
  <Spinner />
{:then list}
  <div class="content">
    <div class="inner">
      {#if rList}
        <h2>Importing {importText ? `From ${importText}` : ""}</h2>
        <h5 class="norm">
          {#if !isImporting}
            Review your imported list and fix any problems.
          {:else}
            You can fix any failed imports when the process completes.
          {/if}
        </h5>
        <table class={isImporting ? "is-importing" : ""}>
          <tr>
            {#if isImporting}
              <th class="loading-col"></th>
            {/if}
            <th>Name</th>
            <th>Year</th>
            <th>Type</th>
            {#if !isImporting}
              <th></th>
            {/if}
          </tr>
          {#each rList as l}
            <tr>
              {#if isImporting}
                <td class="icon-cell">
                  <div>
                    {#if !l.state}
                      <SpinnerTiny style="width: 13px;" />
                    {:else if l.state === ImportResponseType.IMPORT_SUCCESS}
                      <Icon i="check" wh={22} />
                    {:else if l.state === ImportResponseType.IMPORT_NOTFOUND}
                      <Icon i="close" wh={22} />
                    {:else if l.state === ImportResponseType.IMPORT_FAILED}
                      <Icon i="close" wh={22} />
                    {:else if l.state === ImportResponseType.IMPORT_EXISTS}
                      <Icon i="check" wh={22} />
                    {/if}
                  </div>
                </td>
              {/if}
              <td><input class="plain" bind:value={l.name} disabled={isImporting} /></td>
              <td class="year">
                <input
                  class="plain"
                  bind:value={l.year}
                  placeholder="YYYY"
                  type="number"
                  disabled={isImporting}
                />
              </td>
              <td class="type">
                <DropDown
                  options={["movie", "tv"]}
                  bind:active={l.type}
                  placeholder="Type"
                  blendIn={true}
                  disabled={isImporting}
                />
              </td>
              {#if !isImporting}
                <td>
                  <button
                    class="plain delete"
                    on:click={() => {
                      removeRow(l);
                    }}
                  >
                    <Icon i="close" wh="25" />
                  </button>
                </td>
              {/if}
            </tr>
          {/each}
          {#if !isImporting}
            <tr>
              <td><input class="plain" placeholder="Name" on:blur={addRow} /></td>
              <td class="year">
                <input class="plain" id="addYear" placeholder="YYYY" type="number" />
              </td>
              <td class="type"></td>
              <td></td>
            </tr>
          {/if}
        </table>
        <div class="btns">
          <button on:click={() => goto("/import")}><Icon i="arrow" />Back</button>
          <button on:click={startImport} disabled={isImporting}>Start Importing</button>
        </div>
      {:else}
        <h2>No list</h2>
      {/if}
    </div>
  </div>

  <!-- Multiple results found modal -->
  {#if importMultiItem}
    <Modal
      title="Multiple Results Found"
      desc="Select the correct item for {importMultiItem.original.name}"
      onClose={() => {
        importMultiItem?.callback("closed results modal");
        importMultiItem = undefined;
      }}
    >
      <PosterList type="vertical">
        {#each importMultiItem.results as r}
          <Poster
            media={r}
            small={true}
            disableInteraction={true}
            hideButtons={true}
            onClick={async () => {
              const item = rList.find((i) => i.name === importMultiItem?.original.name);
              if (item) {
                item.tmdbId = r.id;
                item.type = r.media_type;
                try {
                  await doImport(item);
                  importMultiItem?.callback(undefined);
                } catch (err) {
                  importMultiItem?.callback(String(err));
                }
                importMultiItem = undefined;
              } else {
                // TODO: show error notif and update state with error icon
              }
              console.log("multi: Poster clicked", r);
            }}
          />
        {/each}
      </PosterList>
    </Modal>
  {/if}
{:catch err}
  <Error error={err} pretty="Failed to process list!" />
{/await}

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
      max-width: 600px;
      overflow: hidden;
    }
  }

  table {
    margin-top: 20px;
    width: 100%;
    border-spacing: 0px;
    border: 1px solid $accent-color;
    border-radius: 10px;
    font-size: 16px;

    th {
      padding: 12px 15px;
      text-align: left;
      transition: padding 100ms ease;

      &:first-of-type {
        border-top-left-radius: 10px;
      }

      &:last-of-type {
        border-top-right-radius: 10px;
      }

      &.loading-col {
        width: 28px;
        padding: 0;
      }
    }

    tr {
      th {
        background-color: $accent-color;
      }

      &:last-child {
        td:first-of-type {
          border-bottom-left-radius: 10px;
        }

        td:last-of-type {
          border-bottom-right-radius: 10px;
        }
      }

      &:nth-child(odd) td {
        background-color: $accent-color;
      }
    }

    &.is-importing {
      th {
        padding-left: 3px;
      }

      td {
        padding-left: 3px;

        input {
          padding: 7px 0;

          &:focus {
            padding: 7px 5px;
            padding-left: 3px;
          }
        }
      }
    }

    td {
      padding: 5px;

      &.icon-cell {
        padding-right: 3px;

        & > div {
          display: flex;
          padding-left: 4px;
        }
      }

      &.year {
        width: 70px;
      }

      &.type {
        width: 120px;
      }

      input {
        background: transparent;
        color: $text-color;
        border: 0;
        font-size: 16px;
        padding: 0;
        padding: 7px 10px;
        transition: padding 100ms ease;

        &[type="number"] {
          appearance: textfield;

          &::-webkit-outer-spin-button,
          &::-webkit-inner-spin-button {
            -webkit-appearance: none;
            margin: 0;
          }
        }
      }

      button {
        &.delete {
          display: flex;
          justify-content: center;
          color: $placeholder-color;

          &:hover {
            color: $error;
          }
        }
      }
    }
  }

  .btns {
    display: flex;
    flex-flow: row;
    margin-top: 20px;

    button {
      width: max-content;
      gap: 3px;

      &:last-of-type {
        margin-left: auto;
      }
    }
  }
</style>
