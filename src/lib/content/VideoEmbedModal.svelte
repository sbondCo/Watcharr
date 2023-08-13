<script lang="ts">
  export let embed: string;
  export let closed: () => void;
</script>

<div class="modal" on:click={closed} on:keydown={closed}>
  {#if embed}
    <div class="wrapper">
      <iframe title="Video Embed" src={embed} frameborder="0" width="100%" height="100%" />
    </div>
  {/if}
</div>

<style lang="scss">
  .modal {
    display: flex;
    align-items: center;
    justify-content: center;
    position: fixed;
    top: 0;
    left: 0;
    width: 100dvw;
    height: 100dvh;
    background-color: rgba($color: #000000, $alpha: 0.5);
    z-index: 9999999;

    .wrapper {
      width: 100%;
      max-width: 800px;
      border-radius: 15px;
      aspect-ratio: 16 / 9;
      overflow: hidden;
      margin: 20px;
    }
  }
</style>
