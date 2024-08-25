<!-- Uses Native Color Input -->

<script lang="ts">
  export let value = "#000000";
  export let style = "";
</script>

<button class="plain" style={`background-color: ${value};${style}`}>
  <input type="color" bind:value />
</button>

<style lang="scss">
  button {
    width: 100%;
    height: 100%;
    border-radius: 10px;
    outline: 1px solid $accent-color;
    cursor: pointer;
  }

  input {
    opacity: 0;
    cursor: pointer;
  }
</style>
