<script lang="ts">
  import { tags } from "@/store";
  import Icon from "../Icon.svelte";
  import CreateTagModal from "../tag/CreateTagModal.svelte";

  function tagClicked() {}

  $: allTags = $tags;
</script>

<div class={`menu`}>
  <div class="inner">
    <div class="title">
      <h4 class="norm sm-caps">My Tags</h4>
      <button class="plain"><Icon i="add" wh={22} /></button>
    </div>
    {#if allTags && allTags.length > 0}
      {#each allTags as t}
        <button class={`plain`} on:click={() => tagClicked()}>
          {t.name}
        </button>
      {/each}
    {:else}
      <span>You have no tags yet!</span>
    {/if}
  </div>
</div>

<CreateTagModal />

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

    & > button {
      text-transform: capitalize;
      position: relative;
      width: max-content;
      border-radius: 8px;
    }
  }
</style>
