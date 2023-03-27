<script lang="ts">
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";

  export let data: import("./$types").PageData;
</script>

<svelte:head>
  <title>Watcharr</title>
</svelte:head>

<PosterList>
  {#if data?.watched && data.watched.length > 0}
    {#each data.watched as w}
      <Poster poster={w.poster} title={w.title} desc={"ah"} />
    {/each}
  {:else}
    You don't have any watched content yet!
  {/if}
</PosterList>
