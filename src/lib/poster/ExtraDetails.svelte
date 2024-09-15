<!-- Extra Details View For Posters -->
<script lang="ts">
  import { userSettings, wlDetailedView } from "@/store";
  import { RatingSystem, type PosterExtraDetails, type WatchedStatus } from "@/types";
  import { getOrdinalSuffix, monthsShort, watchedStatuses } from "../util/helpers";
  import Icon from "../Icon.svelte";
  import { toShowableRating, toWhichThumb } from "../rating/helpers";
  import { page } from "$app/stores";

  export let rating: number | undefined;
  export let status: WatchedStatus | undefined;
  export let details: PosterExtraDetails | undefined;

  $: dve = $wlDetailedView;
  $: settings = $userSettings;
  $: isUsingThumbs = settings && settings.ratingSystem === RatingSystem.Thumbs;

  function formatDate(e: number) {
    if (!e) {
      return "Unknown";
    }
    const d = new Date(e);
    return `${d.getDate()}${getOrdinalSuffix(d.getDate())} ${monthsShort[d.getMonth()]} '${String(
      d.getFullYear()
    ).substring(2, 4)}`;
  }
</script>

{#if ($page.url?.pathname === "/" || $page.url?.pathname.startsWith("/search/")) && details && dve && dve.length > 0}
  <div class="extra-details">
    <!--
      The `if` statements can't be on their own line to look pretty
      because that will cause whitespace in the generated markup,
      which causes the :empty css tag to not work.
      Can be reverted when its possible to trim whitespace in svelte
      OR when :empty tag is updated in browsers to new spec and counts whitespace as empty.
    -->
    <div>
      {#if details.dateAdded && dve.includes("dateAdded")}
        <span title="Date added to watch list">
          <i><Icon i="calendar" /></i>
          <span>
            {formatDate(Date.parse(details.dateAdded))}
          </span>
        </span>
      {/if}{#if details.dateModified && dve.includes("dateModified")}
        <span title="Date last modified">
          <i><Icon i="pencil" wh={15} /></i>
          <span>
            {formatDate(Date.parse(details.dateModified))}
          </span>
        </span>
      {/if}{#if details.lastWatched && dve.includes("lastWatched")}
        <span title="Latest season watched">
          <i><Icon i="play" wh={15} /></i>
          <span>{details.lastWatched}</span>
        </span>
      {/if}{#if dve.includes("statusRating")}
        <span class="status-rating" title="Status and Rating">
          {#if !isUsingThumbs}
            <i><Icon i="star" /></i>
          {/if}
          <span class="rating-span">
            {#if isUsingThumbs}
              {@const r = toWhichThumb(rating)}
              {#if r === -1}
                <i><Icon i="thumb-down" /></i>
              {:else if r === 0}
                <span
                  style="display: flex; transform: translate(2px, 3px); font-size: 40px; font-family: 'Shrikhand';"
                >
                  -
                </span>
              {:else if r === 1}
                <i><Icon i="thumb-up" /></i>
              {/if}
            {:else}
              {toShowableRating(rating)}
            {/if}
          </span>
          {#if status}
            <i><Icon i={watchedStatuses[status]} wh={15} /></i>
          {/if}
        </span>
      {/if}
    </div>
  </div>
{/if}

<style lang="scss">
  .extra-details {
    position: absolute;
    bottom: 5px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    flex-flow: column;
    justify-content: center;
    align-items: center;
    font-size: 14px;
    width: 160px;
    color: white;
    background-color: $poster-extra-detail-bg-color;
    border-radius: 10px;
    transition: opacity 100ms ease-out;

    & > div {
      padding: 8px 3px;

      &:empty {
        padding: 0px;
      }

      & > span {
        display: flex;
        flex-flow: row;
        align-items: center;
        gap: 8px;
        height: 15px;
        font-weight: bold;

        &.status-rating {
          gap: 10px;
        }

        &:not(:last-child) {
          margin-bottom: 5px;
        }

        i {
          display: flex;
          width: 15px;
          fill: white;
        }
      }
    }

    .status-rating i:last-of-type {
      margin-left: auto;
    }

    .status-rating .rating-span:empty {
      /* Can happen when content is not rated and using thumbs system */
      display: none;
    }
  }
</style>
