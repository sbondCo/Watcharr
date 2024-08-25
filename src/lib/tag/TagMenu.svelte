<script lang="ts">
  import { tags } from "@/store";
  import Icon from "../Icon.svelte";
  import CreateTagModal from "./CreateTagModal.svelte";
  import type { Tag as TagT } from "@/types";
  import Tag from "./Tag.svelte";
  import DeleteTagModal from "./DeleteTagModal.svelte";

  export let titleText: string | undefined = undefined;
  export let classes: string | undefined = undefined;
  export let onTagClick: (tag: TagT, remove: boolean) => void | undefined = undefined!;
  export let selectedTags: TagT[] | undefined = undefined;
  /**
   * When `showManageBtn` is true, a manage icon will appear at top
   * of menu for the user to click. When toggled on, clicking a tag
   * will trigger a deletion instead of `onTagClick()`.
   */
  export let showManageBtn = false;

  $: allTags = $tags;

  let tagModalOpen = false;
  let inManageMode = false;
  let tagToDelete: TagT | undefined = undefined;

  function deleteTag(t: TagT) {
    // TODO 1. confirmation 2. run delete req 3. remove from local state & from watcheds state
    tagToDelete = t;
  }
</script>

<div class={[`menu`, classes].join(" ")}>
  <div class="inner">
    <div class="title">
      <h4 class="norm sm-caps">{titleText ? titleText : "My Tags"}</h4>
      {#if showManageBtn}
        <button
          class={["plain", inManageMode ? "manage-on" : ""].join(" ")}
          on:click={() => (inManageMode = !inManageMode)}
        >
          <Icon i="trash" wh={18} />
        </button>
      {/if}
      <button class="plain" on:click={() => (tagModalOpen = !tagModalOpen)}>
        <Icon i="add" wh={22} />
      </button>
    </div>
    {#if allTags && allTags.length > 0}
      {#if inManageMode}
        <strong style="font-size: 12px; margin-bottom: 10px;">Click a tag to delete it.</strong>
      {/if}
      <div class="list">
        {#each allTags as t}
          {@const isSelected = selectedTags
            ? selectedTags.find((tag) => tag.id === t.id)
              ? true
              : false
            : false}
          <div>
            <Tag
              tag={t}
              onClick={() => (inManageMode ? deleteTag(t) : onTagClick(t, isSelected))}
            />
            {#if isSelected}
              <Icon i="check" wh={18} />
            {/if}
          </div>
        {/each}
      </div>
    {:else}
      <span style="margin-top: 0;">You have no tags yet!</span>
    {/if}
  </div>
</div>

{#if tagModalOpen}
  <CreateTagModal onClose={() => (tagModalOpen = false)} />
{/if}

{#if tagToDelete}
  <DeleteTagModal tag={tagToDelete} onClose={() => (tagToDelete = undefined)} />
{/if}

<style lang="scss">
  div.menu {
    width: 200px;

    &:before {
      left: 42px;
    }

    &.from-add-to-tag-btn {
      top: 50px;
      right: -78px;

      &:before {
        left: 87px;
        /* The place where this button will be is always dark, so white works for both themes */
        border-bottom-color: white;
      }
    }
  }

  div.inner {
    h4 {
      color: $text-color;

      &:not(:first-of-type) {
        margin-top: 8px;
      }
    }

    .title {
      display: flex;
      flex-flow: row;
      align-items: center;
      margin-bottom: 8px;
      gap: 5px;

      button {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 28px;
        height: 26px;
        padding: 2px 3px;
        border-radius: 8px;

        &.manage-on {
          color: #f3555a;
          background-color: $text-color;
        }

        &:first-of-type {
          margin-left: auto;
        }
      }
    }

    .list {
      display: flex;
      flex-flow: column;
      gap: 5px;

      & > div {
        display: flex;
        align-items: center;
        gap: 5px;
        color: $text-color;

        :global(svg) {
          min-width: 18px;
        }
      }
    }
  }
</style>
