<script lang="ts">
	import Chip from "@/lib/generic/Chip.svelte";
	import Icon from "@/lib/Icon.svelte";
	import { store } from "@/store.svelte";
</script>

<div class="list">
	{#if store.activeSort && store.activeSort[0]}
		<Chip body={store.activeSort[0]}>
			{#snippet icon()}
				<Icon
					i="arrow"
					facing={store.activeSort[1] === "UP" ? "up" : "down"}
					wh="18"
				/>
			{/snippet}
		</Chip>
	{/if}

	{#snippet filterIcon()}
		<Icon i="add" wh="16" />
	{/snippet}

	{#if store.activeFilters.type}
		{#each store.activeFilters.type as t (t)}
			<Chip body={t === "tv" ? "show" : t} bg="success" icon={filterIcon} />
		{/each}
		{#each store.activeFilters.status as s (s)}
			<Chip body={s} bg="success" icon={filterIcon} />
		{/each}
	{/if}
</div>

<style lang="scss">
	.list {
		display: flex;
		flex-flow: row;
		flex-wrap: wrap;
		gap: 5px;
	}
</style>
