<script lang="ts">
  import Icon from "@/lib/Icon.svelte";
  import Modal from "@/lib/Modal.svelte";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import Setting from "@/lib/settings/Setting.svelte";
  import SettingsList from "@/lib/settings/SettingsList.svelte";
  import { getOrdinalSuffix, monthsShort } from "@/lib/util/helpers";
  import type { ManagedUser } from "@/types";
  import axios from "axios";
  import EditUserModal from "./modals/EditUserModal.svelte";

  const currentYear = new Date(Date.now()).getFullYear();

  let editingUser: ManagedUser | undefined;

  async function getUsers() {
    return (await axios.get(`/server/users`)).data as ManagedUser[];
  }
</script>

<div class="content">
  <div class="inner">
    <h2>User Management</h2>
    <h5 class="norm">Take control and manage all the users using your server.</h5>

    {#await getUsers()}
      <Spinner />
    {:then users}
      <table>
        <tr>
          <th>Name</th>
          <th>Private</th>
          <th>Joined</th>
          <th></th>
        </tr>
        {#each users as u}
          {@const joinDate = new Date(u.createdAt)}
          <tr>
            <td class="username">
              <div class={`type-${u.type}`}>
                {#if u.type == 1}
                  <Icon i="jellyfin" wh={20} />
                {:else if u.type == 2}
                  <Icon i="plex" wh={20} />
                {:else}
                  <span
                    style="font-family: 'Rampart One'; font-weight: bold; font-size: 21px; line-height: 20px; user-select: none;"
                  >
                    W
                  </span>
                {/if}
                {u.username}
              </div>
            </td>
            <td>{u.private === true ? "Yes" : "No"}</td>
            <td>
              {joinDate.getDate()}{getOrdinalSuffix(joinDate.getDate())}
              {monthsShort[joinDate.getMonth()]}
              {joinDate.getFullYear() === currentYear
                ? ""
                : `'${String(joinDate.getFullYear()).substring(2, 4)}`}</td
            >
            <td>
              <button class="plain" on:click={() => (editingUser = u)}>
                <Icon i="chevron" facing="right" wh={24} />
              </button>
            </td>
          </tr>
        {/each}
      </table>

      {#if editingUser}
        <EditUserModal user={editingUser} onClose={() => (editingUser = undefined)} />
      {/if}
    {:catch err}
      <PageError error={err} pretty="Failed to fetch users!" />
    {/await}
  </div>
</div>

<style lang="scss">
  table {
    td {
      padding: 12px 15px;

      &.username > div {
        display: flex;
        align-items: center;
        height: 100%;
        gap: 10px;

        &.type-1 :global(svg) {
          fill: $text-color;
        }
      }

      &:has(button) {
        padding: 0;
      }

      button {
        margin: auto;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 12px 15px;

        &:hover :global(svg) {
          transform: rotate(180deg) translateX(-2px);
        }
      }
    }
  }

  .content {
    display: flex;
    width: 100%;
    justify-content: center;
    padding: 0 30px 30px 30px;

    .inner {
      min-width: 700px;
      max-width: 700px;
      overflow: hidden;

      h2 {
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
      }

      & > div:not(:first-of-type) {
        margin-top: 30px;
      }

      @media screen and (max-width: 740px) {
        width: 100%;
        min-width: unset;
      }
    }
  }
</style>
