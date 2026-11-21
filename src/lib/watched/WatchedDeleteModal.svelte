<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import Modal from "../Modal.svelte";
	import Trans from "../i18n/Trans.svelte";

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
	title={t("watched.delete.title")}
	onClose={() => onClose(false)}
	maxWidth="650px"
>
	<div class="remove-ctr">
		<p>
			<Trans key="watched.delete.desc" values={{ name: mediaName }} />
		</p>
		<button class="confirm-delete-btn" onclick={() => onClose(true)}>
			{t("watched.delete.confirm")}
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
