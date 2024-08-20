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
      <strong>Viewing Tag:</strong>
      <div class="basic-ctr">
        <Tag {tag} />
      </div>
    </div>
  </div>

  {#if watcheds?.length > 0}
    <WatchedList list={watcheds} />
  {:else}
    <div class="content">
      <div class="inner">
        <strong>This tag is empty!</strong>
      </div>
    </div>
  {/if}
{:else}
  <div class="content">
    <div class="inner">
      <strong>Tag does not exist!</strong>
    </div>
  </div>
{/if}

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;

    .inner {
      display: flex;
      flex-flow: column;
      gap: 5px;
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
    display: flex;
    max-width: 300px;
    width: 100%;
  }
</style>
