<script lang="ts">
  import Icon from "@/lib/Icon.svelte";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import CreateTagModal from "@/lib/tag/CreateTagModal.svelte";
  import Tag from "@/lib/tag/Tag.svelte";
  import WatchedList from "@/lib/WatchedList.svelte";
  import { tags, watchedList } from "@/store.js";

  export let data;

  let tagEditModalShown = false;

  $: tag = $tags.find((t) => t.id === data.tagId);
  $: watcheds = $watchedList.filter((w) =>
    w.tags ? (w.tags.find((t) => t.id === data.tagId) ? true : false) : false
  );
</script>

<svelte:head>
  <title>{tag ? `${tag.name} - Tag` : "Tag"}</title>
</svelte:head>

{#if tag && watcheds}
  <div class="content">
    <div class="inner">
      <div class="basic-ctr">
        <Icon i="tag" wh={20} />
        <Tag
          {tag}
          onClick={() => {
            tagEditModalShown = !tagEditModalShown;
          }}
        />
      </div>
    </div>
  </div>

  {#if watcheds?.length > 0}
    <WatchedList list={watcheds} />
  {:else}
    <div class="content">
      <div class="inner">
        <strong>This tag is empty!</strong>
        <span>Add entries to this tag via it's page.</span>
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

{#if tagEditModalShown}
  <CreateTagModal existingTag={tag} onClose={() => (tagEditModalShown = false)} />
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

  .basic-ctr {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-wrap: wrap;
    gap: 10px;
    max-width: 300px;
    width: 100%;
  }
</style>
