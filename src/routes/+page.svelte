<script lang="ts">
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";

  export let data: import("./$types").PageData;
</script>

<svelte:head>
  <title>Watcharr</title>
</svelte:head>

<PosterList>
  {#each data.watched as w}
    <Poster poster={w.poster} />
  {/each}
</PosterList>
