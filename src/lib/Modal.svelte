<script lang="ts">
  import Icon from "./Icon.svelte";

  export let title: string;
  export let desc: string | undefined = undefined;
  export let onClose: (() => void) | undefined = undefined;
</script>

<div class="backdrop"></div>
<div class="modal">
  <div>
    {#if typeof onClose !== "undefined"}
      <button class="close" on:click={onClose}><Icon i="close" wh="20" /></button>
    {/if}
    <h3 class="norm">{title}</h3>
    {#if desc}
      <h5 class="norm">{desc}</h5>
    {/if}
    <slot />
  </div>
</div>

<style lang="scss">
  .backdrop {
    position: absolute;
    top: 0;
    left: 0;
    width: 100dvw;
    height: 100dvh;
    backdrop-filter: blur(2px) saturate(180%);
    background-color: color-mix(in srgb, black 85%, transparent);
    position: fixed;
    z-index: 99998;
  }

  .modal {
    display: flex;
    justify-content: center;
    align-items: center;
    width: 100dvw;
    height: 100dvh;
    top: 0;
    left: 0;
    position: fixed;
    z-index: 99998;

    & > div {
      position: relative;
      display: flex;
      flex-flow: column;
      min-width: 300px;
      width: 100%;
      max-width: 1000px;
      max-height: calc(100% - 20px);
      background-color: $bg-color;
      border-radius: 10px;
      padding: 15px 20px;
      margin: 20px 100px;
      transition: margin 100ms ease;
      outline: 2px solid $text-color;
      overflow: auto;

      h5 {
        margin-bottom: 15px;
      }

      button.close {
        position: absolute;
        top: 8px;
        right: 8px;
        width: max-content;
        padding: 3px 5px;
      }

      @media screen and (max-width: 680px) {
        margin: 20px;
      }
    }
  }
</style>
