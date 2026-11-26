<script lang="ts">
	import PageTitle from "@/lib/generic/PageTitle.svelte";
	import ActiveRefiners from "./ActiveRefiners.svelte";
	import PageRefine from "./PageRefine.svelte";

	interface Props {
		totalResults: number;
		hiddenResults: number;
	}

	let { totalResults, hiddenResults }: Props = $props();

	let countBadge = $derived.by(() => {
		let str = `${totalResults}`;
		if (hiddenResults) {
			str += ` of ${totalResults + hiddenResults}`;
		}
		return str;
	});
</script>

<PageTitle title="Your List" {countBadge}>
	{#snippet middle()}
		<ActiveRefiners />
	{/snippet}
	{#snippet right()}
		<PageRefine />
	{/snippet}
</PageTitle>
