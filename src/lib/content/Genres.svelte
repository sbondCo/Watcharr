<script lang="ts">
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
		<span>Unknown</span>
	{/if}
</div>

<style lang="scss">
</style>
