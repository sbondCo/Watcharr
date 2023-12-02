<script lang="ts">
  import Icon from "@/lib/Icon.svelte";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import WatchedList from "@/lib/WatchedList.svelte";
  import tooltip from "@/lib/actions/tooltip.js";
  import UserAvatar from "@/lib/img/UserAvatar.svelte";
  import { followUser, unfollowUser } from "@/lib/util/api.js";
  import { notify } from "@/lib/util/notify.js";
  import { follows } from "@/store.js";
  import type { PublicUser, Watched } from "@/types.js";
  import axios from "axios";
  import { onMount } from "svelte";

  export let data;

  let followBtnDisabled = false;
  let user: PublicUser;

  $: isFollowing = !!$follows?.find((f) => f.followedUser.id === Number(data.id));

  async function getPublicWatchedList(id?: number, username?: string) {
    if (!id || !username) {
      console.error("getPublicWatchedList requires and id and username", id, username);
      notify({ type: "error", text: "Couldn't fetch list" });
    }
    return (await axios.get(`/watched/${id}/${username}`)).data as Watched[];
  }

  async function getPublicUser() {
    return (await axios.get(`/user/public/${data.id}/${data.username}`)).data as PublicUser;
  }

  async function follow() {
    followBtnDisabled = true;
    console.log(isFollowing);
    if (isFollowing) {
      await unfollowUser(Number(data.id));
    } else {
      await followUser(Number(data.id));
    }
    followBtnDisabled = false;
  }

  onMount(async () => {
    user = await getPublicUser();
  });
</script>

<svelte:head>
  <title>{data.username}'s Watched List</title>
</svelte:head>

<div class="content">
  <div class="inner">
    <UserAvatar img={user?.avatar} />
    <div class="basic-ctr">
      <div class="name-row">
        <h2>
          {data.username}
        </h2>
        <button
          class="plain"
          disabled={followBtnDisabled}
          on:click={follow}
          use:tooltip={{ text: isFollowing ? "Unfollow" : "Follow" }}
        >
          <Icon i={isFollowing ? "person-minus" : "person-add"} />
        </button>
      </div>
      <span>{user?.bio}</span>
    </div>
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
      flex-flow: row;
      gap: 15px;
      align-items: center;
      width: 100%;
      max-width: 1200px;
      margin: 20px 30px;
      margin-top: 0;
    }
  }

  .name-row {
    display: flex;
    flex-flow: row;

    button {
      margin-left: auto;
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
    width: 100%;
    max-width: 300px;
    overflow: hidden;

    span {
      font-family: monospace;
      overflow: hidden;
      text-overflow: ellipsis;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
    }
  }
</style>
