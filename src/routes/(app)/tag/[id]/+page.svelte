<script lang="ts">
  import Icon from "@/lib/Icon.svelte";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import Tag from "@/lib/tag/Tag.svelte";
  import WatchedList from "@/lib/WatchedList.svelte";
  import { tags, watchedList } from "@/store.js";

  export let data;

  $: tag = $tags.find((t) => t.id === data.tagId);
  $: watcheds = $watchedList.filter((w) =>
    w.tags ? (w.tags.find((t) => t.id === data.tagId) ? true : false) : false
  );
</script>

<svelte:head>
  <title>{tag ? `${tag.name}` : "Tag"}</title>
</svelte:head>

{#if tag && watcheds}
  <div class="content">
    <div class="inner">
      <div class="basic-ctr">
        <Tag {tag} />
      </div>
    </div>
  </div>

  {#if watcheds?.length > 0}
    <WatchedList list={watcheds} />
  {:else}
    <span>This tag is empty!</span>
  {/if}
{/if}

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;

    .inner {
      display: flex;
      flex-flow: row;
      gap: 15px;
      justify-content: center;
      align-items: center;
      width: 100%;
      max-width: 1200px;
      margin: 20px 30px;
      margin-top: 0;
    }
  }

  button {
    width: max-content;
  }

  textarea {
    border: 0;
    padding: 0;
    resize: none;
    text-overflow: ellipsis;
  }

  .basic-ctr {
    max-width: 300px;
    overflow: hidden;
  }
</style>
