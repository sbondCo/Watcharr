<script lang="ts">
  let login = true;
</script>

<div>
  <div class="inner">
    <h2>
      {#if login}
        Get Back In!
      {:else}
        Lucky You Found Us!
      {/if}
    </h2>
    <label for="username">Username</label>
    <input type="text" name="username" placeholder="Username" />

    <label for="password">Password</label>
    <input type="text" name="password" placeholder="Password" />

    <button>
      {#if login}
        Login
      {:else}
        Sign Up
      {/if}
    </button>

    <button
      class="plain"
      on:click={() => {
        login = !login;
      }}
    >
      {#if login}
        Not a user?
      {:else}
        Already a user?
      {/if}
    </button>
  </div>
</div>

<style lang="scss">
  div {
    display: flex;
    flex-flow: column;
    align-items: center;
    gap: 10px;

    .inner {
      width: 100%;
      max-width: 400px;
    }

    label {
      align-self: flex-start;
      font-weight: bold;
    }
  }
</style>
