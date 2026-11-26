<script lang="ts">
	import { type Snippet } from "svelte";
	import CountBadge from "./CountBadge.svelte";

	interface Props {
		title: string;
		row?: boolean;
		children?: Snippet;
		countBadge?: number | string;

		middle?: Snippet;
		right?: Snippet;
	}

	let {
		title,
		row = true,
		children,
		countBadge,
		middle,
		right,
	}: Props = $props();

	let ctr: HTMLDivElement | undefined = $state();
	let titleWrap: HTMLDivElement | undefined = $state();
	let middleCtr: HTMLDivElement | undefined = $state();
	let middleWrap: HTMLDivElement | undefined = $state();
	let rightWrap: HTMLDivElement | undefined = $state();

	/**
	 * Splits layout when too small for `title | middle | right` and turns it
	 * into `title | right` with `middle` wrapping to new line with `split` class.
	 */
	function split() {
		if (!ctr || !titleWrap || !middleCtr || !middleWrap || !rightWrap) {
			console.error("split: Required elements not found.");
			return;
		}
		// 10 is our css `gap`, 2 is each gap spacing we need to account for (3 els - 1).
		let otherWidth = 10 * 2;
		otherWidth += titleWrap.getBoundingClientRect().width;
		otherWidth += rightWrap.getBoundingClientRect().width;

		const ctrWidth = ctr.getBoundingClientRect().width;
		const middleWidth = middleWrap.getBoundingClientRect().width;

		if (ctrWidth < middleWidth + otherWidth) {
			console.debug("PageTitle: Split!");
			middleCtr.classList.add("split");
			return;
		}
		console.debug("PageTitle: Unsplit!");
		middleCtr.classList.remove("split");
	}

	function observe() {
		if (!middle) {
			return;
		}
		return (el: Element) => {
			const o = new ResizeObserver(split);
			o.observe(el);
			return () => o.disconnect();
		};
	}
</script>

<div class="ctr" class:row {@attach observe()} bind:this={ctr}>
	<div class="title">
		<div class="wrap" bind:this={titleWrap}>
			<h2>{title}</h2>
			{#if countBadge}
				<CountBadge count={countBadge} />
			{/if}
		</div>
	</div>
	{#if children}
		{@render children()}
	{/if}
	{#if middle}
		<div class="middle" bind:this={middleCtr}>
			<div class="wrap" bind:this={middleWrap}>
				{@render middle()}
			</div>
		</div>
	{/if}
	{#if right}
		<div class="right">
			<div class="wrap" bind:this={rightWrap}>
				{@render right()}
			</div>
		</div>
	{/if}
</div>

<style lang="scss">
	.ctr {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 10px;
		margin: 0 15px;

		&.row {
			flex-flow: row wrap;
		}

		.title .wrap {
			display: flex;
			flex-flow: row;
			align-items: center;
			gap: 10px;
			max-width: max-content;
		}

		&:has(.middle) .title {
			flex: 1;
			white-space: nowrap;
		}

		:global .middle.split {
			order: 1;
			display: flex;
			justify-content: center;
			flex-basis: 100%;
		}

		.right {
			flex: 1;
			display: flex;
			justify-content: end;
		}
	}
</style>
