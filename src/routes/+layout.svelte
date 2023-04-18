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

    & :global(svg) {
      height: 100%;
      width: 100%;
    }

    &:hover,
    &:focus-visible {
      background-color: black;
      color: white;
      fill: white;
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
</style>
