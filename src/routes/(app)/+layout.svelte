<script lang="ts">
  import { goto } from "$app/navigation";
  import { clearAllStores } from "@/store";

  let searchTimeout: number;
  let subMenuShown = false;

  function handleProfileClick() {
    if (!localStorage.getItem("token")) {
      goto("/login");
    } else {
      subMenuShown = !subMenuShown;
    }
  }

  function handleSearch(ev: Event) {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      const target = ev.target as HTMLInputElement;
      const query = target?.value;
      if (query) {
        goto(`/search/${query}`).then(() => {
          target?.focus();
        });
      }
    }, 400);
  }

  function logout() {
    localStorage.removeItem("token");
    clearAllStores();
    goto("/login");
  }
</script>

<nav>
  <a href="/"><h1>Watcharr</h1></a>
  <input type="text" placeholder="Search" on:keydown={handleSearch} />
  <button class="plain face" on:click={handleProfileClick}>:)</button>
  {#if subMenuShown}
    <div>
      <button class="plain" style="text-decoration: line-through;">Profile</button>
      <button class="plain" on:click={() => logout()}>Logout</button>
    </div>
  {/if}
</nav>

<slot />

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
    }

    h1 {
      color: white;
      -webkit-text-stroke: 1.5px black;
      font-size: 35px;
    }

    input {
      width: 250px;
      font-weight: bold;
      text-align: center;
      box-shadow: 4px 4px 0px 0px rgba(0, 0, 0, 1);
    }

    button.face {
      font-family: "Rampart One", system-ui, -apple-system, BlinkMacSystemFont;
      font-size: 25px;
      writing-mode: vertical-rl;
      text-orientation: mixed;
      cursor: pointer;
    }

    div {
      display: flex;
      flex-flow: column;
      position: absolute;
      right: 0;
      top: 55px;
      padding: 10px;
      border: 3px solid black;
      border-radius: 10px;
      background-color: white;
      list-style: none;
      z-index: 50;

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
    }

    @media screen and (max-width: 580px) {
      h1 {
        display: none;
      }

      input {
        width: 100%;
      }
    }
  }
</style>
