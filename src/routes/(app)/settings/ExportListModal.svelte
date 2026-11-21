<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import Trans from "@/lib/i18n/Trans.svelte";
	import Modal from "@/lib/Modal.svelte";
	import { req } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";

	interface Props {
		onClose: () => void;
	}

	let { onClose }: Props = $props();

	async function downloadWatchedList() {
		const nid = notify({ text: t("export.exporting"), type: "loading" });
		try {
			// We re-fetch, to ensure data we export is up to date.
			const r = await req.get<unknown[]>("/watched");
			console.log(r);
			if (!r || r?.length <= 0) {
				notify({
					id: nid,
					text: t("export.emptyList"),
					type: "error",
					time: 10000,
				});
				return;
			}
			const file = new Blob([JSON.stringify(r, undefined, 2)], {
				type: "application/json",
			});
			const a = document.createElement("a");
			a.href = URL.createObjectURL(file);
			a.download = "watcharr-export.json";
			a.click();
			notify({ id: nid, text: t("export.success"), type: "success" });
			onClose();
		} catch (err) {
			console.error("downloadWatchedList failed!", err);
			notify({ id: nid, text: t("export.failed"), type: "error" });
		}
	}
</script>

<Modal title={t("export.title")} maxWidth="600px" {onClose}>
	<p>
		{t("export.p1")}
	</p>
	<p>
		{t("export.p2")}
	</p>
	<p>
		<Trans key="export.warning">
			{#snippet link(text: string)}
				<a href="https://watcharr.app/docs/server_config/backup" target="_blank"
					>{text}</a
				>
			{/snippet}
		</Trans>
	</p>
	<button onclick={() => downloadWatchedList()}>{t("settings.export")}</button>
</Modal>

<style lang="scss">
	p {
		margin-bottom: 10px;

		&:first-of-type {
			margin-top: 10px;
		}
	}
</style>
