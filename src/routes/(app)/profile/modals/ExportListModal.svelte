<script lang="ts">
	import Modal from "@/lib/Modal.svelte";
	import { notify } from "@/lib/util/notify";
	import axios from "axios";

	interface Props {
		onClose: () => void;
	}

	let { onClose }: Props = $props();

	async function downloadWatchedList() {
		const nid = notify({ text: "Exporting", type: "loading" });
		try {
			// We re-fetch, to ensure data we export is up to date.
			const r = await axios.get("/watched");
			console.log(r.data);
			if (!r.data || r.data?.length <= 0) {
				notify({
					id: nid,
					text: "Can't export an empty watch list!",
					type: "error",
					time: 10000,
				});
				return;
			}
			const file = new Blob([JSON.stringify(r.data, undefined, 2)], {
				type: "application/json",
			});
			const a = document.createElement("a");
			a.href = URL.createObjectURL(file);
			a.download = "watcharr-export.json";
			a.click();
			notify({ id: nid, text: "Successfully Exported", type: "success" });
			onClose();
		} catch (err) {
			console.error("downloadWatchedList failed!", err);
			notify({ id: nid, text: "Export Failed!", type: "error" });
		}
	}
</script>

<Modal title="Export Watched List" maxWidth="600px" {onClose}>
	<p>
		Only use this feature if you are leaving Watcharr to use another platform OR
		leaving to another Watcharr instance where you don't own this or the other
		instance.
	</p>
	<p>
		If you are migrating to another server that you own, it's best to take your
		existing database with you.
	</p>
	<p>
		<b>Warning</b>: This is not a backup feature. Backups should be done on the
		server (see:
		<a href="https://watcharr.app/docs/server_config/backup" target="_blank"
			>https://watcharr.app/docs/server_config/backup</a
		>).
	</p>
	<button onclick={() => downloadWatchedList()}>Export</button>
</Modal>

<style lang="scss">
	p {
		margin-bottom: 10px;

		&:first-of-type {
			margin-top: 10px;
		}
	}
</style>
