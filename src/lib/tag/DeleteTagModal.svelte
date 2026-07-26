<script lang="ts">
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
			error = "Tag doesn't have an id!";
			return;
		}
		deleteDisabled = true;
		const nid = notify({ text: "Deleting Tag", type: "loading" });
		try {
			const resp = await req.delete(`/tag/${tag.id}`);
			console.log("deleteTag: Tag was deleted", resp);
			// Remove tag from store.
			store.tags = store.tags.filter((t) => t.id !== tag.id);
			notify({ id: nid, text: "Tag Deleted!", type: "success" });
			onClose();
		} catch (err) {
			console.error("deleteTag: Failed!", err);
			notify({ id: nid, text: "Failed!", type: "error", time: 1 });
			error = "Failed To Delete!";
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
	title="Permanently Delete Tag"
	desc="Are you sure you want to delete this tag?"
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
			Yes, fully delete (unrecoverable)
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
