<script lang="ts">
  export let name: string;
  export let value: string | number;
  export let large = false;
  export let href: string | undefined = undefined;
</script>

<a {href}>
  <span class={large ? "large" : ""}>{value}</span>
  <span>{name}</span>
</a>

<style lang="scss">
  a {
    display: flex;
    flex-flow: column;
    flex-grow: 1;
    flex: 1 1 115px;
    padding: 20px 15px;
    background-color: $accent-color;
    border-radius: 8px;

    > span:first-child {
      font-weight: bold;
      font-size: 20px;

      &.large {
        font-size: 32px;
      }
    }

    > span:last-child {
      margin-top: auto;
    }
  }
</style>
