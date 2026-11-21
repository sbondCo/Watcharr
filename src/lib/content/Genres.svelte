<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import type { MediaGenre } from "@/types";

	interface Props {
		genres?: MediaGenre[];
	}

	let { genres }: Props = $props();
</script>

<div>
	{#if genres && genres.length > 0}
		{#each genres as g, i (g.id)}
			<span>
				{g.name}{i !== genres.length - 1 ? ", " : ""}
			</span>
		{/each}
	{:else}
		<!-- Generic "unknown" text, since this component
		 	 is also used for displaying Game Modes. -->
		<span>{t("common.unknown")}</span>
	{/if}
</div>

<style lang="scss">
</style>
