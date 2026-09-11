<script lang="ts">
	import Modal from "../Modal.svelte";
	import HandbookPage from "@/routes/(app)/handbook/+page.svelte";

	interface Props {
		onClose: () => void;
	}

	let { onClose }: Props = $props();
</script>

<Modal title="Handbook" {onClose} maxWidth="1200px">
	<div class="content">
		<div class="inner">
			<HandbookPage />
		</div>
	</div>
</Modal>

<style lang="scss">
</style>
