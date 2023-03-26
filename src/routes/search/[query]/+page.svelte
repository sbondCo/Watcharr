<script lang="ts">
  import Poster from "@/lib/Poster.svelte";
  import PosterList from "@/lib/PosterList.svelte";
  import type { ContentSearch } from "./+page";

  export let data: ContentSearch;
</script>

<PosterList>
  {#each data.results as w}
    <Poster poster={"https://image.tmdb.org/t/p/w500" + w.poster_path} />
  {/each}
</PosterList>
