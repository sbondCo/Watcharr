<script lang="ts">
  import { tags, watchedList } from "@/store";
  import Modal from "../Modal.svelte";
  import Setting from "../settings/Setting.svelte";
  import SettingsList from "../settings/SettingsList.svelte";
  import ColorSelector from "../ColorSelector.svelte";
  import { notify } from "../util/notify";
  import axios from "axios";
  import type { Tag as TagT, TagAddRequest } from "@/types";
  import { get } from "svelte/store";
  import Tag from "./Tag.svelte";
  import { onMount } from "svelte";

  export let tag: TagT;
  export let onClose: () => void;

  let error = "";
  let deleteDisabled = false;

  async function deleteTag() {
    console.debug("deleteTag:", tag);
    if (!tag || !tag.id) {
      error = "Tag doesn't have an id!";
      return;
    }
    deleteDisabled = true;
    const nid = notify({ text: "Deleting Tag", type: "loading" });
    try {
      const resp = await axios.delete(`/tag/${tag.id}`);
      console.log("deleteTag: Tag was deleted", resp.data);
      // 1. Remove tag from store.
      const _tags = get(tags);
      tags.update((t) => t);
      const newList = _tags.filter((t) => t.id !== tag.id);
      tags.update(() => newList);
      // 2. Remove tag from all watched entries in store.
      try {
        const wList = get(watchedList);
        for (let i = 0; i < wList.length; i++) {
          const wi = wList[i];
          if (wi.tags && wi.tags.length > 0) {
            wi.tags = wi.tags.filter((t) => t.id !== tag.id);
          }
        }
        watchedList.update(() => wList);
      } catch (err) {
        console.error("deleteTag: Failed to remove tags from watched entries:", err);
      }
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
    // Sort of prevent accidental clickage, wait 3s after opening modal before enabling delete btn.
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
    <button class="delete-tag-btn" on:click={() => deleteTag()} disabled={deleteDisabled}>
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
