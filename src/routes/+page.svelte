<script lang="ts">
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import WatchedModal from "@/lib/WatchedModal.svelte";

  export let data: import("./$types").PageData;
</script>

<svelte:head>
  <title>Watcharr</title>
</svelte:head>

<PosterList>
  {#each data.watched as w}
    <Poster poster={w.poster} title={w.title} desc={"ah"} />
  {/each}
</PosterList>
