<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import Modal from "../Modal.svelte";
	import type { Signal } from "../util/signal";

	interface Props {
		onDecision: Signal<boolean>;
	}

	let { onDecision }: Props = $props();
</script>

<Modal
	title={t("watched.countAsPlay.title")}
	desc={t("watched.countAsPlay.desc")}
	maxWidth="450px"
>
	<div class="button-row">
		<button class="danger" onclick={() => onDecision.fire(false)}>
			{t("watched.countAsPlay.no")}
		</button>
		<button onclick={() => onDecision.fire(true)}
			>{t("watched.countAsPlay.yes")}</button
		>
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
