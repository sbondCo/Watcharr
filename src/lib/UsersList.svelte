<script lang="ts">
  import type { PublicUser } from "@/types";

  export let users: PublicUser[];
</script>

<ul>
  {#each users as user}
    <li title={user.username}><a href="/lists/{user.id}/{user.username}">{user.username}</a></li>
  {/each}
</ul>

<style lang="scss">
  ul {
    display: flex;
    flex-flow: row;
    gap: 10px;
    margin: 5px 20px;
    list-style: none;
    justify-content: center;

    li {
      display: flex;
      min-width: 50px;
      max-width: 200px;

      a {
        flex-shrink: 1;
        width: 100%;
        height: 100%;
        padding: 3px 8px;
        border-radius: 4px;
        background-color: $accent-color;
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
        transition:
          background-color 150ms ease,
          color 150ms ease;

        &:hover {
          color: $bg-color;
          background-color: $accent-color-hover;
        }
      }
    }
  }
</style>
