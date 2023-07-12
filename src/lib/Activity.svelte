<script lang="ts">
  import type { Activity } from "@/types";
  import { getOrdinalSuffix, months } from "./util/helpers";

  export let activity: Activity[] | undefined;

  let groupedActivities: { [index: string]: any };

  $: {
    groupedActivities = getGroupedActivity(activity);
  }

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
      case "REMOVED_WATCHED":
        return "removed from watched list";
      case "RATING_CHANGED":
        if (a.data) {
          return `rating changed to ${a.data}`;
        }
        return "rating changed";
      case "STATUS_CHANGED":
        if (a.data) {
          return `status changed to ${a.data?.toLowerCase()}`;
        }
        return "status changed";
      default:
        return a.type;
    }
  }

  function toDayTime(d: Date) {
    return `${d.getDate()}${getOrdinalSuffix(d.getDate())} at ${d.toLocaleTimeString()}`;
  }

  function getGroupedActivity(activities?: Activity[]) {
    activities = activities?.filter((a) => a.type);
    const a = activities?.sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt));
    let grouped: { [index: string]: any } = {};
    if (a) {
      for (let i = 0; i < a.length; i++) {
        const activity = a[i];
        const date = new Date(Date.parse(activity.createdAt));
        const key = `${months[date.getMonth()]} ${date.getFullYear()}`;
        if (grouped[key]) {
          grouped[key].push(activity);
        } else {
          grouped[key] = [activity];
        }
      }
    }
    return grouped;
  }
</script>

<div class="activity">
  <h2>Activity</h2>
  {#if groupedActivities && Object.keys(groupedActivities).length > 0}
    <ul>
      {#each Object.keys(groupedActivities) as k}
        <h3>{k}</h3>

        {#each groupedActivities[k] as a}
          <li>
            <span title={new Date(a.createdAt).toDateString()}>
              {toDayTime(new Date(a.createdAt))}
            </span>
            <span>{getMsg(a)}</span>
          </li>
        {/each}
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
      overflow-y: auto;
      overflow-x: hidden;

      h3 {
        position: sticky;
        top: 0;
        bottom: 100px;
        background-color: white;
        font-size: 16px;
        font-family: sans-serif, system-ui, -apple-system, BlinkMacSystemFont;
        padding-bottom: 1px;
      }

      li {
        display: flex;
        flex-flow: row;
        align-items: center;
        gap: 8px;
        width: max-content;
        max-width: 100%;

        span {
          width: max-content;
          margin-left: 0px;

          &:first-child {
            min-width: max-content;
          }

          &:last-child {
            background-color: rgb(46, 46, 46);
            text-transform: capitalize;
            color: white;
            border-radius: 8px;
            padding: 10px 12px;
          }
        }
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
