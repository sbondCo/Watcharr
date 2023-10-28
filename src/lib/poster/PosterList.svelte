<script lang="ts">
  import { onMount } from "svelte";

  export let type: "wrapped" | "vertical" = "wrapped";

  let ulEl: HTMLUListElement;

  onMount(() => {
    if (ulEl) {
      ulEl.classList.add(type);
    }
  });
</script>

<div>
  <ul bind:this={ulEl}>
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

    &:global(.vertical) {
      flex-wrap: nowrap;
      justify-content: unset;
      overflow-x: auto;
      padding: 15px 8px;
      margin: 5px 0;
    }
  }
</style>
