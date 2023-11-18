<div>
  <slot />
</div>

<style lang="scss">
  div {
    display: flex;
    flex-flow: row;
    gap: 12px;
    margin-top: 15px;

    @media screen and (max-width: 440px) {
      flex-wrap: wrap;
    }
  }
</style>
