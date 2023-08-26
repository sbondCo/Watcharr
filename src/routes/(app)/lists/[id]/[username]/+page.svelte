<script lang="ts">
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import WatchedList from "@/lib/WatchedList.svelte";
  import { notify } from "@/lib/util/notify.js";
  import type { Watched } from "@/types.js";
  import axios from "axios";

  export let data;

  async function getPublicWatchedList(id?: number, username?: string) {
    if (!id || !username) {
      console.error("getPublicWatchedList requires and id and username", id, username);
      notify({ type: "error", text: "Couldn't fetch list" });
    }
    return (await axios.get(`/watched/${id}/${username}`)).data as Watched[];
  }
</script>

<svelte:head>
  <title>{data.username}'s Watched List</title>
</svelte:head>

{#await getPublicWatchedList(Number(data.id), data.username)}
  <Spinner />
{:then watched}
  <WatchedList list={watched} isPublicList={true} />
{:catch err}
  <PageError pretty="Failed to get watched list!" error={err} />
{/await}
