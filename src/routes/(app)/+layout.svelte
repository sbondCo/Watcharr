<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Icon from "@/lib/Icon.svelte";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { isTouch, parseTokenPayload, userHasPermission } from "@/lib/util/helpers";
  import { notify } from "@/lib/util/notify";
  import {
    activeFilters,
    activeSort,
    clearAllStores,
    searchQuery,
    userInfo,
    userSettings,
    watchedList
  } from "@/store";
  import { type Filters, UserPermission } from "@/types";
  import axios from "axios";
  import { onMount } from "svelte";
  import { get } from "svelte/store";

  let navEl: HTMLElement;
  let searchTimeout: number;
  let subMenuShown = false;
  let filterMenuShown = false;
  let sortMenuShown = false;

  $: sort = $activeSort;
  $: filter = $activeFilters;
  $: settings = $userSettings;
  $: user = $userInfo;

  function handleProfileClick() {
    if (!localStorage.getItem("token")) {
      goto("/login");
    } else {
      subMenuShown = !subMenuShown;
      filterMenuShown = false;
      sortMenuShown = false;
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

  function serverSettings() {
    goto("/server");
    subMenuShown = false;
  }

  function shareWatchedList() {
    const nid = notify({ type: "loading", text: "Getting link" });
    const ud = parseTokenPayload();
    console.log(ud);
    if (ud?.userId && ud?.username) {
      const shareLink = `${window.location.origin}/lists/${ud.userId}/${ud.username}`;
      navigator.clipboard
        .writeText(shareLink)
        .then(() => {
          notify({ id: nid, type: "success", text: "Copied share link" });
        })
        .catch((r) => {
          console.error("Failed to copy list share link", r);
          notify({
            id: nid,
            type: "error",
            text: `Failed to copy share link:<br/><a href="${shareLink}" target="_blank">${shareLink}</a>`,
            time: 20000
          });
        });
    } else {
      notify({ id: nid, type: "error", text: "Failed to get link" });
    }
  }

  async function getInitialData() {
    if (localStorage.getItem("token")) {
      const [w, u, s] = await Promise.all([
        axios.get("/watched"),
        axios.get("/user"),
        axios.get("/user/settings")
      ]);
      if (w?.data?.length > 0) {
        watchedList.update((wl) => (wl = w.data));
      }
      if (u?.data) {
        userInfo.update((ui) => (ui = u.data));
      }
      if (s?.data) {
        userSettings.update((us) => (us = s.data));
      }
    } else {
      goto("/login?again=1");
    }
  }

  function sortClicked(type: string, modeType: string = "UPDOWN") {
    const af = get(activeSort);
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
    activeSort.update((af) => (af = [type, mode]));
  }

  function filterClicked(type: keyof Filters, f: string) {
    const af = get(activeFilters);
    if (af[type]?.includes(f)) {
      af[type] = af[type]?.filter((a) => a !== f);
    } else {
      af[type]?.push(f);
    }
    activeFilters.update((a) => (a = af));
  }

  onMount(() => {
    if (navEl) {
      let scroll = window.scrollY;
      window.document.addEventListener("scroll", (ev: Event) => {
        if (scroll > window.scrollY) {
          navEl?.classList.remove("scrolled-down");
        } else {
          navEl?.classList.add("scrolled-down");
          subMenuShown = false;
          filterMenuShown = false;
          sortMenuShown = false;
        }
        scroll = window.scrollY;
      });
    } else {
      console.error("navEl doesn't exist, failed to initialize up/down listener");
    }
  });
</script>

<nav bind:this={navEl}>
  <a href="/">
    <span class="large">Watcharr</span>
    <span class="small">W</span>
  </a>
  <input type="text" placeholder="Search" bind:value={$searchQuery} on:keydown={handleSearch} />
  <div class="btns">
    <!-- Only show on watched list -->
    {#if $page.url?.pathname === "/"}
      <button
        class="plain other filter"
        on:click={() => {
          filterMenuShown = !filterMenuShown;
          sortMenuShown = false;
          subMenuShown = false;
        }}
      >
        <Icon i="filter" />
      </button>
      <button
        class="plain other sort"
        on:click={() => {
          sortMenuShown = !sortMenuShown;
          filterMenuShown = false;
          subMenuShown = false;
        }}
      >
        <Icon i="sort" />
      </button>
      {#if sortMenuShown}
        <div class="menu sort-menu">
          <button
            class={`plain ${sort[0] == "DATEADDED" ? sort[1].toLowerCase() : ""}`}
            on:click={() => sortClicked("DATEADDED")}
          >
            Date Added
          </button>
          <button
            class={`plain ${sort[0] == "LASTCHANGED" ? sort[1].toLowerCase() : ""}`}
            on:click={() => sortClicked("LASTCHANGED")}
          >
            Last Changed
          </button>
          <button
            class={`plain ${sort[0] == "RATING" ? sort[1].toLowerCase() : ""}`}
            on:click={() => sortClicked("RATING")}
          >
            Rating
          </button>
          <button
            class={`plain ${sort[0] == "ALPHA" ? sort[1].toLowerCase() : ""}`}
            on:click={() => sortClicked("ALPHA")}
          >
            Alphabetical
          </button>
        </div>
      {/if}
      {#if filterMenuShown}
        <div class="menu filter-menu">
          <h4 class="norm sm-caps">type</h4>
          <div class="type-filter">
            <button
              class={`${filter.type.includes("tv") ? "active" : ""}`}
              on:click={() => filterClicked("type", "tv")}
            >
              SHOW
            </button>
            <button
              class={`${filter.type.includes("movie") ? "active" : ""}`}
              on:click={() => filterClicked("type", "movie")}
            >
              MOVIE
            </button>
          </div>
          <h4 class="norm sm-caps">status</h4>
          <button
            class={`plain ${filter.status.includes("planned") ? "on" : ""}`}
            on:click={() => filterClicked("status", "planned")}
          >
            planned
          </button>
          <button
            class={`plain ${filter.status.includes("watching") ? "on" : ""}`}
            on:click={() => filterClicked("status", "watching")}
          >
            watching
          </button>
          <button
            class={`plain ${filter.status.includes("finished") ? "on" : ""}`}
            on:click={() => filterClicked("status", "finished")}
          >
            finished
          </button>
          <button
            class={`plain ${filter.status.includes("hold") ? "on" : ""}`}
            on:click={() => filterClicked("status", "hold")}
          >
            held
          </button>
          <button
            class={`plain ${filter.status.includes("dropped") ? "on" : ""}`}
            on:click={() => filterClicked("status", "dropped")}
          >
            dropped
          </button>
        </div>
      {/if}
    {/if}
    <button class="plain other discover" on:click={() => goto("/discover")}>
      <Icon i="compass" />
    </button>
    <button class="plain face" on:click={handleProfileClick}>:)</button>
    {#if subMenuShown}
      <div class="menu face-menu">
        {#if user?.username}
          <h5 title={user.username}>Hi {user.username}!</h5>
        {/if}
        <button class="plain" on:click={() => profile()}>Profile</button>
        {#if !settings?.private}
          <button class="plain" on:click={() => shareWatchedList()}>Share List</button>
        {/if}
        {#if user && userHasPermission(user.permissions, UserPermission.PERM_ADMIN)}
          <button class="plain" on:click={() => serverSettings()}>Settings</button>
        {/if}
        <button class="plain" on:click={() => logout()}>Logout</button>
        <!-- svelte-ignore missing-declaration -->
        <span>v{__WATCHARR_VERSION__}</span>
      </div>
    {/if}
  </div>
</nav>

{#await getInitialData()}
  <Spinner />
{:then}
  <slot />
{:catch err}
  <PageError pretty="Failed to retrieve user data!" error={err} />
{/await}

<style lang="scss">
  nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;
    padding: 10px 20px;
    position: sticky;
    top: 0;
    gap: 20px;
    z-index: 999999999;
    backdrop-filter: blur(2.5px) saturate(120%);
    background-color: $nav-color;
    transition: top 200ms ease-in-out;

    &:global(.scrolled-down) {
      top: -71px;
    }

    a {
      text-decoration: none;
      font-family:
        "Shrikhand",
        system-ui,
        -apple-system,
        BlinkMacSystemFont;
      font-size: 35px;
      transition:
        -webkit-text-stroke 150ms ease,
        color 150ms ease,
        font-weight 150ms ease;

      &:hover,
      &:focus-visible {
        color: $bg-color;
        -webkit-text-stroke: 3px $text-color;
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
      box-shadow: 4px 4px 0px 0px $text-color;
      transition:
        width 150ms ease,
        box-shadow 150ms ease;

      &:hover,
      &:focus {
        box-shadow: 2px 2px 0px 0px $text-color;
      }
    }

    .btns {
      display: flex;
      flex-flow: row;
      /* gap: 20px; */

      button.other {
        padding-top: 2px;
        width: 28px;
        transition:
          fill 150ms ease,
          stroke 150ms ease,
          stroke-width 150ms ease;
        fill: $text-color;

        &:hover,
        &:focus-visible {
          :global(path) {
            fill: none;
            stroke: $text-color;
            stroke-width: 30px;
            stroke-linejoin: round;
          }
        }
      }

      button.filter {
        margin-right: 15px;
        &:hover,
        &:focus-visible {
          :global(path) {
            stroke-width: 15px;
          }
        }
      }

      button.sort {
        margin-right: 12px;
      }

      button.discover {
        margin-right: 17px;
        transition:
          fill 150ms ease,
          stroke 150ms ease,
          stroke-width 150ms ease,
          transform 150ms ease;

        &:hover,
        &:focus-visible {
          transform: rotate(60deg);
        }
      }

      button.face {
        font-family:
          "Shrikhand",
          system-ui,
          -apple-system,
          BlinkMacSystemFont;
        font-size: 25px;
        transform: rotate(90deg);
        cursor: pointer;
        margin-left: 3px;
        transition:
          -webkit-text-stroke 150ms ease,
          color 150ms ease;

        &:hover,
        &:focus-visible {
          color: $bg-color;
          -webkit-text-stroke: 1.5px $text-color;
        }
      }

      div.menu {
        display: flex;
        flex-flow: column;
        position: absolute;
        right: 3px;
        top: 55px;
        width: 125px;
        padding: 10px;
        border: 3px solid $text-color;
        border-radius: 10px;
        background-color: $bg-color;
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

          &:hover,
          &:focus-visible {
            background-color: $text-color;
            color: $bg-color;
          }
        }

        span {
          margin-top: 8px;
          font-size: 11px;
          color: gray;
          text-align: center;
        }
      }

      div.sort-menu {
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
            font-family:
              system-ui,
              -apple-system,
              BlinkMacSystemFont;
            font-size: 18px;
          }
        }
      }

      div.filter-menu {
        width: 180px;
        right: 35px;

        h4 {
          margin-bottom: 8px;

          &:not(:first-of-type) {
            margin-top: 8px;
          }
        }

        & > button {
          text-transform: capitalize;
          position: relative;

          &.on::before {
            content: "\2713";
          }

          &::before {
            position: absolute;
            top: 4px;
            left: 12px;
            font-family:
              system-ui,
              -apple-system,
              BlinkMacSystemFont;
            font-size: 18px;
          }
        }

        .type-filter {
          display: flex;
          flex-flow: row;
          width: 100%;

          button {
            border-radius: 0;
            padding: 8px 0;
            width: 100%;

            &:first-of-type {
              border-right: 0;
              border-radius: 5px 0 0 5px;
            }

            &:last-of-type {
              border-radius: 0 5px 5px 0;
            }
          }
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
