<!-- A menu designed to pop up around a button after it is pressed. -->
<script lang="ts">
	export type MenuConfig = Omit<Props, "children">;

	interface Props {
		children: import("svelte").Snippet;
		/**
		 * Instead of `top` or `right` manual props, pass an element for the
		 * menu to get positioning from.
		 */
		trackEl: HTMLElement;
		/**
		 * CSS Width.
		 */
		width?: string;
		/**
		 * Color of arrow.
		 */
		arrowColor?: string;
	}

	let {
		children,
		width = "125px",
		trackEl,
		arrowColor = "unset",
	}: Props = $props();

	let menuEl: HTMLDivElement | undefined = $state();

	let trackElBRect = $derived(trackEl?.getBoundingClientRect());
	let pos = $derived.by(() => {
		if (!trackElBRect || !menuEl) {
			// If these don't exist, we can't do our calculations, just return
			// this for now..
			return { menu: { x: 10, y: 10 } };
		}

		let menuRect = menuEl.getBoundingClientRect();
		const newRect = DOMRect.fromRect(menuRect);

		console.log(trackElBRect);
		const proposedY = trackElBRect.bottom + 10;
		newRect.y = proposedY;

		const mw = (menuRect?.width || 0) / 2;
		// Menu centered to target
		const proposedX =
			window.innerWidth - trackElBRect.x - mw - trackElBRect.width / 2;
		// Min right allowed (overflow protection on right)
		newRect.x = Math.max(proposedX, 10);
		// Max right allowed (overflow protection on left)
		newRect.x = Math.min(newRect.x, window.innerWidth - menuRect.width - 10);

		let arrow = mw - 12;
		if (newRect.x != proposedX) {
			// Move arrow the difference of our proposedX and X used to keep it
			// under the targetEl.
			let d = newRect.x - proposedX;
			arrow -= d;
		}

		return { menu: newRect, arrow };
	});
</script>

<div
	class="menu"
	style:--w={width}
	style:--r={`${pos.menu.x}px`}
	style:--t={`${pos.menu.y}px`}
	style:--ar={`${pos.arrow}px`}
	style:--ac={arrowColor}
	bind:this={menuEl}
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
					transition:
						background-color 200ms ease,
						border-radius 400ms ease-in-out;

					&:hover,
					&:focus-visible {
						background-color: $text-color;
						color: $bg-color;
						border-radius: 5px;
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
