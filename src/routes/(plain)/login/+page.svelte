<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Icon from "@/lib/Icon.svelte";
  import { UserType, type Icon as Icons, type AvailableAuthProviders } from "@/types";
  import { noAuthAxios } from "@/lib/util/api";
  import { onMount, afterUpdate } from "svelte";
  import { notify, unNotify } from "@/lib/util/notify";

  let error: string;
  let login = true;
  let availableProviders: string[] = [];
  let signupEnabled = true;
  let useEmby = false;

  onMount(() => {
    if (localStorage.getItem("token")) {
      goto("/");
    }

    noAuthAxios.get<AvailableAuthProviders>("/auth/available").then((r) => {
      if (r?.data) {
        if (r.data.isInSetup) {
          console.log("Server is in setup.. navigating to web setup page.");
          goto("/setup");
        }
        availableProviders = r.data.available;
        signupEnabled = r.data.signupEnabled;
        useEmby = r.data.useEmby;
      }
    });
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

    const nid = notify({ text: "Logging in", type: "loading" });
    noAuthAxios
      .post(`/auth${login ? `/${customAuthEP}` : "/register"}`, {
        username: user,
        password: pass
      })
      .then((resp) => {
        if (resp.data?.token) {
          console.log("Received token... logging in.");
          localStorage.setItem("token", resp.data.token);
          if (useEmby) {
            localStorage.setItem("useEmby", "1");
          } else {
            localStorage.removeItem("useEmby");
          }
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

  function proxyLogin() {
    const nid = notify({ text: "Logging in", type: "loading" });
    noAuthAxios
      .post(`/auth/proxy`)
      .then((resp) => {
        if (resp.data?.token) {
          console.log("Received token... logging in.");
          localStorage.setItem("token", resp.data.token);
          goto("/");
          notify({ id: nid, text: `Welcome!`, type: "success" });
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

  async function plexLogin() {
    try {
      const { preparePlexAuth, doPlexLogin, plexPinPoll } = await import("@/lib/util/plex");
      const p = preparePlexAuth();
      const pin = await doPlexLogin(p);
      plexPinPoll(pin, p, (err, token) => {
        if (err) {
          error = "Plex Auth Failed";
          console.error("Plex auth failed!", err);
          return;
        }
        const nid = notify({ text: "Logging in", type: "loading" });
        noAuthAxios
          .post("/auth/plex", {
            token,
            clientIdentifier: p.clientId
          })
          .then((resp) => {
            if (resp.data?.token) {
              console.log("Received token... logging in.");
              localStorage.setItem("token", resp.data.token);
              goto("/");
              notify({ id: nid, text: `Welcome!`, type: "success" });
            }
          })
          .catch((err) => {
            console.error("plexLogin: Fail", err);
            if (err.response) {
              error = err.response.data.error;
            } else {
              error = err.message;
            }
            notify({ id: nid, text: `Failed!`, type: "error" });
          });
      });
    } catch (err) {
      console.error("plexLogin: failed!", err);
      error = "Plex login failed";
    }
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
          {#if availableProviders?.length > 0}
            {#if availableProviders.find((ap) => ap === "jellyfin")}
              {#if useEmby}
                <button type="submit" name="jellyfin" class="other">
                  <Icon i="emby" wh={18} />
                  emby
                </button>
              {:else}
                <button type="submit" name="jellyfin" class="other">
                  <Icon i="jellyfin" wh={18} />
                  jellyfin
                </button>
              {/if}
            {/if}
          {/if}
          {#if availableProviders?.findIndex((provider) => provider == "proxy") > -1}
            <button type="button" name="proxy" class="other" on:click={() => { proxyLogin() }}>
                SSO
            </button>
          {/if}
        </div>
        {#if availableProviders?.findIndex((provider) => provider == "plex") > -1}
          <p style="font-weight: bold; font-size: 14px;">or</p>
          <div class="login-btns">
            <button
              type="button"
              on:click={() => {
                plexLogin();
              }}
              name="plex"
              class="plex other"
            >
              <Icon i="plex" wh={18} />Continue with Plex
            </button>
          </div>
        {/if}
      {:else}
        <div class="login-btns">
          <button type="submit">Sign Up</button>
        </div>
      {/if}
    </form>

    {#if signupEnabled}
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
    {/if}
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
      text-transform: capitalize;

      .watcharr {
        font-family: "Rampart One";
        font-size: 19px;
        line-height: 19px;
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
