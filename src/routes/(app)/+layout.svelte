<script lang="ts">
  import { goto } from "$app/navigation";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { isTouch } from "@/lib/util/helpers";
  import { clearAllStores, watchedList } from "@/store";
  import axios from "axios";

  const username = localStorage.getItem("username");

  let searchTimeout: number;
  let subMenuShown = false;

  function handleProfileClick() {
    if (!localStorage.getItem("token")) {
      goto("/login");
    } else {
      subMenuShown = !subMenuShown;
    }
  }

  function handleSearch(ev: KeyboardEvent) {
    if (
      ev.key === "Tab" ||
      ev.key === "CapsLock" ||
      ev.key === "OS" ||
      ev.key === "ArrowLeft" ||
      ev.key === "ArrowRight" ||
      ev.key === "ArrowUp" ||
      ev.key === "ArrowDown"
    )
      return;
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(
      () => {
        const target = ev.target as HTMLInputElement;
        const query = target?.value.trim();
        if (!query) return;
        if (query) {
          goto(`/search/${query}`).then(() => {
            target?.focus();
          });
        }
      },
      isTouch() ? 800 : 400
    );
  }

  function logout() {
    localStorage.removeItem("token");
    clearAllStores();
    goto("/login");
  }

  async function getWatchedList() {
    if (localStorage.getItem("token")) {
      const w = await axios.get("/watched");
      if (w?.data?.length > 0) {
        watchedList.update((wl) => (wl = w.data));
      }
    } else {
      goto("/login?again=1");
    }
  }
</script>

<nav>
  <a href="/">
    <span class="large">Watcharr</span>
    <span class="small">W</span>
  </a>
  <input type="text" placeholder="Search" on:keydown={handleSearch} />
  <button class="plain face" on:click={handleProfileClick}>:)</button>
  {#if subMenuShown}
    <div>
      {#if username}
        <h5 title={username}>Hi {username}!</h5>
      {/if}
      <button class="plain" style="text-decoration: line-through;">Profile</button>
      <button class="plain" on:click={() => logout()}>Logout</button>
      <!-- svelte-ignore missing-declaration -->
      <span>v{__WATCHARR_VERSION__}</span>
    </div>
  {/if}
</nav>

{#await getWatchedList()}
  <Spinner />
{:then}
  <slot />
{:catch err}
  <PageError pretty="Failed to load watched list!" error={err} />
{/await}

<style lang="scss">
  nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 10px 20px 28px 20px;
    position: relative;
    gap: 20px;

    a {
      text-decoration: none;
      color: black;
      font-family: "Shrikhand", system-ui, -apple-system, BlinkMacSystemFont;
      font-size: 35px;
      transition: -webkit-text-stroke 150ms ease, color 150ms ease, font-weight 150ms ease;

      &:hover,
      &:focus-visible {
        color: white;
        -webkit-text-stroke: 3px black;
        font-weight: bold;
      }

      span.large {
        display: block;
        width: 185.2px;
      }

      span.small {
        display: none;
        width: 40px;
      }

      @media screen and (max-width: 580px) {
        span.large {
          display: none;
        }
        span.small {
          display: block;
        }
      }
    }

    input {
      width: 250px;
      font-weight: bold;
      text-align: center;
      box-shadow: 4px 4px 0px 0px rgba(0, 0, 0, 1);
      transition: width 150ms ease, box-shadow 150ms ease;

      &:hover,
      &:focus {
        box-shadow: 2px 2px 0px 0px rgba(0, 0, 0, 1);
      }
    }

    button.face {
      font-family: "Shrikhand", system-ui, -apple-system, BlinkMacSystemFont;
      font-size: 25px;
      transform: rotate(90deg);
      cursor: pointer;
      transition: -webkit-text-stroke 150ms ease, color 150ms ease;

      &:hover,
      &:focus-visible {
        color: white;
        -webkit-text-stroke: 1.5px black;
      }
    }

    div {
      display: flex;
      flex-flow: column;
      position: absolute;
      right: 0;
      top: 55px;
      width: 125px;
      padding: 10px;
      border: 3px solid black;
      border-radius: 10px;
      background-color: white;
      list-style: none;
      z-index: 50;

      h5 {
        margin-bottom: 2px;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        cursor: default;
      }

      button {
        font-size: 14px;
        padding: 8px 16px;
        cursor: pointer;
        transition: background-color 200ms ease;

        &:hover {
          background-color: rgba(0, 0, 0, 1);
          color: white;
        }
      }

      span {
        margin-top: 8px;
        font-size: 11px;
        color: gray;
        text-align: center;
      }
    }

    @media screen and (max-width: 580px) {
      input {
        width: 100%;
      }
    }
  }
</style>
