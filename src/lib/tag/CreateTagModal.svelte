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

  export let onClose: () => void;

  $: allTags = $tags;

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

  async function addTag() {
    console.debug("addTag:", tagName, textColor, bgColor);
    if (!tagName) {
      error = "Tag must have a name!";
      return;
    }
    submitDisabled = true;
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
    submitDisabled = false;
  }
</script>

<div class="wrap">
  <Modal title="Create A Tag" desc="Create a new tag" maxWidth="500px" {onClose} {error}>
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
      <button class="add-tag-btn" on:click={() => addTag()} disabled={submitDisabled}>
        Create Tag
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
