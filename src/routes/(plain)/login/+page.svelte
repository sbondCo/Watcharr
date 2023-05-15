<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Icon from "@/lib/Icon.svelte";
  import type { Icon as Icons } from "@/types";
  import { noAuthAxios } from "@/lib/api";
  import { onMount, afterUpdate } from "svelte";

  let error: string;
  let login = true;

  onMount(() => {
    if (localStorage.getItem("token")) {
      goto("/");
    }
  });

  afterUpdate(() => {
    if (!error && $page.url.searchParams.get("again")) {
      error = "Please Login Again";
    }
  });

  function handleLogin(ev: SubmitEvent) {
    const fd = new FormData(ev.target! as HTMLFormElement);
    const user = fd.get("username");
    const pass = fd.get("password");

    if (!user || !pass) {
      error = "Username and Password fields are required";
      return;
    }

    let customAuthEP = "";
    if ((ev.submitter as HTMLButtonElement)?.name === "jellyfin") {
      customAuthEP = "jellyfin";
    }

    noAuthAxios
      .post(`/auth${login ? `/${customAuthEP}` : "/register"}`, {
        username: fd.get("username"),
        password: fd.get("password")
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

  async function getLoginProviders() {
    return (await noAuthAxios.get("/auth/available")).data as Icons[];
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
      <input type="password" name="password" placeholder="Password" />

      {#if login}
        <span class="login-with" style="font-weight: bold">Login With</span>
        <div class="login-btns">
          <button type="submit"><span class="watcharr">W</span>Watcharr</button>
          {#await getLoginProviders() then providers}
            {#each providers as p}
              <button type="submit" name="jellyfin" class="other"><Icon i={p} wh={18} />{p}</button>
            {/each}
          {/await}
        </div>
      {:else}
        <div class="login-btns">
          <button type="submit">Sign Up</button>
        </div>
      {/if}
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
    margin: 0 35px;
  }

  .inner,
  form {
    width: 100%;
    max-width: 400px;
  }

  .inner h2 {
    font-weight: normal;
  }

  label {
    align-self: flex-start;
    font-weight: bold;
  }

  span.login-with {
    font-size: 14px;
  }

  .login-btns {
    display: flex;
    flex-flow: row;
    gap: 10px;
    width: 100%;

    button {
      display: flex;
      flex-flow: row;
      gap: 10px;

      .watcharr {
        font-family: "Rampart One";
        font-size: 18px;
        line-height: 18px;
      }

      &.other {
        overflow: hidden;
        animation: 250ms ease otherbtn;

        @keyframes otherbtn {
          from {
            width: 0px;
          }
          to {
            width: 100%;
          }
        }
      }
    }
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
