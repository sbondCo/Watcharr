<div>
  <slot />
</div>

<style lang="scss">
  div {
    display: flex;
    flex-flow: row;
    gap: 12px;
    margin-top: 15px;

    flex-wrap: wrap;
    @media screen and (max-width: 440px) {
    }
  }
</style>
