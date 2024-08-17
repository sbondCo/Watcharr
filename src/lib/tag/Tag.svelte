<script lang="ts">
  import type { Tag } from "@/types";
  import { onMount } from "svelte";

  export let tag: Tag;
  export let onClick: () => void | undefined = undefined!;

  let tagBtn: HTMLButtonElement;

  /**
   * Default on tag clicked action.
   * Goes to tag.
   */
  function defaultOnClick() {
    // TODO navigate to tag page
    console.debug("Tag: Navigating to tag page.");
  }

  onMount(() => {
    if (tagBtn) {
      if (tag.color) {
        tagBtn.style.color = tag.color;
      }
      if (tag.bgColor) {
        tagBtn.style.background = tag.bgColor;
      }
    }
  });
</script>

<button
  bind:this={tagBtn}
  class={`plain`}
  on:click={() => {
    if (typeof onClick === "function") {
      onClick();
    } else {
      defaultOnClick();
    }
  }}
>
  {tag.name}
</button>

<style lang="scss">
  button {
    text-transform: capitalize;
    position: relative;
    width: max-content;
    border-radius: 8px;
    padding: 5px 8px;
    transition: opacity 150ms ease-in-out;

    &:hover {
      opacity: 0.8;
    }
  }
</style>
