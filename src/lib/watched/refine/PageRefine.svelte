<script lang="ts">
	import Icon from "@/lib/Icon.svelte";
	import SortMenu from "./SortMenu.svelte";
	import FilterMenu from "./FilterMenu.svelte";
	import dismissable from "@/lib/attachments/dismissable.svelte";

	let filterMenuShown = $state(false);
	let sortMenuShown = $state(false);

	let sortBtn: HTMLElement | undefined = $state();
	let filterBtn: HTMLElement | undefined = $state();
</script>

<div class="ctr">
	<div {@attach sortMenuShown && dismissable(() => (sortMenuShown = false))}>
		<button
			onclick={() => {
				sortMenuShown = !sortMenuShown;
			}}
			bind:this={sortBtn}
		>
			<Icon i="sort" wh={16} />
			Sort
		</button>

		{#if sortMenuShown}
			<SortMenu trackEl={sortBtn} />
		{/if}
	</div>

	<div
		{@attach filterMenuShown && dismissable(() => (filterMenuShown = false))}
	>
		<button
			onclick={() => {
				filterMenuShown = !filterMenuShown;
			}}
			bind:this={filterBtn}
		>
			<Icon i="filter" wh={16} />
			Filter
		</button>

		{#if filterMenuShown}
			<FilterMenu trackEl={filterBtn} />
		{/if}
	</div>
</div>

<style lang="scss">
	.ctr {
		display: flex;
		flex-flow: row;
		gap: 10px;
	}

	button {
		position: relative;
		display: flex;
		flex-flow: row;
		gap: 5px;
		align-items: center;
		justify-content: center;
		width: min-content;
		min-width: 0;
	}
</style>
