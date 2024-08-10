<script lang="ts">
  import { tags } from "@/store";
  import type { WLDetailedViewOption } from "@/types";

  function detailClicked(d: WLDetailedViewOption) {}

  $: allTags = $tags;
</script>

<div class={`menu`}>
  <div class="inner">
    <h4 class="norm sm-caps">My Tags</h4>
    {#if allTags && allTags.length > 0}
      {#each allTags as t}
        <button class={`plain`} on:click={() => detailClicked("statusRating")}>
          {t.name}
        </button>
      {/each}
    {:else}
      <span>You have no tags yet!</span>
    {/if}
  </div>
</div>

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
      margin-bottom: 8px;

      &:not(:first-of-type) {
        margin-top: 8px;
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
