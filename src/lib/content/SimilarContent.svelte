<script lang="ts">
  import type { ContentType, TMDBMovieSimilar, TMDBShowSimilar } from "@/types";
  import HorizontalList from "../HorizontalList.svelte";
  import Poster from "../Poster.svelte";
  import { watchedList } from "@/store";
  import { getWatchedDependedProps } from "@/lib/util/helpers";

  $: wList = $watchedList;

  export let type: ContentType;
  export let similar: TMDBShowSimilar | TMDBMovieSimilar;
</script>

<HorizontalList title="Similar">
  {#each similar.results as content}
    <Poster
      media={{ ...content, media_type: type }}
      {...getWatchedDependedProps(content.id, type, wList)}
      small={true}
    />
  {/each}
</HorizontalList>
