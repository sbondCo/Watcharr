<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Icon from "@/lib/Icon.svelte";
  import { UserType, type Icon as Icons, type AvailableAuthProviders } from "@/types";
  import { noAuthAxios } from "@/lib/util/api";
  import { onMount, afterUpdate } from "svelte";
  import { notify, unNotify } from "@/lib/util/notify";
  import { PlexOauth, type IPlexClientDetails } from "plex-oauth";

  let error: string;
  let login = true;
  let availableProviders: string[] = [];
  let signupEnabled = true;
  let plexPin: number;
  let plexOauthClient: PlexOauth;

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
        if (r.data.plexOauthId.length > 0) {
          let clientInformation: IPlexClientDetails = {
            clientIdentifier: r.data.plexOauthId,
            product: "Watcharr",
            device: "Watcharr",
            version: __WATCHARR_VERSION__
          };
          plexOauthClient = new PlexOauth(clientInformation);
        }
      }
    });
  });

  afterUpdate(() => {
    if (!error && $page.url.searchParams.get("again")) {
      error = "Please Login Again";
    }
  });

  function handleLogin(ev: SubmitEvent) {
    if ((ev.submitter as HTMLButtonElement)?.name === "plex") {
      showPlexOAuthWindow();
      return;
    }

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

  function showPlexOAuthWindow() {
    plexOauthClient
      .requestHostedLoginURL()
      .then((data) => {
        let [hostedUILink, pinId] = data;

        plexPin = pinId;

        let oAuthWindow = window.open(hostedUILink, "plex login", "poput,width=600,height=800");
        oAuthWindow?.focus();

        let pollTimer = window.setInterval(function () {
          if (oAuthWindow?.closed !== false) {
            window.clearInterval(pollTimer);
            getPlexToken();
          }
        }, 250);
      })
      .catch((err) => {
        throw err;
      });
  }

  function getPlexToken() {
    plexOauthClient
      .checkForAuthToken(plexPin)
      .then((authToken) => {
        if (authToken == null) {
          return;
        }
        const nid = notify({ text: "Logging in", type: "loading" });
        const endpoint = login ? "/auth/plex" : "/auth/register/plex";
        noAuthAxios
          .post(endpoint, {
            authtoken: authToken
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
            if (err.response) {
              error = err.response.data.error;
            } else {
              error = err.message;
            }
            unNotify(nid);
          });
      })
      .catch((err) => {
        throw err;
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
      <input type="password" name="password" placeholder="Password" />

      {#if login}
        <span class="login-with" style="font-weight: bold">Login With</span>
        <div class="login-btns">
          <button type="submit"><span class="watcharr">W</span>Watcharr</button>
          {#if availableProviders.findIndex((provider) => provider == "jellyfin") > -1}
            <button type="submit" name="jellyfin" class="jellyfin other">
              <Icon i="jellyfin" wh={18} />Jellyfin
            </button>
          {/if}
        </div>
      {:else}
        <div class="login-btns">
          <button type="submit">Sign Up</button>
        </div>
      {/if}
      {#if availableProviders.findIndex((provider) => provider == "plex") > -1}
        <p>or</p>
        <div class="login-btns">
          <button type="submit" name="plex" class="plex other">
            <Icon i="plex" wh={18} />Plex
          </button>
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
