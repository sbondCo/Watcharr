<script lang="ts">
  import type {
    TMDBSeasonDetailsEpisode,
    WatchedStatus,
    WatchedEpisodeAddResponse,
    Watched
  } from "@/types";
  import Icon from "./Icon.svelte";
  import { userSettings } from "@/store";
  import PosterRating from "./poster/PosterRating.svelte";
  import PosterStatus from "./poster/PosterStatus.svelte";
  import axios from "axios";
  import { notify } from "./util/notify";
  import { get } from "svelte/store";
  import { watchedList } from "@/store";

  $: settings = $userSettings;

  export let ep: TMDBSeasonDetailsEpisode;
  export let watchedItem: Watched | undefined;

  let isHidden = false;

  $: {
    if (settings) isHidden = settings.hideSpoilers;
  }

  function updateWatchedEpisode(status?: WatchedStatus, rating?: number) {
    if (!watchedItem) {
      console.error("SeasonListEpisode: updateWatchedEpisode: No watched item.");
      return;
    }
    const nid = notify({ text: `Saving`, type: "loading" });
    axios
      .post<WatchedEpisodeAddResponse>(`/watched/episode`, {
        watchedId: watchedItem.id,
        seasonNumber: ep.season_number,
        episodeNumber: ep.episode_number,
        status,
        rating
      })
      .then((r) => {
        const wList = get(watchedList);
        const wEntry = wList.find((w) => w.id === watchedItem.id);
        if (!wEntry) {
          notify({
            id: nid,
            text: `Request succeeded, but failed to find local data. Please refresh.`,
            type: "error"
          });
          return;
        }
        if (r.status === 200) {
          wEntry.watchedEpisodes = r.data.watchedEpisodes;
          if (wEntry.activity?.length > 0) {
            wEntry.activity.push(r.data.addedActivity);
          } else {
            wEntry.activity = [r.data.addedActivity];
          }
          try {
            const epHookResp = r?.data?.episodeStatusChangedHookResponse;
            if (epHookResp && Object.keys(epHookResp).length > 0) {
              if (epHookResp.errors && epHookResp.errors.length > 0) {
                console.error(
                  "episodeStatusChangedHookResponse contained errors! All possible automations may not have been completed.",
                  epHookResp.errors
                );
                notify({
                  type: "error",
                  text: "Some automations have failed, check console for more info."
                });
              }
              if (epHookResp.addedActivities && epHookResp.addedActivities.length > 0) {
                wEntry.activity.push(...epHookResp.addedActivities);
              }
              if (epHookResp.watchedSeason) {
                if (!wEntry.watchedSeasons) {
                  wEntry.watchedSeasons = [epHookResp.watchedSeason];
                } else {
                  const watchedSeasonIdx = wEntry.watchedSeasons.findIndex(
                    (s) => s.id === epHookResp.watchedSeason?.id
                  );
                  if (watchedSeasonIdx === -1) {
                    wEntry.watchedSeasons.push(epHookResp.watchedSeason);
                  } else {
                    wEntry.watchedSeasons[watchedSeasonIdx] = epHookResp.watchedSeason;
                  }
                }
              }
              if (epHookResp.newShowStatus) {
                wEntry.status = epHookResp.newShowStatus;
              }
            }
          } catch (err) {
            console.error("Failed to process episodeStatusChangedHookResponse", err);
            notify({
              type: "error",
              text: "Failed to process automation response, check console for more info."
            });
          }
          watchedList.update((w) => w);
          notify({ id: nid, text: `Saved!`, type: "success" });
        }
      })
      .catch((err) => {
        console.error(err);
        notify({ id: nid, text: "Failed To Update!", type: "error" });
      });
  }

  function handleStatusClick(type: WatchedStatus | "DELETE") {
    if (!watchedItem) {
      console.error("SeasonListEpisode: handleStatusClick: No watched item.");
      return;
    }
    if (type === "DELETE") {
      const ws = watchedItem.watchedEpisodes?.find(
        (s) => s.seasonNumber === ep.season_number && s.episodeNumber === ep.episode_number
      );
      if (!ws) {
        notify({
          text: "Failed to find watched episode id. Please try refreshing.",
          type: "error"
        });
        return;
      }
      const nid = notify({ text: `Saving`, type: "loading" });
      axios
        .delete(`/watched/episode/${ws.id}`)
        .then((r) => {
          const wList = get(watchedList);
          const wEntry = wList.find((w) => w.id === watchedItem.id);
          if (!wEntry) {
            notify({
              id: nid,
              text: `Request succeeded, but failed to find local data. Please refresh.`,
              type: "error"
            });
            return;
          }
          if (r.status === 200) {
            wEntry.watchedEpisodes = wEntry.watchedEpisodes?.filter((s) => s.id !== ws.id);
            if (r.data) {
              if (wEntry.activity?.length > 0) {
                wEntry.activity.push(r.data);
              } else {
                wEntry.activity = [r.data];
              }
            }
            watchedList.update((w) => w);
            notify({ id: nid, text: `Removed!`, type: "success" });
          }
        })
        .catch((err) => {
          console.error(err);
          notify({ id: nid, text: "Failed To Remove!", type: "error" });
        });
      return;
    }
    updateWatchedEpisode(type);
  }

  function handleStarClick(rating: number) {
    updateWatchedEpisode(undefined, rating);
  }
</script>

<li class={isHidden ? "dont-spoil" : ""}>
  {#if ep.still_path}
    <img src={`https://www.themoviedb.org/t/p/w227_and_h127_bestv2/${ep.still_path}`} alt="" />
  {:else}
    <div class="no-still" />
  {/if}
  <div class="info">
    <div>
      <span>
        <b>{ep.episode_number}</b>
        <span class="episode-name">{ep.name}</span>
        {#if ep.runtime}
          <span class="episode-runtime" title="This episode has a runtime of {ep.runtime} minutes."
            >{ep.runtime} min</span
          >
        {/if}
      </span>
      <span
        class="rating"
        title={`TMDB Rating: ${ep.vote_average} out of 10 (based on ${ep.vote_count} votes)`}
      >
        <span>*</span>
        {Math.round(ep.vote_average * 10) / 10}
      </span>
    </div>
    <span class="overview">{ep.overview}</span>
  </div>
  {#if watchedItem}
    {@const we = watchedItem.watchedEpisodes?.find(
      (s) => s.seasonNumber === ep.season_number && s.episodeNumber === ep.episode_number
    )}
    <div class="status-rating-ctr">
      <div class="rating" style={"width: 45px"}>
        <PosterRating
          rating={we?.rating}
          btnTooltip={`Episode ${ep.episode_number} Rating`}
          handleStarClick={(r) => handleStarClick(r)}
          minimal={true}
          direction="bot"
          hideStarWhenRated
        />
      </div>
      <div class="status">
        <PosterStatus
          status={we?.status}
          btnTooltip={`Episode ${ep.episode_number} Status`}
          handleStatusClick={(t) => handleStatusClick(t)}
          direction="bot"
          width="100%"
          small
        />
      </div>
    </div>
  {/if}
  {#if isHidden}
    <button class="plain spoiler-text" on:click={() => (isHidden = false)}>
      <Icon i="eye-closed" wh={34} />
      <span>Click To Reveal</span>
    </button>
  {/if}
</li>

<style lang="scss">
  li {
    display: flex;
    flex-flow: row;
    gap: 8px;
    position: relative;

    img,
    .no-still {
      width: 227px;
      min-width: 227px;
      height: 127px;
      min-height: 127px;
      border-radius: 10px;
      background-color: rgb(0, 0, 0);
      object-fit: fill;

      @media screen and (max-width: 590px) {
        width: 80%;
        height: auto;
      }

      @media screen and (max-width: 450px) {
        width: 100%;
      }
    }

    .info {
      display: flex;
      flex-flow: column;

      & > div {
        display: flex;
        flex-flow: row;
        align-items: center;

        .episode-name,
        .episode-runtime {
          text-transform: lowercase;
          font-variant: small-caps;
          font-weight: bold;
          font-size: 16px;
        }

        .episode-runtime {
          font-size: 14px;
          padding: 0 2px;
        }

        .rating {
          display: flex;
          align-items: start;
          justify-content: center;
          font-size: 15px;
          color: $rating-color;
          font-weight: bolder;
          overflow: hidden;

          span {
            margin-top: 2px;
            font-family: "Rampart One";
            -webkit-text-stroke: 1px $rating-color;
            font-size: 25px;
            line-height: 0.7;
          }
        }
      }
    }

    .status-rating-ctr {
      display: flex;
      align-items: center;
      flex-flow: column-reverse;
      gap: 10px;
      margin-bottom: auto;
      min-height: 40px;
      margin-left: auto;

      div {
        transition: width 100ms ease;

        &:first-of-type {
          margin-left: auto;
        }

        &.rating {
          height: 40px;
          min-height: 40px;
        }

        &.status {
          width: 45px;
          min-height: 40px;
          height: 40px;
          overflow: visible;
        }
      }
    }

    span {
      padding: 3px 5px;

      @media screen and (max-width: 590px) {
        text-align: center;
      }
    }

    .spoiler-text {
      display: flex;
      flex-flow: column;
      align-items: center;
      justify-content: center;
      gap: 8px;
      position: absolute;
      width: 100%;
      height: 100%;
      font-weight: bolder;
      font-size: 20px;
      fill: $text-color;
      opacity: 0;
      transition:
        visibility 150ms ease-in,
        opacity 150ms ease-in;
      cursor: pointer;

      span {
        text-shadow: 0 0 6px $bg-color;
      }

      :global(svg) {
        filter: drop-shadow(0 0 8px $bg-color);
      }

      &:hover,
      &:active,
      &:focus {
        opacity: 1;
      }
    }

    img,
    .episode-name,
    .rating,
    .overview {
      transition: filter 150ms ease-out;
    }

    &.dont-spoil {
      .episode-name,
      .rating,
      .overview {
        filter: blur(4px);
      }

      img {
        filter: blur(6px);
      }
    }
  }

  @media screen and (max-width: 590px) {
    li {
      align-items: center;
      flex-flow: column;
      width: 100%;
      height: 100%;

      .status-rating-ctr {
        flex-flow: row;
        justify-content: center;
        margin-left: unset;
      }
    }

    .rating {
      margin-left: auto;
    }
  }
</style>
