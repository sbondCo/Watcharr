<script lang="ts">
	import { resolve } from "$app/paths";
	import type { HandbookEntryId } from "./entries";
	import { openHandbookModal } from "./handbook";

	interface Props {
		handbook: HandbookEntryId;
	}

	let { handbook }: Props = $props();

	/**
	 * When the link is clicked "normally", we want to open our modal.
	 */
	function linkClicked(ev: MouseEvent) {
		if (
			ev.button !== 0 ||
			ev.ctrlKey ||
			ev.shiftKey ||
			ev.metaKey ||
			ev.defaultPrevented
		) {
			// Let browser do its thing for normal link actions.
			return;
		}
		ev.preventDefault();
		openHandbookModal(handbook);
	}
</script>

<a href={resolve(`/handbook?entry=${handbook}`)} onclick={linkClicked}>?</a>

<style lang="scss">
	a {
		font-size: 15px;
	}
</style>
