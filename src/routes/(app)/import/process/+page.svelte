<!-- 
  /import/process is for processing the
  selected files data. Here it will be
  displayed and imported.

  TODO:
    - Go back to main import page if no list
 -->

<script lang="ts">
  import { goto } from "$app/navigation";
  import DropDown from "@/lib/DropDown.svelte";
  import Error from "@/lib/Error.svelte";
  import Icon from "@/lib/Icon.svelte";
  import Modal from "@/lib/Modal.svelte";
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
  import { sleep } from "@/lib/util/helpers";
  import { importedList } from "@/store";
  import {
    ImportResponseType,
    type ImportResponse,
    type ContentSearchTv,
    type ContentSearchMovie,
    type ContentType
  } from "@/types";
  import axios from "axios";
  import { get } from "svelte/store";

  interface ImportedList {
    tmdbId?: number;
    // TODO: this property can be unique if we remove duplicates
    name: string;
    year?: string;
    type?: ContentType;
    state?: string;
  }

  interface ImportedListItemMultiProblem {
    original: ImportedList;
    results: (ContentSearchMovie | ContentSearchTv)[];
    callback: (err: Error | string | undefined) => void;
  }

  let rList: ImportedList[] = [];
  let isImporting = false;

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
    // TODO transform the data into readable depending on type
    if (list?.file.type === "text/plain") {
      // Regex to match a year in between brackets,
      // which we assume is the release year of content.
      const yearRegex = new RegExp(/\([0-9]{4}\)/);
      const s = list.data.split("\n");
      for (let i = 0; i < s.length; i++) {
        const el = s[i];
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
    }
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

  async function startImport() {
    console.log(rList);
    isImporting = true;
    for (let i = 0; i < rList.length; i++) {
      const li = rList[i];
      console.log("Importing", li);
      await doImport(li);
      await sleep(2000);
    }
  }

  async function doImport(item: ImportedList) {
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
              rej(err);
            } else {
              res(0);
            }
          }
        };
      } else if (resp.data.type === ImportResponseType.IMPORT_SUCCESS) {
        item.state = ImportResponseType.IMPORT_SUCCESS;
        const match = resp.data.match;
        if (match) {
          const release = match.media_type === "movie" ? match.release_date : match.first_air_date;
          if (release) item.year = String(new Date(Date.parse(release)).getFullYear());
          item.type = match.media_type;
        }
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
        <h2>Importing {list?.file.name ? list.file.name : ""}</h2>
        <h5 class="norm">Review your imported list and fix any problems.</h5>
        <table class={isImporting ? "is-importing" : ""}>
          <tr>
            {#if isImporting}
              <th class="loading-col"></th>
            {/if}
            <th>Name</th>
            <th>Year</th>
            <th>Type</th>
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
                    {/if}
                  </div>
                </td>
              {/if}
              <td><input class="plain" bind:value={l.name} /></td>
              <td class="year">
                <input class="plain" bind:value={l.year} placeholder="YYYY" type="number" />
              </td>
              <td class="type">
                <DropDown
                  options={["movie", "tv"]}
                  bind:active={l.type}
                  placeholder="Type"
                  blendIn={true}
                />
              </td>
            </tr>
          {/each}
          {#if !isImporting}
            <tr>
              <td><input class="plain" placeholder="Name" on:blur={addRow} /></td>
              <td class="year">
                <input class="plain" id="addYear" placeholder="YYYY" type="number" />
              </td>
              <td class="type"></td>
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
    padding: 0 30px 0 30px;

    .inner {
      display: flex;
      flex-flow: column;
      min-width: 400px;
      max-width: 400px;
      overflow: hidden;
    }
  }

  table {
    margin-top: 20px;
    // table-layout: fixed;
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
