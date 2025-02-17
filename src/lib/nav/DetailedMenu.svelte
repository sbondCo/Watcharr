<script lang="ts">
	import { store } from "@/store.svelte";
	import type { WLDetailedViewOption } from "@/types";
	import { page } from "$app/state";
	import Menu from "../Menu.svelte";

	function detailClicked(d: WLDetailedViewOption) {
		if (store.wlDetailedView.includes(d)) {
			store.wlDetailedView = store.wlDetailedView.filter((a) => a !== d);
		} else {
			store.wlDetailedView.push(d);
			store.wlDetailedView = store.wlDetailedView;
		}
	}
</script>

<Menu
	conf={{
		width: "200px",
		right: "92px",
		arrowLeft: page.url?.pathname.startsWith("/search") ? "84px" : "3px",
	}}
>
	<h4 class="norm sm-caps">Shown Details</h4>
	<button
		class={`plain ${store.wlDetailedView?.includes("statusRating") ? "on" : ""}`}
		onclick={() => detailClicked("statusRating")}
	>
		Status & Rating
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
