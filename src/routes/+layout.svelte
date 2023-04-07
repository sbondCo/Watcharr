<script lang="ts">
  import { goto } from "$app/navigation";

  let searchTimeout: number;

  function handleProfileClick() {
    if (!localStorage.getItem("token")) {
      goto("/login");
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
</script>

<nav>
  <a href="/"><h1>Watcharr</h1></a>
  <input type="text" placeholder="Search" on:keydown={handleSearch} />
  <button class="plain" on:click={handleProfileClick}>:)</button>
</nav>

<slot />

<style lang="scss">
  @import url("https://fonts.googleapis.com/css2?family=Rampart+One&display=swap");

  :global(*) {
    padding: 0;
    margin: 0;
    box-sizing: border-box;
  }

  :global(body) {
    font-family: sans-serif, system-ui, -apple-system, BlinkMacSystemFont;
  }

  :global(h1, h2, h3, h4, h5) {
    font-family: "Rampart One", system-ui, -apple-system, BlinkMacSystemFont;
  }

  :global(input) {
    padding: 5px 10px;
    border: 2px solid black;
    border-radius: 5px;
    width: 100%;
  }

  :global(button) {
    cursor: pointer;
  }

  :global(button:not(.plain)) {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 5px 10px;
    border: 2px solid black;
    border-radius: 5px;
    background-color: white;
    color: black;
    font-weight: bold;
    width: 100%;
    transition: background-color 100ms ease, opacity 100ms ease;

    &:hover,
    &:focus-visible {
      background-color: black;
      color: white;
      opacity: 1;
    }
  }

  :global(button.secondary) {
    border: 2px solid transparent;

    &:hover,
    &:focus-visible {
      background-color: white;
      color: black;
      border: 2px solid black;
    }
  }

  :global(button.plain) {
    background-color: transparent;
    color: black;
    border: 0;
  }

  :global(button.not-active) {
    opacity: 0.5;
  }

  nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 10px 20px 28px 20px;
    position: relative;

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

    button {
      font-family: "Rampart One", system-ui, -apple-system, BlinkMacSystemFont;
      font-size: 25px;
      writing-mode: vertical-rl;
      text-orientation: mixed;
      cursor: pointer;
    }
  }
</style>
