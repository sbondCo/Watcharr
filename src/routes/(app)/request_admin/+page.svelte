<script lang="ts">
  import axios from "axios";

  let page = 0;

  function generateAdminToken() {
    axios.get("/auth/admin_token");
  }
</script>

<div class="content">
  <div class="inner">
    <h2>Request Admin</h2>

    {#if page == 0}
      <button on:click={generateAdminToken}>Request</button>
    {:else if page == 1}
      <p>Check your sever log to view you token. Once you have it type it down below.</p>
      <input type="text" />
      <button>Check Token</button>
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
      min-width: 400px;
      max-width: 400px;
      overflow: hidden;
    }
  }
</style>
