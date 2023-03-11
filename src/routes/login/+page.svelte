<div>
  <div class="inner">
    <h2>Get Back In!</h2>
    <label for="username">Username</label>
    <input type="text" name="username" placeholder="Username" />

    <label for="password">Password</label>
    <input type="text" name="password" placeholder="Password" />

    <button>Login</button>
  </div>
</div>

<style lang="scss">
  div {
    display: flex;
    flex-flow: column;
    align-items: center;
    gap: 10px;

    .inner {
      width: 100%;
      max-width: 400px;
    }

    label {
      align-self: flex-start;
      font-weight: bold;
    }
  }
</style>
