<script lang="ts">
  import Poster from "@/lib/Poster.svelte";

  export let data: import("./$types").PageData;
</script>

<svelte:head>
  <title>Watcharr</title>
</svelte:head>

<div>
  <div><h1>Watched</h1></div>

  <ul>
    {#each data.watched as w}
      <li><Poster content={w} /></li>
    {/each}
  </ul>
</div>

<style lang="scss">
  div {
    display: flex;
    flex-flow: column;
    justify-content: center;

    ul {
      display: flex;
      flex-flow: row;
      gap: 10px;
      list-style: none;
      flex-wrap: wrap;
    }
  }
</style>
