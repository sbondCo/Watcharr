<script lang="ts">
  export let title: string = "";
  export let desc: string = "";
  export let row: boolean = false;
  export let tag: string | undefined = undefined;
</script>

<div class={[row ? "row" : "", "setting-ctr"].join(" ")}>
  {#if row}
    <div>
      <h4 class="norm">
        {title}
        {#if tag}
          <span class="tag">{tag}</span>
        {/if}
      </h4>
      <h5 class="norm">{desc}</h5>
    </div>
  {:else}
    <h4 class="norm">
      {title}
      {#if tag}
        <span class="tag">{tag}</span>
      {/if}
    </h4>
    <h5 class="norm">{desc}</h5>
  {/if}

  <slot />
</div>

<style lang="scss">
  h4 {
    font-size: 16px;
    display: flex;
    flex-flow: row;
    align-items: center;
    gap: 5px;

    & > .tag {
      font-size: 10px;
      background-color: $accent-color;
      padding: 3px 5px;
      border-radius: 5px;
    }
  }

  h5 {
    font-weight: normal;
    font-size: 13.28px;
  }

  .setting-ctr {
    margin: 0 15px;
  }

  div {
    &.row {
      display: flex;
      flex-flow: row;
      gap: 10px;
      align-items: center;

      & > div:first-of-type {
        margin-right: auto;
      }
    }
  }
</style>
