<script lang="ts">
  import type { PublicUser } from "@/types";

  export let users: PublicUser[];
</script>

<ul>
  {#each users as user}
    <li><a href="/lists/{user.id}/{user.username}">{user.username}</a></li>
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
      min-width: 50px;

      a {
        display: flex;
        justify-content: center;
        width: 100%;
        height: 100%;
        padding: 3px 8px;
        border-radius: 4px;
        background-color: $accent-color;
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
