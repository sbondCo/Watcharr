<script lang="ts">
  import type { ContentType, TMDBMovieSimilar, TMDBShowSimilar } from "@/types";
  import HorizontalList from "../HorizontalList.svelte";
  import { watchedList } from "@/store";
  import { getWatchedDependedProps } from "@/lib/util/helpers";
  import Poster from "../poster/Poster.svelte";

  $: wList = $watchedList;

  export let type: ContentType;
  export let similar: TMDBShowSimilar | TMDBMovieSimilar;
</script>

{#if similar?.results?.length > 0}
  <HorizontalList title="Similar">
    {#each similar.results as content}
      <Poster
        media={{ ...content, media_type: type }}
        {...getWatchedDependedProps(content.id, type, wList)}
        small={true}
      />
    {/each}
  </HorizontalList>
{/if}
