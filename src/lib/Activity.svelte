<script lang="ts">
  import type { Activity } from "@/types";
  import { getOrdinalSuffix, months, seasonAndEpToReadable } from "./util/helpers";
  import ActivityEditor from "./ActivityEditor.svelte";

  export let activity: Activity[] | undefined;
  export let wListId: number;

  let clickedActivity: Activity;
  let groupedActivities: { [index: string]: any };
  let isActivityEditorVisible: boolean;

  $: {
    groupedActivities = getGroupedActivity(activity);
  }

  function getMsg(a: Activity) {
    switch (a?.type) {
      case "ADDED_WATCHED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `Added to Watched List${data?.status ? ` as ${toFullTitleCase(data.status)}` : ""}${
            data?.rating ? ` with ${data.rating} Stars` : ""
          }`;
        }
        return "Added to Watched List";
      case "REMOVED_WATCHED":
        return "Removed from Watched List";
      case "RATING_CHANGED":
        if (a.data) {
          return `Rating Changed to ${a.data}`;
        }
        return "Rating Changed";
      case "STATUS_CHANGED":
        if (a.data) {
          return `Status Changed to ${toFullTitleCase(a.data)}`;
        }
        return "Status Changed";
      case "THOUGHTS_CHANGED":
        return "Thoughts Changed";
      case "THOUGHTS_REMOVED":
        return "Thoughts Removed";
      case "IMPORTED_WATCHED":
        return "Imported";
      case "IMPORTED_WATCHED_JF":
      case "IMPORTED_WATCHED_PLEX":
        return "Synced";
      case "IMPORTED_RATING":
        if (a.data) {
          const data = JSON.parse(a.data);
          if (data.rating) {
            return `Rating Changed to ${data.rating}`;
          } else {
            return "Added to Watchlist with No Rating";
          }
        }
        return "Imported Rating";
      case "IMPORTED_ADDED_WATCHED":
      case "IMPORTED_ADDED_WATCHED_JF":
      case "IMPORTED_ADDED_WATCHED_PLEX":
        return "Imported Watch Date";
      case "SEASON_ADDED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `Season ${data.season} Added as ${toFullTitleCase(data.status)}`;
        }
        return "Season Added";
      case "SEASON_ADDED_JF":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `Season ${data.season} Synced as ${toFullTitleCase(data.status)}`;
        }
        return "Season Synced";
      case "SEASON_RATING_CHANGED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `Changed Season ${data.season} Rating to ${data.rating}`;
        }
        return "Season Rating Changed";
      case "SEASON_STATUS_CHANGED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `Changed Season ${data.season} Status to ${toFullTitleCase(data.status)}`;
        }
        return "Season Status Changed";
      case "SEASON_REMOVED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `Season ${data.season} Status Removed`;
        }
        return "Season Removed";
      case "EPISODE_ADDED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `${seasonAndEpToReadable(data.season, data.episode)} Added ${data.status ? `as ${toFullTitleCase(data.status)}` : data.rating ? `with Rating ${data.rating}` : ""}`;
        }
        return "Episode Added";
      case "EPISODE_ADDED_JF":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `${seasonAndEpToReadable(data.season, data.episode)} Synced ${data.status ? `as ${toFullTitleCase(data.status)}` : data.rating ? `with Rating ${data.rating}` : ""}`;
        }
        return "Episode Synced";
      case "EPISODE_RATING_CHANGED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `${seasonAndEpToReadable(data.season, data.episode)} Rating Changed to ${data.rating}`;
        }
        return "Episode Rating Changed";
      case "EPISODE_STATUS_CHANGED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `${seasonAndEpToReadable(data.season, data.episode)} Status Changed to ${toFullTitleCase(data.status)}`;
        }
        return "Episode Status Changed";
      case "EPISODE_REMOVED":
        if (a.data) {
          const data = JSON.parse(a.data);
          return `${seasonAndEpToReadable(data.season, data.episode)} Removed`;
        }
        return "Episode Removed";
      default:
        return a.type;
    }
  }

  function toFullTitleCase(text: string | undefined) {
    if (text) {
      return text
        .split(" ")
        .map((l) => l[0].toUpperCase() + l.substring(1).toLowerCase())
        .join(" ");
    }
    return "Unknown";
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

  function openEditor(a: Activity) {
    clickedActivity = a;
    isActivityEditorVisible = true;
    return;
  }
</script>

{#if isActivityEditorVisible}
  <ActivityEditor
    watchedId={wListId}
    activity={clickedActivity}
    activityMessage={getMsg(clickedActivity)}
    onClose={() => (isActivityEditorVisible = false)}
  />
{/if}

<div class="activity">
  <h2>Activity</h2>
  {#if groupedActivities && Object.keys(groupedActivities).length > 0}
    <ul>
      {#each Object.keys(groupedActivities) as k}
        <h3>{k}</h3>

        {#each groupedActivities[k] as a}
          {@const d = new Date(getCreatedAtVis(a))}
          <li>
            <button class="unset" on:click={() => openEditor(a)}>
              <span title={d.toDateString()}>{toDayTime(d)}</span>
              <span>{getMsg(a)}</span>
            </button>
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
        button {
          all: unset;
          display: flex;
          flex-flow: row;
          align-items: center;
          gap: 8px;
          width: max-content;
          max-width: 100%;
          cursor: pointer;
        }

        span {
          width: max-content;
          margin-left: 0px;

          &:first-child {
            min-width: max-content;
          }

          &:last-child {
            background-color: $accent-color;
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
