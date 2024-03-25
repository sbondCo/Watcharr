<script lang="ts">
  import Checkbox from "@/lib/Checkbox.svelte";
  import Modal from "@/lib/Modal.svelte";
  import Setting from "@/lib/settings/Setting.svelte";
  import SettingsList from "@/lib/settings/SettingsList.svelte";
  import { userHasPermission } from "@/lib/util/helpers";
  import { notify } from "@/lib/util/notify";
  import { UserPermission, type ManagedUser } from "@/types";
  import axios from "axios";

  interface UpdateUserRequest {
    permissions?: number;
  }

  export let user: ManagedUser;
  export let onClose: () => void;

  let error: string;
  let formDisabled = false;

  // Things we have changed
  let changedPerms = false;

  async function save() {
    // If nothing changed.. error
    if (!changedPerms) {
      error = "Nothing has been changed";
      return;
    }
    if (!error) {
      try {
        const toUpdate: UpdateUserRequest = {};
        if (changedPerms) {
          toUpdate["permissions"] = user.permissions;
        }
        const res = await axios.post(`/server/users/${user.id}`, toUpdate);
        if (res.status === 200) {
          notify({
            type: "success",
            text: "Changes saved!"
          });
          onClose();
        }
      } catch (err: any) {
        console.error("Failed to save user!", err);
        error = `Failed to save`;
        if (err?.response?.data?.error) {
          error = err.response.data.error;
        }
      }
    }
  }

  function userTogglePermission(perm: UserPermission) {
    user.permissions ^= perm;
    changedPerms = true;
  }
</script>

<Modal title={`Edit User`} desc={`Configuring ${user.username}`} maxWidth="500px" {onClose}>
  {#if error}
    <span class="error">{error}!</span>
  {/if}

  <h3 class="norm">Permissions</h3>

  <SettingsList>
    <Setting title="Admin" desc="Give user admin, overrides all other permissions." row>
      <Checkbox
        name="USER_PERM_ADMIN"
        value={userHasPermission(user.permissions, UserPermission.PERM_ADMIN)}
        toggled={() => {
          userTogglePermission(UserPermission.PERM_ADMIN);
        }}
      />
    </Setting>

    <Setting title="Request Content" desc="Give user permission to request content." row>
      <Checkbox
        name="USER_PERM_REQUEST_CONTENT"
        value={userHasPermission(user.permissions, UserPermission.PERM_REQUEST_CONTENT)}
        toggled={() => {
          userTogglePermission(UserPermission.PERM_REQUEST_CONTENT);
        }}
      />
    </Setting>

    <div class="btns">
      <button on:click={() => save()}>Save</button>
    </div>
  </SettingsList>
</Modal>

<style lang="scss">
  h3 {
    margin-bottom: 10px;
  }

  .btns {
    display: flex;
    flex-flow: row;
    gap: 10px;

    :first-child {
      margin-left: auto;
    }

    button {
      width: max-content;
      padding-left: 15px;
      padding-right: 15px;
    }
  }

  .error {
    position: sticky;
    top: 0;
    display: flex;
    justify-content: center;
    width: 100%;
    padding: 10px;
    background-color: rgb(221, 48, 48);
    text-transform: capitalize;
    color: white;
    margin-bottom: 15px;
  }
</style>
