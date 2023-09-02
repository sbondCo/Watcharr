<!-- 
  /import/process is for processing the
  selected files data. Here it will be
  displayed and imported.

  TODO:
    - Go back to main import page if no list
 -->

<script lang="ts">
  import { goto } from "$app/navigation";
  import Error from "@/lib/Error.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { importedList } from "@/store";
  import { get } from "svelte/store";

  interface ImportedList {
    name: string;
    year?: string;
    type?: string;
  }

  let rList: ImportedList[] = [];

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

  function doImport() {
    console.log(rList);
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
        <table>
          <tr>
            <th>Name</th>
            <th>Year</th>
            <!-- <th>Type</th> -->
          </tr>
          {#each rList as l}
            <tr>
              <td><input class="plain" bind:value={l.name} /></td>
              <td><input class="plain" bind:value={l.year} placeholder="Unknown" /></td>
              <!-- <td>Unknown</td> -->
            </tr>
          {/each}
          <tr>
            <td><input class="plain" placeholder="Name" on:blur={addRow} /></td>
            <td><input class="plain" id="addYear" placeholder="Unknown" /></td>
            <!-- <td>Unknown</td> -->
          </tr>
        </table>
        <div class="btns">
          <button on:click={doImport}>Start Importing</button>
        </div>
      {:else}
        <h2>No list</h2>
      {/if}
    </div>
  </div>
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
    table-layout: fixed;
    width: 100%;
    border-spacing: 0px;
    border: 1px solid $accent-color;
    border-radius: 10px;
    font-size: 16px;

    th {
      padding: 12px 15px;
      text-align: left;

      &:first-of-type {
        border-top-left-radius: 10px;
      }

      &:last-of-type {
        border-top-right-radius: 10px;
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

    td {
      padding: 5px;

      input {
        background: transparent;
        border: 0;
        font-size: 16px;
        padding: 0;
        padding: 7px 10px;
      }
    }
  }

  .btns {
    display: flex;
    flex-flow: row;
    margin-top: 20px;

    button {
      width: max-content;

      &:last-of-type {
        margin-left: auto;
      }
    }
  }
</style>