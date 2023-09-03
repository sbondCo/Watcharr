<script lang="ts">
  export let title: string;
  export let desc: string;
</script>

<div class="backdrop"></div>
<div class="modal">
  <div>
    <h3 class="norm">{title}</h3>
    <h5 class="norm">{desc}</h5>
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
    z-index: 9998;
    backdrop-filter: blur(2px) saturate(180%);
    background-color: color-mix(in srgb, $text-color 85%, transparent);
  }

  .modal {
    display: flex;
    justify-content: center;
    align-items: center;
    width: 100dvw;
    height: 100dvh;
    top: 0;
    left: 0;
    position: absolute;
    z-index: 9999;

    & > div {
      min-width: 300px;
      min-height: 300px;
      width: 100%;
      max-width: 1000px;
      background-color: $bg-color;
      border-radius: 10px;
      padding: 15px 20px;
      margin: 20px 100px;
      transition: margin 100ms ease;

      @media screen and (max-width: 680px) {
        margin: 20px;
      }
    }
  }
</style>
