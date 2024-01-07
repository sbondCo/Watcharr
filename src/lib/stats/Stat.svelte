<script lang="ts">
  import tooltip from "../actions/tooltip";

  export let name: string;
  export let value: string | number;
  export let large = false;
  export let href: string | undefined = undefined;
  export let disc: string | undefined = undefined;
</script>

<a {href}>
  {#if disc}
    <div class="disclaimer" use:tooltip={{ text: disc, pos: "top" }}>*</div>
  {/if}
  <span class={large ? "large" : ""}>{value}</span>
  <span>{name}</span>
</a>

<style lang="scss">
  a {
    position: relative;
    display: flex;
    flex-flow: column;
    flex-grow: 1;
    flex: 1 1 115px;
    padding: 20px 15px;
    background-color: $accent-color;
    border-radius: 8px;

    > span:first-of-type {
      font-weight: bold;
      font-size: 20px;

      &.large {
        font-size: 32px;
      }
    }

    > span:last-child {
      margin-top: auto;
    }

    .disclaimer {
      position: absolute;
      top: 5px;
      right: 8px;
      font-weight: bold;
      font-size: 20px;
    }
  }
</style>
