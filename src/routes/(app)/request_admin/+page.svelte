<script lang="ts">
  import { goto } from "$app/navigation";
  import { notify } from "@/lib/util/notify";
  import { userInfo } from "@/store";
  import { UserPermission } from "@/types";
  import axios from "axios";
  import { get } from "svelte/store";

  let page = 0;
  let adminToken: string;

  function generateAdminToken() {
    axios
      .get("/auth/admin_token")
      .then(() => {
        page = 1;
      })
      .catch((err) => {
        console.error("Failed to generate admin token", err);
        notify({ type: "error", text: "Failed to generate a token" });
      });
  }

  function useAdminToken() {
    if (!adminToken) {
      console.error("useAdminToken: No admin token provided!");
      return;
    }
    axios
      .post("/auth/admin_token", { token: adminToken })
      .then(() => {
        notify({ type: "success", text: "You now have admin!" });
        const uinf = get(userInfo);
        if (uinf) {
          uinf.permissions = UserPermission.PERM_ADMIN;
          userInfo.update((ui) => (ui = uinf));
        }
        goto("/");
      })
      .catch((err) => {
        console.error("Failed to use admin token", err?.response?.data?.error || err);
        notify({ type: "error", text: "Failed to use admin token!" });
      });
  }
</script>

<div class="content">
  <div class="inner">
    <h2>Request Admin</h2>

    {#if page == 0}
      <button on:click={generateAdminToken}>Request</button>
    {:else if page == 1}
      <p>Check your sever log to view you token. Once you have it type it down below.</p>
      <input bind:value={adminToken} type="text" placeholder="Admin Token" />
      <button on:click={useAdminToken}>Check Token</button>
    {/if}
  </div>
</div>

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;
    padding: 0 30px 0 30px;

    .inner {
      display: flex;
      flex-flow: column;
      gap: 10px;
      min-width: 400px;
      max-width: 400px;
      overflow: hidden;
    }
  }
</style>
