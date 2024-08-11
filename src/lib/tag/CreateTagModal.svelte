<script lang="ts">
  import { tags } from "@/store";
  import Modal from "../Modal.svelte";
  import Setting from "../settings/Setting.svelte";
  import SettingsList from "../settings/SettingsList.svelte";
  import ColorSelector from "../ColorSelector.svelte";

  $: allTags = $tags;

  const colorPresets = [
    ["#36BA98", "#F5F7F8"],
    ["#FF4191", "#FFF078"],
    ["#173B45", "#FF8225"],
    ["#6EACDA", "#021526"],
    ["#180161", "#FB773C"],
    ["#FFBC0A", "#C200FB"],
    ["#C1FF9B", "#AB87FF"],
    ["#F5FFC6", "#AB87FF"],
    ["#6D213C", "#FAFF70"],
    ["#F7C4A5", "#605770"],
    ["#F7EFE5", "#674188"]
  ];

  function getRandomColorPreset() {
    return colorPresets[Math.floor(Math.random() * colorPresets.length)];
  }

  const defaultPreset = getRandomColorPreset();
  let textColor = defaultPreset[1];
  let bgColor = defaultPreset[0];

  function addTag() {}
</script>

<Modal title="Create A Tag" desc="Create a new tag" maxWidth="500px">
  <SettingsList>
    <Setting title="Name" desc="Name for your tag.">
      <input type="text" name="name" placeholder="Name" />
    </Setting>
    <Setting title="Text Color" desc="Color for your tags text." row>
      <ColorSelector bind:value={textColor} style="max-width: 150px;" />
    </Setting>
    <Setting title="Background Color" desc="Color for your tags background." row>
      <ColorSelector bind:value={bgColor} style="max-width: 150px;" />
    </Setting>
    <button class="add-tag-btn" on:click={() => addTag()}>Create Tag</button>
  </SettingsList>
</Modal>

<style lang="scss">
  .add-tag-btn {
    width: max-content;
    margin-left: auto;
  }
</style>
