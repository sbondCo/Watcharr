<script lang="ts">
	import Modal from "../Modal.svelte";
	import type { Signal } from "../util/signal";

	interface Props {
		onDecision: Signal<boolean>;
	}

	let { onDecision }: Props = $props();
</script>

<Modal
	title="Count as Whole Play"
	desc="Do you want to increment the Plays for this media by one?"
	maxWidth="450px"
>
	<div class="button-row">
		<button class="danger" onclick={() => onDecision.fire(false)}>
			No, Don't Count
		</button>
		<button onclick={() => onDecision.fire(true)}>Yes, Count as Play</button>
	</div>
</Modal>

<style lang="scss">
	.button-row {
		display: flex;
		flex-flow: row;
		justify-content: space-between;
		margin-top: 10px;

		button {
			margin-top: auto;
			width: max-content;
		}
	}
</style>
