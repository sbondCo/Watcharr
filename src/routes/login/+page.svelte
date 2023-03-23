<script lang="ts">
  import { goto } from "$app/navigation";
  import axios from "axios";

  let error: string;
  let login = true;

  function handleLogin(ev: SubmitEvent) {
    const fd = new FormData(ev.target! as HTMLFormElement);

    axios({
      baseURL: "http://127.0.0.1:3080/auth",
      url: login ? "/" : "/register",
      method: "POST",
      data: {
        username: fd.get("username"),
        password: fd.get("password")
      }
    })
      .then((resp) => {
        if (resp.data?.token) {
          console.log("Received token... logging in.");
          localStorage.setItem("token", resp.data.token);
          goto("/");
        }
      })
      .catch((err) => {
        if (err.response) {
          error = err.response.data.error;
        } else {
          error = err.message;
        }
      });
  }
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

    {#if error}
      <span class="error">{error}!</span>
    {/if}

    <form on:submit|preventDefault={handleLogin}>
      <label for="username">Username</label>
      <input type="text" name="username" placeholder="Username" />

      <label for="password">Password</label>
      <input type="text" name="password" placeholder="Password" />

      <button type="submit">
        {#if login}
          Login
        {:else}
          Sign Up
        {/if}
      </button>
    </form>

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
  div,
  form {
    display: flex;
    flex-flow: column;
    align-items: center;
    gap: 10px;
  }

  .inner,
  form {
    width: 100%;
    max-width: 400px;
  }

  label {
    align-self: flex-start;
    font-weight: bold;
  }

  .error {
    display: flex;
    justify-content: center;
    width: 100%;
    padding: 10px;
    background-color: rgb(221, 48, 48);
    text-transform: capitalize;
    color: white;
  }
</style>
