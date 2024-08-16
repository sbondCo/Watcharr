<script lang="ts">
  import tooltip from "../actions/tooltip";
  import Icon from "../Icon.svelte";
  import TagMenu from "./TagMenu.svelte";

  let menuOpen = false;
</script>

<div>
  <button
    on:click={() => (menuOpen = !menuOpen)}
    use:tooltip={{
      text: `Add to a Tag`,
      pos: "bot"
    }}
  >
    <Icon i={"tag"} wh={19} />
  </button>

  {#if menuOpen}
    <TagMenu titleText="Add To Tag" classes="from-add-to-tag-btn" />
  {/if}
</div>

<style lang="scss">
  div {
    position: relative;
  }
</style>
