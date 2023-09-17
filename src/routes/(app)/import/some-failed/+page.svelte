<!-- 
  /import/some-failed shows the user the
  failed imports so they can manually import
  them instead.
 -->

<script lang="ts">
  import { goto } from "$app/navigation";
  import Icon from "@/lib/Icon.svelte";
  import { parsedImportedList } from "@/store";
  import { ImportResponseType, type ImportedList } from "@/types";
  import { onMount } from "svelte";
  import { get } from "svelte/store";

  let failed: ImportedList[] = [];
  let successCount = 0;

  onMount(() => {
    let list = get(parsedImportedList);
    if (list) {
      for (let i = 0; i < list.length; i++) {
        const item = list[i];
        console.log(item);
        if (
          item.state === ImportResponseType.IMPORT_FAILED ||
          item.state === ImportResponseType.IMPORT_NOTFOUND
        ) {
          failed.push(item);
          failed = failed;
        } else {
          successCount++;
        }
      }
      console.log("failedlen", failed.length);
    } else {
      goto("/import");
    }
  });
</script>

<div class="content">
  <div class="inner">
    <h2>Some Content Failed To Import</h2>
    <h5 class="norm">You can search for the failed imports and manually add them.</h5>
    <h4 class="norm">{successCount} succeeded and {failed.length} failed.</h4>

    {#if failed}
      <ul>
        {#each failed as l}
          <li>
            <span>{l.name}</span>
            <a href="/search/{l.name}" target="_blank">
              <button><Icon i="search" /></button>
            </a>
          </li>
        {/each}
      </ul>
    {:else}
      No List
    {/if}
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
      max-width: 600px;
      overflow: hidden;
    }
  }

  h4 {
    margin-top: 15px;
  }

  ul {
    display: flex;
    flex-flow: column;
    gap: 5px;
    margin: 10px;
    list-style: none;

    li {
      display: flex;
      flex-flow: row;
      align-items: center;
      padding: 10px;
      background-color: $accent-color;
      border-radius: 5px;

      a {
        margin-left: auto;

        button {
          width: max-content;
        }
      }
    }
  }
</style>
