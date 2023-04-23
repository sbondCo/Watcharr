<div>
  <ul>
    <slot />
  </ul>
</div>

<style lang="scss">
  div {
    display: flex;
    justify-content: center;
  }

  ul {
    display: flex;
    flex-flow: row;
    justify-content: center;
    gap: 10px;
    list-style: none;
    flex-wrap: wrap;
    margin: 20px 10px;
    max-width: 1200px;
  }
</style>
