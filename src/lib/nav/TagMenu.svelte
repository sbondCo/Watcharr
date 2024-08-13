<script lang="ts">
  import { tags } from "@/store";
  import Icon from "../Icon.svelte";
  import CreateTagModal from "../tag/CreateTagModal.svelte";
  import Tag from "../tag/Tag.svelte";

  $: allTags = $tags;

  let tagModalOpen = false;
</script>

<div class={`menu`}>
  <div class="inner">
    <div class="title">
      <h4 class="norm sm-caps">My Tags</h4>
      <button class="plain" on:click={() => (tagModalOpen = !tagModalOpen)}>
        <Icon i="add" wh={22} />
      </button>
    </div>
    {#if allTags && allTags.length > 0}
      <div class="list">
        {#each allTags as t}
          <Tag tag={t} />
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

<style lang="scss">
  div.menu {
    width: 200px;
    right: 52px;

    &:before {
      left: 3px;
    }
  }

  div.inner {
    h4 {
      &:not(:first-of-type) {
        margin-top: 8px;
      }
    }

    .title {
      display: flex;
      flex-flow: row;
      align-items: center;
      margin-bottom: 8px;

      button {
        display: flex;
        align-items: center;
        justify-content: center;
        width: min-content;
        height: min-content;
        margin-left: auto;
        padding: 2px 3px;
        border-radius: 8px;
      }
    }

    .list {
      display: flex;
      flex-flow: column;
      gap: 5px;
    }
  }
</style>
