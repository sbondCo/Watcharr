<script lang="ts">
  import Icon from "../Icon.svelte";

  export let title: string;
  export let desc: string;
  export let onClick: () => void;
</script>

<button class="plain" on:click={onClick}>
  <div>
    <h4 class="norm">{title}</h4>
    <h5 class="norm">{desc}</h5>
  </div>
  <Icon i="arrow" facing="right" />
</button>

<style lang="scss">
  h4 {
    font-size: 16px;
  }

  h5 {
    font-weight: normal;
    font-size: 13.28px;
  }

  button {
    display: flex;
    flex-flow: row;
    align-items: center;
    margin: 0 15px;
    text-align: left;
    border: 2px dashed $text-color;
    border-radius: 10px;
    padding: 15px;
    transition: all 200ms ease;

    & > div {
      margin-right: auto;
    }

    &:hover {
      border-style: solid;
    }
  }
</style>
