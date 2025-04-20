<script lang="ts">
	import type { ContentType, TMDBMovieSimilar, TMDBShowSimilar } from "@/types";
	import HorizontalList from "../HorizontalList.svelte";
	import Poster from "../poster/Poster.svelte";

	interface Props {
		type: ContentType;
		similar: TMDBShowSimilar | TMDBMovieSimilar;
	}

	let { type, similar }: Props = $props();
</script>

{#if similar?.results?.length > 0}
	<HorizontalList title="Similar">
		{#each similar.results as content, i}
			<Poster
				media={{ ...content, media_type: type }}
				small={true}
				bind:watched={similar.results[i].watched}
			/>
		{/each}
	</HorizontalList>
{/if}
