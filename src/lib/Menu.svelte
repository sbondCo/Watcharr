<!-- A menu designed to pop up around a
 	button after it is pressed. -->
<script lang="ts">
	import stayInView from "./actions/stayInView";

	export interface MenuConfig {
		width?: string;
		top?: string;
		right?: string;
		arrowRight?: string;
		arrowLeft?: string;
		arrowColor?: string;
	}

	interface Props {
		children: import("svelte").Snippet;
		conf?: MenuConfig;
	}

	let { children, conf }: Props = $props();
</script>

<div
	class="menu"
	style={`--w: ${conf?.width || "125px"}; --r: ${
		conf?.right || "10px"
	}; --t: ${conf?.top || "55px"}; --ar: ${
		conf?.arrowRight || "unset"
	}; --al: ${conf?.arrowLeft || "unset"}; --ac: ${conf?.arrowColor || "unset"};`}
	use:stayInView={{ elToShiftSelector: "& > .arrow" }}
>
	<i class="arrow"></i>
	<div>
		{@render children?.()}
	</div>
</div>

<style lang="scss">
	.menu {
		display: flex;
		flex-flow: column;
		position: absolute;
		right: var(--r);
		top: var(--t);
		width: var(--w);
		border: 3px solid $text-color;
		border-radius: 10px;
		background-color: $bg-color;
		list-style: none;
		z-index: 50;

		// The lil arrow, adjust its pos under each specific menu.
		&:not(:has(> .arrow)):before,
		:global(.arrow) {
			// Works by showing arrow in ::before pseudoelement,
			// or if more control is needed, applies to any .arrow
			// element.
			content: "";
			position: absolute;
			bottom: 100%;
			width: 0;
			border-top: 10px solid transparent;
			border-left: 10px solid transparent;
			border-right: 10px solid transparent;
			border-bottom: 10px solid var(--ac, $text-color);
			right: var(--ar);
			left: var(--al);
			pointer-events: none;
		}

		& > div {
			display: flex;
			flex-flow: column;
			padding: 10px;
			width: 100%;
			max-height: calc(100dvh - 65px);
			overflow: auto;

			:global {
				h5 {
					margin-bottom: 2px;
					white-space: nowrap;
					overflow: hidden;
					text-overflow: ellipsis;
					cursor: default;
				}

				button,
				a {
					font-size: 14px;
					padding: 8px 16px;
					text-align: center;
					cursor: pointer;
					transition: background-color 200ms ease;

					&:hover,
					&:focus-visible {
						background-color: $text-color;
						color: $bg-color;
					}
				}

				a.menu-footer,
				button.menu-footer {
					font-size: 11px;
					padding: 2px;
					text-align: center;
					transition: inherit;
					color: $text-color-accent;
					cursor: pointer;
					border: none;
					display: inline;
					font-weight: initial;
					width: auto;

					&:hover,
					&:focus-visible {
						text-decoration: underline;
						background-color: $bg-color;
						color: $text-color;
					}
				}

				span {
					margin-top: 8px;
					font-size: 11px;
					color: $text-color-accent;
					text-align: center;
				}
			}
		}
	}
</style>
