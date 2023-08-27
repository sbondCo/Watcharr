<script lang="ts">
  export let value: boolean = false;
  export let toggled: (on: boolean) => void;
  export let disabled = false;

  let actualDisabled = false;

  $: {
    // In cases where we disable and enable the checkbox super fast
    // it ends up looking like a jittery mess. To circumvent this
    // issue, we add a small delay before undisabling.
    // (eg. fast internet and net request dis and re enables it quickly)
    if (disabled) {
      actualDisabled = true;
    } else {
      setTimeout(() => {
        actualDisabled = disabled;
      }, 150);
    }
  }

  function checkboxChange(e: Event) {
    toggled((e.target as HTMLInputElement).checked);
  }
</script>

<div class="toggle-pill-color">
  <input
    bind:checked={value}
    disabled={actualDisabled}
    type="checkbox"
    id="checkbox"
    on:change={checkboxChange}
  />
  <label for="checkbox"></label>
</div>

<style lang="scss">
  .toggle-pill-color {
    width: min-content;

    input {
      display: none;

      &:checked + label {
        background: $success;

        &::before {
          left: 1.6em;
        }
      }

      &:disabled + label {
        opacity: 0.5;
        cursor: not-allowed;
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
      transition:
        opacity 100ms ease-in-out,
        opacity 100ms ease-in-out;

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
