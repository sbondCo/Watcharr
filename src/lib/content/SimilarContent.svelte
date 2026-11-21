<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import type { Media } from "@/types";
	import HorizontalList from "../HorizontalList.svelte";
	import Poster from "../poster/Poster.svelte";

	interface Props {
		similar: Media[];
	}

	let { similar }: Props = $props();
</script>

{#if similar?.length > 0}
	<HorizontalList title={t("content.similar")}>
		{#each similar as content, i (content.ids)}
			<Poster media={content} small={true} bind:watched={similar[i].watched} />
		{/each}
	</HorizontalList>
{/if}
