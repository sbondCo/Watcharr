<script lang="ts">
  import type { Activity } from "@/types";
  import { getOrdinalSuffix, months } from "./util/helpers";

  export let activity: Activity[] | undefined;

  let groupedActivities: { [index: string]: any };

  $: {
    groupedActivities = getGroupedActivity(activity);
  }

  function seasonAndEpToReadable(season: number, episode: number) {
    return `S${season ? String(season)?.padStart(2, "0") : "(unknown)"}E${episode ? String(episode)?.padStart(2, "0") : "(unknown)"}`;
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
      case "THOUGHTS_CHANGED":
        return "thoughts changed";
      case "THOUGHTS_REMOVED":
        return "thoughts removed";
      case "IMPORTED_WATCHED":
        return "imported";
      case "IMPORTED_RATING":
        if (a.data) {
          const data = JSON.parse(a.data);
          if (data.rating) {
            return `rating changed to ${data.rating}`;
          } else {
            return "added to watchlist with no rating";
          }
        }
        return "imported rating";
      case "IMPORTED_ADDED_WATCHED":
        return "Imported Watch Date";
      case "SEASON_ADDED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `season ${data.season} added as ${data.status?.toLowerCase()}`;
        }
        return "season added";
      case "SEASON_RATING_CHANGED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `changed season ${data.season} rating to ${data.rating}`;
        }
        return "season rating changed";
      case "SEASON_STATUS_CHANGED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `changed season ${data.season} status to ${data.status?.toLowerCase()}`;
        }
        return "season status changed";
      case "SEASON_REMOVED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `season ${data.season} status removed`;
        }
        return "season removed";
      case "EPISODE_ADDED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `${seasonAndEpToReadable(data.season, data.episode)} added ${data.status ? `as ${data.status?.toLowerCase()}` : data.rating ? `with rating ${data.rating}` : ""}`;
        }
        return "episode added";
      case "EPISODE_RATING_CHANGED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `${seasonAndEpToReadable(data.season, data.episode)} rating changed to ${data.rating}`;
        }
        return "episode rating changed";
      case "EPISODE_STATUS_CHANGED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `${seasonAndEpToReadable(data.season, data.episode)} status changed to ${data.status?.toLowerCase()}`;
        }
        return "episode status changed";
      case "EPISODE_REMOVED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `${seasonAndEpToReadable(data.season, data.episode)} removed`;
        }
        return "season removed";
      default:
        return a.type;
    }
  }

  function toDayTime(d: Date) {
    return `${d.getDate()}${getOrdinalSuffix(d.getDate())} at ${d.toLocaleTimeString()}`;
  }

  /**
   * Get what will be the visible date the activity was 'created'.
   * @returns customDate if defined or createdAt if not.
   */
  function getCreatedAtVis(a: Activity) {
    return Date.parse(a.customDate ?? a.createdAt);
  }

  function getGroupedActivity(activities?: Activity[]) {
    activities = activities?.filter((a) => a.type);
    const a = activities?.sort((a, b) => getCreatedAtVis(b) - getCreatedAtVis(a));
    let grouped: { [index: string]: any } = {};
    if (a) {
      for (let i = 0; i < a.length; i++) {
        const activity = a[i];
        const date = new Date(getCreatedAtVis(activity));
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
          {@const d = new Date(getCreatedAtVis(a))}
          <li>
            <span title={d.toDateString()}>
              {toDayTime(d)}
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
        font-size: 16px;
        font-family:
          sans-serif,
          system-ui,
          -apple-system,
          BlinkMacSystemFont;
        padding-bottom: 1px;
        background-color: $bg-color;
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
            background-color: $accent-color;
            text-transform: capitalize;
            color: $text-color;
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
    font-family:
      sans-serif,
      system-ui,
      -apple-system,
      BlinkMacSystemFont;
    font-size: 30px;
    font-weight: bold;
    margin-left: 30px;
  }
</style>
