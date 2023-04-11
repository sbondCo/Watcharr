<nav>
  <a href="/"><h1>Watcharr</h1></a>
</nav>

<slot />

<style lang="scss">
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
  }
</style>
