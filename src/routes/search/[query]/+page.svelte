<script lang="ts">
  import Poster from "@/lib/Poster.svelte";
  import type { ContentSearch } from "./+page";

  export let data: ContentSearch;
</script>

<div>
  <ul>
    {#each data.results as w}
      <li><Poster poster={"https://image.tmdb.org/t/p/w500" + w.poster_path} /></li>
    {/each}
  </ul>
</div>
