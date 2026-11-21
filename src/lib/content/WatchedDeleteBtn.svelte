<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import tooltip from "../actions/tooltip";
	import Icon from "../Icon.svelte";
	import { removeWatched } from "../util/api";
	import { notify } from "../util/notify";
	import WatchedDeleteModal from "../watched/WatchedDeleteModal.svelte";

	interface Props {
		watchedId: number;
		mediaName?: string;
		onDelete: () => void;
	}

	let { watchedId, mediaName, onDelete }: Props = $props();

	// If confirm delete modal should be shown, then this will be
	// set to a callback.
	let showConfirmDeleteModalCallback: (() => void) | undefined = $state();

	function onDeleteClicked() {
		showConfirmDeleteModalCallback = () => {
			removeWatched(watchedId).then((removed) => {
				if (removed) {
					onDelete();
				}
			});
		};
	}
</script>

<button
	class="delete-btn"
	onclick={() => onDeleteClicked()}
	use:tooltip={{ text: t("common.delete"), pos: "bot" }}
>
	<Icon i="trash" wh={19} />
</button>

{#if showConfirmDeleteModalCallback !== undefined}
	<WatchedDeleteModal
		{mediaName}
		onClose={(confirmed) => {
			if (!showConfirmDeleteModalCallback) {
				notify({
					type: "error",
					text: t("content.deleteCallbackMissing"),
					time: 5000,
				});
				return;
			}
			if (confirmed) {
				showConfirmDeleteModalCallback();
			}
			showConfirmDeleteModalCallback = undefined;
		}}
	/>
{/if}

<style lang="scss">
	.delete-btn {
		&:hover {
			color: $error;
		}
	}
</style>
