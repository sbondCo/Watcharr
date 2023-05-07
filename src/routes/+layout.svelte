<div id="tooltip" />
<slot />

<style lang="scss">
  @font-face {
    font-family: "Shrikhand";
    src: local("Shrikhand Regular"), local("Shrikhand-Regular"),
      url("/Shrikhand-Regular.woff2") format("woff2");
    font-weight: normal;
    font-style: italic;
    font-display: swap;
  }

  @font-face {
    font-family: "Rampart One";
    src: local("Rampart One Regular"), local("RampartOne-Regular"),
      url("/RampartOne-Regular.woff2") format("woff2");
    font-weight: normal;
    font-style: normal;
    font-display: swap;
  }

  :global(*) {
    padding: 0;
    margin: 0;
    box-sizing: border-box;
  }

  :global(body) {
    font-family: sans-serif, system-ui, -apple-system, BlinkMacSystemFont;
  }

  :global(h1, h2, h3, h4, h5) {
    font-family: "Shrikhand", system-ui, -apple-system, BlinkMacSystemFont;
  }

  :global(a) {
    text-decoration: none;
    color: white;
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
      width: auto;
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

  :global(#tooltip) {
    visibility: hidden;
    position: absolute;
    padding: 8px 5px;
    background-color: black;
    color: white;
    border: 1px solid white;
    border-radius: 6px;
    z-index: 99999;
    transition: top 100ms ease, left 100ms ease;
    pointer-events: none;
  }
</style>
