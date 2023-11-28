<script lang="ts">
  import type { PublicUser } from "@/types";

  export let users: PublicUser[];
</script>

<div>
  <h2>Users</h2>
  <ul>
    {#each users as user}
      <li title={user.username}>
        <a href="/lists/{user.id}/{user.username}">
          <span>{user.username}</span>
        </a>
      </li>
    {/each}
  </ul>
</div>

<style lang="scss">
  div {
    display: flex;
    flex-flow: column;
    margin: 5px 15px;
  }

  ul {
    display: flex;
    flex-flow: row;
    gap: 10px;
    list-style: none;
    overflow: auto;
    padding: 15px;

    li {
      display: flex;
      max-width: 200px;

      a {
        display: flex;
        flex-flow: column;
        gap: 8px;
        width: 100%;
        height: 100%;
        padding: 3px 8px;
        padding: 10px 20px;
        border-radius: 8px;
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
