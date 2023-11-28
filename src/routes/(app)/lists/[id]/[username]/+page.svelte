<script lang="ts">
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import WatchedList from "@/lib/WatchedList.svelte";
  import { followUser, unfollowUser } from "@/lib/util/api.js";
  import { notify } from "@/lib/util/notify.js";
  import { follows } from "@/store.js";
  import type { Watched } from "@/types.js";
  import axios from "axios";

  export let data;

  let followBtnDisabled = false;

  $: isFollowing = !!$follows?.find((f) => f.followedUser.id === Number(data.id));

  async function getPublicWatchedList(id?: number, username?: string) {
    if (!id || !username) {
      console.error("getPublicWatchedList requires and id and username", id, username);
      notify({ type: "error", text: "Couldn't fetch list" });
    }
    return (await axios.get(`/watched/${id}/${username}`)).data as Watched[];
  }

  async function follow() {
    followBtnDisabled = true;
    if (isFollowing) {
      await unfollowUser(Number(data.id));
    } else {
      await followUser(Number(data.id));
    }
    followBtnDisabled = false;
  }
</script>

<svelte:head>
  <title>{data.username}'s Watched List</title>
</svelte:head>

<div class="content">
  <div class="inner">
    <h2 class="norm">{data.username}</h2>

    <button disabled={followBtnDisabled} on:click={follow}>
      {isFollowing ? "Unfollow" : "Follow"}
    </button>
  </div>
</div>

{#await getPublicWatchedList(Number(data.id), data.username)}
  <Spinner />
{:then watched}
  <WatchedList list={watched} isPublicList={true} />
{:catch err}
  <PageError pretty="Failed to get watched list!" error={err} />
{/await}

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;

    .inner {
      display: flex;
      flex-flow: column;
      gap: 5px;
      width: 100%;
      max-width: 1200px;
      margin: 20px 30px;
      margin-top: 0;
    }
  }

  button {
    width: max-content;
  }
</style>
