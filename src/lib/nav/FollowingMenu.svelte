<script lang="ts">
  import { follows } from "@/store";

  $: following = $follows;
</script>

<div class="menu">
  {#if following?.length > 0}
    <h4 class="norm sm-caps">followed</h4>
    {#each following as f}
      <a href="/lists/{f.followedUser.id}/{f.followedUser.username}">
        {f.followedUser.username}
      </a>
    {/each}
  {:else}
    <span>no follows</span>
  {/if}
</div>

<style lang="scss">
  div {
    width: 180px;

    &:before {
      right: 53px;
    }

    a {
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
  }
</style>
