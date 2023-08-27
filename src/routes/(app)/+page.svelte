<script lang="ts">
  import WatchedList from "@/lib/WatchedList.svelte";
  import { watchedList } from "@/store";
</script>

<svelte:head>
  <title>Watched List</title>
</svelte:head>

<WatchedList list={$watchedList} />
