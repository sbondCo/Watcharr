<script lang="ts">
  import type { ContentType, PublicUser, WatchedStatus } from "@/types";
  import HorizontalList from "../HorizontalList.svelte";
  import Modal from "../Modal.svelte";
  import axios from "axios";
  import Spinner from "../Spinner.svelte";
  import Error from "../Error.svelte";
  import Icon from "../Icon.svelte";
  import { watchedStatuses } from "../util/helpers";

  interface FollowThoughts {
    followedUser: PublicUser;
    thoughts: string;
    status: WatchedStatus;
    rating: number;
  }

  export let mediaType: ContentType | "game";
  // The tmdbId for movie/tv, igdbId for games.
  export let mediaId: number;

  let modalShownFor: FollowThoughts | undefined = undefined;

  async function getFollowsThoughts() {
    return (await axios.get<FollowThoughts[]>(`/follow/thoughts/${mediaType}/${mediaId}`)).data;
  }
</script>

{#await getFollowsThoughts()}
  <Spinner />
{:then fts}
  {#if fts?.length > 0}
    <HorizontalList title="Followed Thoughts">
      {#each fts as ft}
        <button
          class={["thoughts-card plain", ft.thoughts ? "" : "no-thoughts"].join(" ")}
          on:click={() => (modalShownFor = ft)}
        >
          <div>
            <h4 title={ft.followedUser.username}>{ft.followedUser.username}</h4>
            {#if ft.status}
              <div class="status-icon">
                <Icon i={watchedStatuses[ft.status]} wh={22} />
              </div>
            {/if}
            {#if ft.rating}
              <span class="rating">
                <span>*</span>
                {ft.rating}
              </span>
            {/if}
          </div>
          <div class="thought">
            {ft.thoughts || "No thoughts yet."}
          </div>
        </button>
      {/each}
    </HorizontalList>
  {/if}
{:catch err}
  <Error error={err} pretty="Failed to load followed thoughts!" />
{/await}

{#if modalShownFor}
  <Modal
    title={`${modalShownFor.followedUser.username}'s Thoughts`}
    onClose={() => (modalShownFor = undefined)}
  >
    <span>{modalShownFor.thoughts}</span>
  </Modal>
{/if}

<style lang="scss">
  .thoughts-card {
    display: flex;
    flex-flow: column;
    padding: 15px 20px;
    background-color: $accent-color;
    fill: $text-color;
    border-radius: 10px;
    min-width: 250px;
    max-width: 250px;
    font-size: 16px;
    transition: background-color 200ms ease;
    text-align: unset;
    overflow: hidden;
    cursor: pointer;

    & > div {
      display: flex;
      flex-flow: row;
      align-items: center;
      gap: 8px;
      margin-bottom: 3px;

      h4 {
        overflow: hidden;
        text-overflow: ellipsis;
      }

      .rating {
        display: flex;
        align-items: start;
        justify-content: center;
        gap: 5px;
        font-size: 18px;
        font-weight: bolder;

        span {
          font-family: "Rampart One";
          -webkit-text-stroke: 1px $text-color;
          font-size: 32px;
          line-height: 0.55;
          margin-top: 7px;
        }
      }

      .status-icon {
        margin-left: auto;
        margin-right: 2px;
        width: 20px;
        height: 20px;
        min-width: 20px;
        min-height: 20px;
      }
    }

    .thought {
      display: -webkit-box;
      -webkit-line-clamp: 9;
      -webkit-box-orient: vertical;
      hyphens: auto;
      overflow: hidden;
    }

    &.no-thoughts {
      pointer-events: none;

      .thought {
        font-weight: lighter;
      }
    }

    &:hover {
      color: $bg-color;
      fill: $bg-color;
      background-color: $accent-color-hover;

      .rating span {
        -webkit-text-stroke: 1px $bg-color;
      }
    }
  }
</style>
