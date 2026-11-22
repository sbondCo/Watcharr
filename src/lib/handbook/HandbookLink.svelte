<script lang="ts">
	import { resolve } from "$app/paths";
	import tooltip from "../actions/tooltip";
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

<a
	href={resolve(`/handbook?entry=${handbook}`)}
	onclick={linkClicked}
	use:tooltip={{
		text: "Open Handbook",
		pos: "top",
	}}
	aria-label={`Open handbook for ${handbook}`}
>
	help
</a>

<style lang="scss">
	a {
		all: unset;
		cursor: pointer;
		font-weight: normal;
		font-size: 11px;
		background-color: $accent-color;
		padding: 3px 5px;
		border-radius: 5px;

		&:hover,
		&:focus-visible {
			color: $bg-color;
			background-color: $accent-color-hover;
		}
	}
</style>
