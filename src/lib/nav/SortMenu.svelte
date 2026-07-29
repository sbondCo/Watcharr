<script lang="ts">
	import { store } from "@/store.svelte";
	import Menu from "../Menu.svelte";

	function sortClicked(type: string) {
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

<Menu conf={{ width: "180px", right: "90px", arrowLeft: "21px" }}>
	<button
		class={`plain ${getDirectionClass("DATEADDED")}`}
		onclick={() => sortClicked("DATEADDED")}
	>
		Date Added
	</button>
	<button
		class={`plain ${getDirectionClass("LASTCHANGED")}`}
		onclick={() => sortClicked("LASTCHANGED")}
	>
		Last Changed
	</button>
	<button
		class={`plain ${getDirectionClass("LASTFIN")}`}
		onclick={() => sortClicked("LASTFIN")}
	>
		Last Finished
	</button>
	<button
		class={`plain ${getDirectionClass("RATING")}`}
		onclick={() => sortClicked("RATING")}
	>
		Rating
	</button>
	<button
		class={`plain ${getDirectionClass("ALPHA")}`}
		onclick={() => sortClicked("ALPHA")}
	>
		Alphabetical
	</button>
	<button
		class={`plain ${getDirectionClass("DATERELEASED")}`}
		onclick={() => sortClicked("DATERELEASED")}
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
