<script lang="ts">
  export let title: string;
  export let desc: string = "";
  export let row: boolean = false;
</script>

<div class={[row ? "row" : "", "setting-ctr"].join(" ")}>
  {#if row}
    <div>
      <h4 class="norm">{title}</h4>
      <h5 class="norm">{desc}</h5>
    </div>
  {:else}
    <h4 class="norm">{title}</h4>
    <h5 class="norm">{desc}</h5>
  {/if}

  <slot />
</div>

<style lang="scss">
  h3 {
    font-variant: small-caps;
  }

  h4 {
    font-size: 16px;
  }

  h5 {
    font-weight: normal;
    font-size: 13.28px;
  }

  .setting-ctr {
    margin: 0 15px;
  }

  div {
    input[type="text"],
    input[type="password"] {
      margin-top: 1px;
    }

    &.row {
      display: flex;
      flex-flow: row;
      gap: 10px;
      align-items: center;

      & > div:first-of-type {
        margin-right: auto;
      }

      &.btns button {
        width: min-content;
      }
    }
  }
</style>
