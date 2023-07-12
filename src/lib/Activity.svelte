<script lang="ts">
  import type { Activity } from "@/types";

  export let activity: Activity[] | undefined;

  function getMsg(a: Activity) {
    switch (a?.type) {
      case "ADDED_WATCHED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `added to watched list${data?.status ? ` as ${data.status?.toLowerCase()}` : ""}${
            data?.rating ? ` with ${data.rating} stars` : ""
          }`;
        }
        return "added to watched list";
      case "RATING_CHANGED":
        if (a.data) {
          return `rating changed to ${a.data}`;
        }
        return "rating changed";
      case "STATUS_CHANGED":
        if (a.data) {
          return `watched status changed to ${a.data?.toLowerCase()}`;
        }
        return "watched status changed";
      default:
        return a.type;
    }
  }
</script>

<div class="activity">
  <h2>Activity</h2>
  {#if activity && activity.length > 0}
    <ul>
      {#each activity?.sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt)) as a}
        <li>
          {new Date(a.createdAt).toLocaleString()}
          {getMsg(a)}
        </li>
      {/each}
    </ul>
  {:else}
    <span>You Have No Activity!</span>
  {/if}
</div>

<style lang="scss">
  .activity {
    width: 100%;

    ul {
      display: flex;
      flex-flow: column;
      gap: 8px;
      margin-top: 8px;
      margin-left: calc(30px + 8px);
      list-style: none;
      max-height: 250px;
      overflow: auto;

      li {
        width: max-content;
        padding: 10px 12px;
        color: white;
        background-color: rgb(46, 46, 46);
        border-radius: 8px;
        text-transform: capitalize;
      }
    }

    span {
      margin-left: calc(30px);
      width: 100%;
      display: flex;
      justify-content: center;
    }
  }

  h2 {
    font-family: sans-serif, system-ui, -apple-system, BlinkMacSystemFont;
    font-size: 30px;
    font-weight: bold;
    margin-left: 30px;
  }
</style>
