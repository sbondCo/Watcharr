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
      <Poster
        poster={"http://localhost:3080/img" + w.content.poster_path}
        title={w.content.title ?? w.content.name}
        desc={w.content.overview}
      />
    {/each}
  {:else}
    You don't have any watched content yet!
  {/if}
</PosterList>
