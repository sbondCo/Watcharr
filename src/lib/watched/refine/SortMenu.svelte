<script lang="ts">
	import Menu, { type MenuConfig } from "@/lib/Menu.svelte";
	import { store } from "@/store.svelte";
	import { Sorts } from "@/types";

	interface Props extends Pick<MenuConfig, "trackEl"> {}

	let { trackEl }: Props = $props();

	function sortClicked(type: Sorts) {
		window.scrollTo({ top: 0 });
		let mode = "UP";
		// If this sort is already the `activeSort`
		if (store.activeSort[0] == type) {
			if (store.activeSort[1] === "UP") {
				mode = "DOWN";
			} else if (store.activeSort[1] === "DOWN") {
				mode = "";
			}
		}
		if (!mode) {
			// If there is no mode, then we are turning this
			// sort off, so reset the activeSort array.
			store.activeSort = [];
			return;
		}
		store.activeSort = [type, mode];
	}

	function getDirectionClass(sort: string): string {
		if (store.activeSort[0] !== sort) {
			return "";
		}
		if (store.activeSort[1]) {
			return store.activeSort[1].toLowerCase();
		}
		return "";
	}
</script>

<Menu width="180px" {trackEl}>
	<button
		class={`plain ${getDirectionClass(Sorts.DateAdded)}`}
		onclick={() => sortClicked(Sorts.DateAdded)}
	>
		Date Added
	</button>
	<button
		class={`plain ${getDirectionClass(Sorts.LastChanged)}`}
		onclick={() => sortClicked(Sorts.LastChanged)}
	>
		Last Changed
	</button>
	<button
		class={`plain ${getDirectionClass(Sorts.LastFinished)}`}
		onclick={() => sortClicked(Sorts.LastFinished)}
	>
		Last Finished
	</button>
	<button
		class={`plain ${getDirectionClass(Sorts.Rating)}`}
		onclick={() => sortClicked(Sorts.Rating)}
	>
		Rating
	</button>
	<button
		class={`plain ${getDirectionClass(Sorts.Alphabetical)}`}
		onclick={() => sortClicked(Sorts.Alphabetical)}
	>
		Alphabetical
	</button>
	<button
		class={`plain ${getDirectionClass(Sorts.DateReleased)}`}
		onclick={() => sortClicked(Sorts.DateReleased)}
	>
		Release Date
	</button>
</Menu>

<style lang="scss">
	button {
		position: relative;

		&.down::before {
			content: "\2193";
		}

		&.up::before {
			content: "\2191";
		}

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
