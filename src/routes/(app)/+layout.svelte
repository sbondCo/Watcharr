<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Icon from "@/lib/Icon.svelte";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { isTouch } from "@/lib/util/helpers";
  import { activeFilter, clearAllStores, watchedList } from "@/store";
  import axios from "axios";
  import { get } from "svelte/store";

  const username = localStorage.getItem("username");

  let searchTimeout: number;
  let subMenuShown = false;
  let filterMenuShown = false;

  $: filter = $activeFilter;

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

  function profile() {
    goto("/profile");
    subMenuShown = false;
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

  function filterClicked(type: string, modeType: string = "UPDOWN") {
    const af = get(activeFilter);
    let mode: string;
    if (modeType === "UPDOWN") {
      mode = "UP";
      if (af[0] == type) {
        if (af[1] === "UP") {
          mode = "DOWN";
        } else if (af[1] === "DOWN") {
          mode = "";
        }
      }
    } else if (modeType === "TOGGLE") {
      mode = "ON";
      if (af[0] == type) {
        if (af[1] === "ON") {
          mode = "OFF";
        }
      }
    } else {
      console.error("filterClicked() ran without a valid modeType:", modeType);
      return;
    }
    activeFilter.update((af) => (af = [type, mode]));
  }
</script>

<nav>
  <a href="/">
    <span class="large">Watcharr</span>
    <span class="small">W</span>
  </a>
  <input type="text" placeholder="Search" on:keydown={handleSearch} />
  <div class="btns">
    <!-- Only show on watched list -->
    {#if $page.url?.pathname === "/"}
      <button class="plain filter" on:click={() => (filterMenuShown = !filterMenuShown)}>
        <Icon i="filter" />
      </button>
      {#if filterMenuShown}
        <div class="filter-menu">
          <button
            class={`plain ${filter[0] == "DATEADDED" ? filter[1].toLowerCase() : ""}`}
            on:click={() => filterClicked("DATEADDED")}
          >
            Date Added
          </button>
          <button
            class={`plain ${filter[0] == "ALPHA" ? filter[1].toLowerCase() : ""}`}
            on:click={() => filterClicked("ALPHA")}
          >
            Alphabetical
          </button>
        </div>
      {/if}
    {/if}
    <button class="plain face" on:click={handleProfileClick}>:)</button>
    {#if subMenuShown}
      <div class="face-menu">
        {#if username}
          <h5 title={username}>Hi {username}!</h5>
        {/if}
        <button class="plain" on:click={() => profile()}>Profile</button>
        <button class="plain" on:click={() => logout()}>Logout</button>
        <!-- svelte-ignore missing-declaration -->
        <span>v{__WATCHARR_VERSION__}</span>
      </div>
    {/if}
  </div>
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

    .btns {
      display: flex;
      flex-flow: row;
      gap: 30px;

      button.filter {
        padding-top: 2px;
        width: 28px;
        transition: fill 150ms ease, stroke 150ms ease, stroke-width 150ms ease;

        &:hover,
        &:focus-visible {
          fill: white;
          stroke: black;
          stroke-width: 10px;
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

      div.filter-menu {
        width: 180px;

        & > button {
          position: relative;

          &.down::before {
            content: "\2193";
          }

          &.up::before {
            content: "\2191";
          }

          &.on::before {
            content: "\2713";
          }

          &::before {
            position: absolute;
            top: 4px;
            left: 12px;
            font-family: system-ui, -apple-system, BlinkMacSystemFont;
            font-size: 18px;
          }
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
    }

    @media screen and (max-width: 580px) {
      input {
        width: 100%;
      }
    }
  }
</style>
