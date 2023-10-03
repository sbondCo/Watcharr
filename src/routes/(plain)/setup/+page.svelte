<script lang="ts">
  import { goto } from "$app/navigation";
  import type { AvailableAuthProviders } from "@/types";
  import { noAuthAxios } from "@/lib/util/api";
  import { onMount } from "svelte";
  import { notify, unNotify } from "@/lib/util/notify";

  let error: string;

  onMount(() => {
    if (localStorage.getItem("token")) {
      goto("/");
    }

    noAuthAxios.get<AvailableAuthProviders>("/auth/available").then((r) => {
      if (r?.data) {
        if (!r?.data?.isInSetup) {
          console.log("Server not in setup.. navigating to login page.");
          goto("/login");
        }
      }
    });
  });

  function handleLogin(ev: SubmitEvent) {
    const fd = new FormData(ev.target! as HTMLFormElement);
    const user = fd.get("username");
    const pass = fd.get("password");

    if (!user || !pass) {
      error = "Username and Password fields are required";
      return;
    }

    const nid = notify({ text: "Setting Up Admin User", type: "loading" });
    noAuthAxios
      .post("/setup/create_admin", {
        username: user,
        password: pass
      })
      .then((resp) => {
        if (resp.data?.token) {
          console.log("Received token... logging in.");
          localStorage.setItem("token", resp.data.token);
          localStorage.setItem("username", String(user));
          goto("/");
          notify({ id: nid, text: `Welcome ${user}!`, type: "success" });
        }
      })
      .catch((err) => {
        if (err.response) {
          error = err.response.data.error;
        } else {
          error = err.message;
        }
        unNotify(nid);
      });
  }
</script>

<svelte:head>
  <title>Setup Watcharr</title>
</svelte:head>

<div>
  <div class="inner">
    <div class="headers">
      <h2>Setup Admin User</h2>
      <h5 class="norm">Welcome to Watcharr! Setup your admin acount below.</h5>
    </div>

    {#if error}
      <span class="error">{error}!</span>
    {/if}

    <form on:submit|preventDefault={handleLogin}>
      <label for="username">Username</label>
      <input type="text" name="username" placeholder="Username" />

      <label for="password">Password</label>
      <input type="password" name="password" placeholder="Password" />

      <div class="login-btns">
        <button type="submit"><span class="watcharr">W</span>Set Up</button>
      </div>
    </form>
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

  .headers {
    display: flex;
    flex-flow: column;
    width: 100%;
    gap: 0;
    margin-bottom: 10px;
  }

  label {
    align-self: flex-start;
    font-weight: bold;
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
