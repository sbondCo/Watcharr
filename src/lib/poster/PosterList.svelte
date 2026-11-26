<script lang="ts">
	import { onMount } from "svelte";

	interface Props {
		type?: "wrapped" | "vertical";
		children?: import("svelte").Snippet;
	}

	let { type = "wrapped", children }: Props = $props();

	let ulEl: HTMLUListElement | undefined = $state();

	onMount(() => {
		if (ulEl) {
			ulEl.classList.add(type);
		}
	});
</script>

<div>
	<ul bind:this={ulEl}>
		{@render children?.()}
	</ul>
</div>

<style lang="scss">
	div {
		display: flex;
		justify-content: center;
	}

	ul {
		display: flex;
		flex-flow: row;
		justify-content: center;
		gap: 10px;
		list-style: none;
		flex-wrap: wrap;
		margin: 20px 10px;
		max-width: 1200px;

		&:global(.vertical) {
			flex-wrap: nowrap;
			justify-content: unset;
			overflow-x: auto;
			padding: 15px 8px;
			margin: 5px 0;
		}

		&:global(.wrapped li) {
			position: relative;
			animation: float-in 150ms ease both;
			/* Incrementing animation delay, resetting each 40 elements (our
				page size). Technically a bug if your viewport is really large
				that multiple pages load at once, but mine isn't that big...
				also not sure how to fix that without complicated js. */
			animation-delay: calc(mod(sibling-index() - 1, 40) * 25ms);

			&:hover {
				z-index: 999;
			}
		}

		@keyframes float-in {
			0% {
				top: 100px;
				opacity: 0;
			}
			100% {
				top: 0;
				opacity: 1;
			}
		}

		&:global(.wrapped li) {
			@media screen and (max-width: 390px) {
				flex: 0 48.5%;
			}
			@media screen and (max-width: 352px) {
				flex: 0 48%;
			}
			@media screen and (width <= 349px) {
				flex: unset;
			}
		}

		&:global(.wrapped .container) {
			min-width: 150px !important;

			@media screen and (width <= 349px), screen and (width > 390px) {
				width: 170px !important;
				min-height: 256.367px !important;
			}
		}
	}
</style>
