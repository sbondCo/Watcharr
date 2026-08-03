<script lang="ts">
	import Icon from "@/lib/Icon.svelte";
	import Modal from "@/lib/Modal.svelte";

	let modalShown = $state(false);
</script>

<button
	onclick={() => (modalShown = true)}
	style="width: min-content; margin-left:auto; gap: 5px;"
>
	<Icon i="filter" wh={16} /> Filters
</button>

{#if modalShown}
	<Modal title="Filters Help" onClose={() => (modalShown = false)}>
		<div class="modal-inner">
			<p>
				Currently to apply filters you must include them in your search bar
				manually.
			</p>
			<p><b>Supported Filters</b></p>
			<ul>
				<li>
					<b>year</b>, <b>y</b> - Media has <i>any</i> release date matching the year
					(only works with a media type selected, like Movies).
				</li>
				<li>
					<b>fyear</b>, <b>fy</b> - Media was <i>first</i> released on the year (only
					works with a media type selected, like Movies).
				</li>
				<li>
					<b>adult</b> - Toggle on adult results. This always defaults to "off" (Only
					Movies, TV Shows, People). Set this to any value to turn it "on".
				</li>
			</ul>
			<p><b>Examples</b></p>
			<ul>
				<li>
					<b>Filter by any matching release year:</b>
					<ul>
						<li>Spider Man year:2026</li>
						<li>Spider Man y:2026</li>
					</ul>
				</li>
				<li>
					<b>Filter by primary/first release year:</b>
					<ul>
						<li>Spider Man fyear:2026</li>
						<li>Spider Man fy:2026</li>
					</ul>
				</li>
				<li>
					<b>Toggle on adult movies, tv or people in results:</b>
					<ul>
						<li>Spider Man adult:1</li>
					</ul>
				</li>
			</ul>
		</div>
	</Modal>
{/if}

<style lang="scss">
	.modal-inner {
		display: flex;
		flex-flow: column;
		gap: 5px;

		ul {
			list-style: inside;
			margin-left: 5px;
		}

		ul li > ul {
			list-style: inside circle;
			margin-left: 15px;
		}
	}
</style>
