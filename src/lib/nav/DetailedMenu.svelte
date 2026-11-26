<script lang="ts">
	import { store } from "@/store.svelte";
	import type { WLDetailedViewOption } from "@/types";
	import Menu, { type MenuConfig } from "../Menu.svelte";

	interface Props extends Pick<MenuConfig, "trackEl"> {}

	let { trackEl }: Props = $props();

	function detailClicked(d: WLDetailedViewOption) {
		if (store.wlDetailedView.includes(d)) {
			store.wlDetailedView = store.wlDetailedView.filter((a) => a !== d);
		} else {
			store.wlDetailedView.push(d);
			store.wlDetailedView = store.wlDetailedView;
		}
	}
</script>

<Menu width="200px" {trackEl}>
	<h4 class="norm sm-caps">Shown Details</h4>
	<button
		class={`plain ${store.wlDetailedView?.includes("statusRating") ? "on" : ""}`}
		onclick={() => detailClicked("statusRating")}
	>
		Rating
	</button>
	<button
		class={`plain ${store.wlDetailedView?.includes("status") ? "on" : ""}`}
		onclick={() => detailClicked("status")}
	>
		Status
	</button>
	<button
		class={`plain ${store.wlDetailedView?.includes("lastWatched") ? "on" : ""}`}
		onclick={() => detailClicked("lastWatched")}
	>
		Watching Season
	</button>
	<button
		class={`plain ${store.wlDetailedView?.includes("dateAdded") ? "on" : ""}`}
		onclick={() => detailClicked("dateAdded")}
	>
		Date Added
	</button>
	<button
		class={`plain ${store.wlDetailedView?.includes("dateModified") ? "on" : ""}`}
		onclick={() => detailClicked("dateModified")}
	>
		Date Modified
	</button>
</Menu>

<style lang="scss">
	h4 {
		margin-bottom: 8px;

		&:not(:first-of-type) {
			margin-top: 8px;
		}
	}

	button.plain {
		text-transform: capitalize;
		position: relative;

		&.on::before {
			content: "\2713";
		}

		&::before {
			position: absolute;
			top: 4px;
			left: 12px;
			font-family:
				system-ui,
				-apple-system,
				BlinkMacSystemFont;
			font-size: 18px;
		}
	}
</style>
