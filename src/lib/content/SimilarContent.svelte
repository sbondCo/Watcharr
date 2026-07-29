<script lang="ts">
	import type { Media } from "@/types";
	import HorizontalList from "../HorizontalList.svelte";
	import Poster from "../poster/Poster.svelte";

	interface Props {
		similar: Media[];
	}

	let { similar }: Props = $props();
</script>

{#if similar?.length > 0}
	<HorizontalList title="Similar">
		{#each similar as content, i (content.ids)}
			<Poster media={content} small={true} bind:watched={similar[i].watched} />
		{/each}
	</HorizontalList>
{/if}
