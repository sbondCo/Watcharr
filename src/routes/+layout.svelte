<nav><h1>Watcharr</h1></nav>

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
    padding: 5px 10px;
    border: 2px solid black;
    border-radius: 5px;
    background-color: white;
    color: black;
    font-weight: bold;
    width: 100%;

    &:hover,
    &:focus-visible {
      background-color: black;
      color: white;
    }
  }

  :global(button.plain) {
    background-color: transparent;
    color: black;
    border: 0;
  }

  nav {
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 12px;

    h1 {
      color: white;
      -webkit-text-stroke: 1.5px black;
      font-size: 35px;
    }
  }
</style>
