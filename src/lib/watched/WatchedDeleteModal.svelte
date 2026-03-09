<script lang="ts">
	import Modal from "../Modal.svelte";

	interface Props {
		mediaName?: string;
		/**
		 * If `confirmed=true` then the deletion request was confirmed.
		 */
		onClose: (confirmed: boolean) => void;
	}

	let { mediaName, onClose }: Props = $props();
</script>

<Modal
	title="Remove From Your List?"
	onClose={() => onClose(false)}
	maxWidth="650px"
>
	<div class="remove-ctr">
		<p>
			This will delete <b>{mediaName}</b> from your main list. Data will be lost.
		</p>
		<button class="confirm-delete-btn" onclick={() => onClose(true)}>
			Remove from my main list.
		</button>
	</div>
</Modal>

<style lang="scss">
	div.remove-ctr {
		display: flex;
		flex-flow: column;
		gap: 10px;

		button.confirm-delete-btn {
			width: max-content;
			margin-left: auto;

			&:hover {
				color: $error;
			}
		}
	}
</style>
