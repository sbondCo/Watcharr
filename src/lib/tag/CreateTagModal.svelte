<script lang="ts">
  import { tags } from "@/store";
  import Modal from "../Modal.svelte";
  import Setting from "../settings/Setting.svelte";
  import SettingsList from "../settings/SettingsList.svelte";
  import ColorSelector from "../ColorSelector.svelte";
  import { notify } from "../util/notify";
  import axios from "axios";
  import type { Tag, TagAddRequest } from "@/types";
  import { get } from "svelte/store";
  import { onMount } from "svelte";

  export let onClose: () => void;
  // Passing an existing tag will enable 'Edit Tag' mode.
  export let existingTag: Tag | undefined = undefined;

  const colorPresets = [
    ["#36BA98", "#000"],
    ["#FF4191", "#231731"],
    ["#173B45", "#FFE0C9"],
    ["#6EACDA", "#021526"],
    ["#180161", "#FF9E73"],
    ["#FFBC0A", "#2F0A39"],
    ["#C1FF9B", "#321673"],
    ["#F5FFC6", "#5126B5"],
    ["#6D213C", "#FAFF70"],
    ["#F7C4A5", "#2f0082"],
    ["#F7EFE5", "#63338c"]
  ];

  function getRandomColorPreset() {
    return colorPresets[Math.floor(Math.random() * colorPresets.length)];
  }

  const defaultPreset = getRandomColorPreset();
  let textColor = defaultPreset[1];
  let bgColor = defaultPreset[0];
  let tagName = "";
  let error = "";
  let submitDisabled = false;
  let modalTitle = "Create A Tag";
  let modalDesc = "Create a new tag";
  let submitBtnText = "Create Tag";

  async function addTag() {
    console.debug("addTag:", tagName, textColor, bgColor);
    if (!tagName) {
      error = "Tag must have a name!";
      return;
    }
    const nid = notify({ text: "Creating Tag", type: "loading" });
    try {
      const resp = await axios.post<Tag>("/tag", {
        name: tagName,
        color: textColor,
        bgColor
      } as TagAddRequest);
      console.log("addTag: Tag was created", resp.data);
      const _tags = get(tags);
      _tags.push(resp.data);
      tags.update((t) => t);
      notify({ id: nid, text: "Tag Created!", type: "success" });
      onClose();
    } catch (err) {
      console.error("addTag: Failed!", err);
      notify({ id: nid, text: "Failed!", type: "error", time: 1 });
      error = "Failed!";
    }
  }

  async function updateTag() {
    console.debug("updateTag:", existingTag, tagName, textColor, bgColor);
    if (!tagName) {
      error = "Tag must have a name!";
      return;
    }
    const nid = notify({ text: "Modifying Tag", type: "loading" });
    try {
      const resp = await axios.put<Tag>(`/tag/${existingTag!.id}`, {
        name: tagName,
        color: textColor,
        bgColor
      } as TagAddRequest);
      console.log("updateTag: Tag was edited", resp.data);
      existingTag!.name = tagName;
      existingTag!.color = textColor;
      existingTag!.bgColor = bgColor;
      // Doesn't update `updatedAt`... may need to in the future if we need to sort by it, etc
      tags.update((t) => t);
      notify({ id: nid, text: "Tag Modified!", type: "success" });
      onClose();
    } catch (err) {
      console.error("updateTag: Failed!", err);
      notify({ id: nid, text: "Failed!", type: "error", time: 1 });
      error = "Failed!";
    }
  }

  async function submitClicked() {
    submitDisabled = true;
    try {
      if (existingTag) {
        await updateTag();
      } else {
        await addTag();
      }
    } catch (err) {
      console.log("CreateTagModal: Submit failed!", err);
    }
    submitDisabled = false;
  }

  onMount(() => {
    if (existingTag) {
      console.log("CreateTagModal: Entering edit mode for tag:", existingTag);
      modalTitle = "Edit Tag";
      modalDesc = "Edit an existing tag";
      submitBtnText = "Edit Tag";
      tagName = existingTag.name;
      textColor = existingTag.color;
      bgColor = existingTag.bgColor;
    }
  });
</script>

<div class="wrap">
  <Modal title={modalTitle} desc={modalDesc} maxWidth="500px" {onClose} {error}>
    <SettingsList>
      <Setting title="Name" desc="What should we call this tag?">
        <input type="text" name="name" placeholder="Name" bind:value={tagName} />
      </Setting>
      <Setting title="Text Color" desc="Color for your tags text." row>
        <ColorSelector bind:value={textColor} style="max-width: 150px;" />
      </Setting>
      <Setting title="Background Color" desc="Color for your tags background." row>
        <ColorSelector bind:value={bgColor} style="max-width: 150px;" />
      </Setting>
      <button class="add-tag-btn" on:click={() => submitClicked()} disabled={submitDisabled}>
        {submitBtnText}
      </button>
    </SettingsList>
  </Modal>
</div>

<style lang="scss">
  .add-tag-btn {
    width: max-content;
    margin-left: auto;
  }

  .wrap {
    color: $text-color;
  }
</style>
