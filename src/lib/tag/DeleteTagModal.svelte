<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import { store } from "@/store.svelte";
	import Modal from "../Modal.svelte";
	import { notify } from "../util/notify";
	import type { Tag as TagT } from "@/types";
	import Tag from "./Tag.svelte";
	import { onMount } from "svelte";
	import { req } from "../util/api";

	interface Props {
		tag: TagT;
		onClose: () => void;
	}

	let { tag, onClose }: Props = $props();

	let error = $state("");
	let deleteDisabled = $state(false);

	async function deleteTag() {
		console.debug("deleteTag:", tag);
		if (!tag || !tag.id) {
			error = t("tags.noId");
			return;
		}
		deleteDisabled = true;
		const nid = notify({ text: t("tags.deleting"), type: "loading" });
		try {
			const resp = await req.delete(`/tag/${tag.id}`);
			console.log("deleteTag: Tag was deleted", resp);
			// Remove tag from store.
			store.tags = store.tags.filter((t) => t.id !== tag.id);
			notify({ id: nid, text: t("tags.deleted"), type: "success" });
			onClose();
		} catch (err) {
			console.error("deleteTag: Failed!", err);
			notify({ id: nid, text: t("common.failed"), type: "error", time: 1 });
			error = t("common.failedToDelete");
		}
		deleteDisabled = false;
	}

	onMount(() => {
		// Sort of prevent accidental clickage, wait 3s after opening modal
		// before enabling delete btn.
		deleteDisabled = true;
		setTimeout(() => {
			deleteDisabled = false;
		}, 3000);
	});
</script>

<Modal
	title={t("tags.delete.title")}
	desc={t("tags.delete.desc")}
	maxWidth="500px"
	{onClose}
	{error}
>
	<div class="inner">
		<Tag {tag} />
		<button
			class="delete-tag-btn"
			onclick={() => deleteTag()}
			disabled={deleteDisabled}
		>
			{t("tags.delete.confirm")}
		</button>
	</div>
</Modal>

<style lang="scss">
	.delete-tag-btn {
		width: max-content;
		margin-left: auto;

		&:hover {
			color: $error;
		}
	}

	.inner {
		display: flex;
		align-items: center;
		justify-content: center;
		flex-wrap: wrap;
		gap: 10px;
		max-width: 100%;
		width: 100%;
	}
</style>
