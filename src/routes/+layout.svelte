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
    font-family: sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue';
  }

  :global(h1, h2, h3, h4, h5) {
    font-family: "Rampart One", system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
      Oxygen, Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
  }

  nav {
    display: flex;
    align-items: center;
    justify-content: center;

    h1 {
      color: white;
      -webkit-text-stroke: 1.5px black;
      font-size: 35px;
    }
  }
</style>
