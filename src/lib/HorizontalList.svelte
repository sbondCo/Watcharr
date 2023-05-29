<script lang="ts">
  export let title: string | undefined;
</script>

{#if title}
  <h2>{title}</h2>
{/if}
<ul>
  <slot />
</ul>

<style lang="scss">
  h2 {
    font-family: sans-serif, system-ui, -apple-system, BlinkMacSystemFont;
    font-size: 30px;
    font-weight: bold;
    margin-left: 30px;
  }

  ul {
    display: flex;
    flex-flow: row;
    gap: 10px;
    list-style: none;
    padding: 15px 30px;
    justify-content: space-between;
  }
</style>
