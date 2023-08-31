<!-- 
  /import/process is for processing the
  selected files data. Here it will be
  displayed and imported.
 -->

<script lang="ts">
  import Error from "@/lib/Error.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { importedList } from "@/store";
  import { get } from "svelte/store";

  console.log(importedList);

  async function getList() {
    const list = get(importedList);
    console.log("getList", list);
    // TODO transform the data into readable depending on type
    return list;
  }
</script>

{#await getList()}
  <Spinner />
{:then list}
  <div class="content">
    <div class="inner">
      {#if list}
        <h2>Importing {list.file.name ? list.file.name : ""}</h2>
        <h4 class="norm">Review your imported list.</h4>
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
</style>
