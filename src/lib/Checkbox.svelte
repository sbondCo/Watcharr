<div class="toggle-pill-color">
  <input type="checkbox" id="checkbox" />
  <label for="checkbox"></label>
</div>

<style lang="scss">
  .toggle-pill-color {
    width: min-content;

    input {
      display: none;
    }

    input:checked + label {
      background: $success;

      &::before {
        left: 1.6em;
      }
    }

    label {
      display: block;
      position: relative;
      width: 3em;
      height: 1.6em;
      border-radius: 1em;
      background: $error;
      cursor: pointer;
      user-select: none;
      transition: background 0.1s ease-in-out;

      &::before {
        content: "";
        display: block;
        width: 1.2em;
        height: 1.2em;
        border-radius: 1em;
        background: #fff;
        position: absolute;
        left: 0.2em;
        top: 0.2em;
        transition: all 0.2s ease-in-out;
      }
    }
  }
</style>
