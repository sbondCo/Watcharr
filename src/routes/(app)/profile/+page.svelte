<div class="content">
  <div class="inner">
    <h2>Hey {localStorage.getItem("username")}</h2>

    <div class="stats" />
  </div>
</div>

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;

    .inner {
      min-width: 400px;
      max-width: 400px;
    }
  }
</style>
