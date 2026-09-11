<script lang="ts">
	import type { HandbookEntryId } from "../handbook/entries";
	import HandbookLink from "../handbook/HandbookLink.svelte";

	interface Props {
		title?: string;
		desc?: string;
		row?: boolean;
		tag?: string | undefined;
		children?: import("svelte").Snippet;
		handbook?: HandbookEntryId;
	}

	let {
		title = "",
		desc = "",
		row = false,
		tag = undefined,
		children,
		handbook,
	}: Props = $props();
</script>

<div class={[row ? "row" : "", "setting-ctr"].join(" ")}>
	{#if row}
		<div>
			<h4 class="norm">
				{title}
				{#if handbook}
					<HandbookLink {handbook} />
				{/if}
				{#if tag}
					<span class="tag">{tag}</span>
				{/if}
			</h4>
			<h5 class="norm">{desc}</h5>
		</div>
	{:else}
		<h4 class="norm">
			{title}
			{#if tag}
				<span class="tag">{tag}</span>
			{/if}
		</h4>
		<h5 class="norm">{desc}</h5>
	{/if}

	{@render children?.()}
</div>

<style lang="scss">
	h4 {
		font-size: 16px;
		display: flex;
		flex-flow: row;
		align-items: center;
		gap: 5px;

		& > .tag {
			font-size: 10px;
			background-color: $accent-color;
			padding: 3px 5px;
			border-radius: 5px;
		}
	}

	h5 {
		font-weight: normal;
		font-size: 13.28px;
	}

	.setting-ctr {
		&.row {
			display: flex;
			flex-flow: row;
			gap: 10px;
			align-items: center;

			& > div:first-of-type {
				margin-right: auto;
			}
		}

		&:not(.row) {
			h5 {
				margin-bottom: 5px;
			}
		}
	}
</style>
