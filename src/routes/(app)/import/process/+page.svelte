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
            l.name = l.name.replace(yearRegex, "");
          }
          rList.push(l);
        }
      }
    }
    return list;
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
              <td>{l.name}</td>
              <td>{l.year ? l.year : "Unknown"}</td>
              <!-- <td>Unknown</td> -->
            </tr>
          {/each}
        </table>
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

    th {
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

    td,
    th {
      padding: 8px;
    }
  }
</style>
