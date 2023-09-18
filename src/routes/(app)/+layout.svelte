<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Icon from "@/lib/Icon.svelte";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { isTouch, parseTokenPayload } from "@/lib/util/helpers";
  import { notify } from "@/lib/util/notify";
  import { activeFilter, clearAllStores, searchQuery, userSettings, watchedList } from "@/store";
  import axios from "axios";
  import { onMount } from "svelte";
  import { get } from "svelte/store";

  const username = localStorage.getItem("username");

  let navEl: HTMLElement;
  let searchTimeout: number;
  let subMenuShown = false;
  let filterMenuShown = false;

  $: filter = $activeFilter;
  $: settings = $userSettings;

  function handleProfileClick() {
    if (!localStorage.getItem("token")) {
      goto("/login");
    } else {
      subMenuShown = !subMenuShown;
      filterMenuShown = false;
    }
  }

  function handleFilterClick() {
    filterMenuShown = !filterMenuShown;
    subMenuShown = false;
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
      const [w, u] = await Promise.all([axios.get("/watched"), axios.get("/user/settings")]);
      if (w?.data?.length > 0) {
        watchedList.update((wl) => (wl = w.data));
      }
      if (u?.data) {
        userSettings.update((us) => (us = u.data));
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

  onMount(() => {
    if (navEl) {
      let scroll = window.scrollY;
      window.document.addEventListener("scroll", (ev: Event) => {
        if (scroll > window.scrollY) {
          navEl.classList.remove("scrolled-down");
        } else {
          navEl.classList.add("scrolled-down");
          subMenuShown = false;
          filterMenuShown = false;
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
      <button class="plain other" on:click={handleFilterClick}>
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
            class={`plain ${filter[0] == "LASTCHANGED" ? filter[1].toLowerCase() : ""}`}
            on:click={() => filterClicked("LASTCHANGED")}
          >
            Last Changed
          </button>
          <button
            class={`plain ${filter[0] == "RATING" ? filter[1].toLowerCase() : ""}`}
            on:click={() => filterClicked("RATING")}
          >
            Rating
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
    <button class="plain other discover" on:click={() => goto("/discover")}>
      <Icon i="compass" />
    </button>
    <button class="plain face" on:click={handleProfileClick}>:)</button>
    {#if subMenuShown}
      <div class="face-menu">
        {#if username}
          <h5 title={username}>Hi {username}!</h5>
        {/if}
        <button class="plain" on:click={() => profile()}>Profile</button>
        {#if !settings.private}
          <button class="plain" on:click={() => shareWatchedList()}>Share List</button>
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
      gap: 20px;

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
          stroke: $bg-color;
          stroke-width: 10px;
        }
      }

      button.discover {
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
            font-family:
              system-ui,
              -apple-system,
              BlinkMacSystemFont;
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
