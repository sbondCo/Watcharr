<script lang="ts">
	import type { ResolvedPathname } from "$app/types";
	import tooltip from "../actions/tooltip";

	interface Props {
		name: string;
		value: string | number;
		large?: boolean;
		/**
		 * **NOTE:** Make sure to use resolve() when passing in the href!
		 */
		href?: ResolvedPathname;
		disc?: string | undefined;
	}

	let {
		name,
		value,
		large = false,
		href = undefined,
		disc = undefined,
	}: Props = $props();
</script>

<a {href}>
	{#if disc}
		<div class="disclaimer" use:tooltip={{ text: disc, pos: "top" }}>*</div>
	{/if}
	<span class={large ? "large" : ""}>{value}</span>
	<span>{name}</span>
</a>

<style lang="scss">
	a {
		position: relative;
		display: flex;
		flex-flow: column;
		flex-grow: 1;
		flex: 1 1 115px;
		padding: 20px 15px;
		background-color: $accent-color;
		border-radius: 8px;

		> span:first-of-type {
			font-weight: bold;
			font-size: 20px;
			margin-top: auto;

			&.large {
				font-size: 32px;
			}
		}

		.disclaimer {
			position: absolute;
			top: 5px;
			right: 8px;
			font-weight: bold;
			font-size: 20px;
		}
	}
</style>
