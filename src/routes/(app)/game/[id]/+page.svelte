<script lang="ts">
  import PageError from "@/lib/PageError.svelte";
  import PersonPoster from "@/lib/poster/PersonPoster.svelte";
  import Rating from "@/lib/Rating.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import Status from "@/lib/Status.svelte";
  import HorizontalList from "@/lib/HorizontalList.svelte";
  import { serverFeatures, watchedList } from "@/store";
  import { GameWebsiteCategory, type GameDetailsResponse, type WatchedStatus } from "@/types";
  import axios from "axios";
  import Activity from "@/lib/Activity.svelte";
  import Title from "@/lib/content/Title.svelte";
  import VideoEmbedModal from "@/lib/content/VideoEmbedModal.svelte";
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import Error from "@/lib/Error.svelte";
  import FollowedThoughts from "@/lib/content/FollowedThoughts.svelte";
  import { removeWatched, updatePlayed } from "@/lib/util/api.js";
  import GamePoster from "@/lib/poster/GamePoster.svelte";
  import { getPlayedDependedProps } from "@/lib/util/helpers";
  import tooltip from "@/lib/actions/tooltip.js";
  import Icon from "@/lib/Icon.svelte";
  import MyThoughts from "@/lib/content/MyThoughts.svelte";
  import AddToTagButton from "@/lib/tag/AddToTagButton.svelte";

  export let data;

  let trailer: string | undefined;
  let trailerShown = false;

  $: wList = $watchedList;
  $: wListItem = $watchedList.find((w) => w.game?.igdbId === data.gameId);

  let gameId: number | undefined;
  let game: GameDetailsResponse | undefined;
  let pageError: Error | undefined;

  onMount(() => {
    const unsubscribe = page.subscribe((value) => {
      console.log(value);
      const params = value.params;
      if (params && params.id) {
        gameId = Number(params.id);
      }
    });

    return unsubscribe;
  });

  $: {
    (async () => {
      try {
        game = undefined;
        pageError = undefined;
        if (!gameId) {
          return;
        }
        const data = (await axios.get(`/game/${gameId}`)).data as GameDetailsResponse;
        if (data.videos?.length > 0) {
          const t = data.videos.find((v) => v.name?.toLowerCase() === "trailer");
          // Doc says the video_id is "usually youtube", so we are gonna go with that assumption too ( 0 _ 0 )
          if (t?.video_id) {
            trailer = `https://www.youtube.com/embed/${t?.video_id}`;
          }
        }
        game = data;
      } catch (err: any) {
        game = undefined;
        pageError = err;
      }
    })();
  }

  async function contentChanged(
    newStatus?: WatchedStatus,
    newRating?: number,
    newThoughts?: string,
    pinned?: boolean
  ): Promise<boolean> {
    if (!gameId) {
      console.error("contentChanged: no gameId");
      return false;
    }
    return await updatePlayed(gameId, newStatus, newRating, newThoughts, pinned);
  }
</script>

{#if pageError}
  <PageError pretty="Failed to load game!" error={pageError} />
{:else if !game}
  <Spinner />
{:else if Object.keys(game).length > 0}
  <div>
    <div class="content">
      {#if game?.artworks?.length > 0}
        <img
          class="backdrop"
          src={"https://images.igdb.com/igdb/image/upload/t_720p/" +
            game.artworks[Math.floor(Math.random() * game.artworks.length)].image_id +
            ".jpg"}
          alt=""
        />
      {:else if game?.cover?.image_id}
        <!-- Fallback to using the game cover for backdrop if there is no artwork -->
        <img
          class="backdrop"
          src={"https://images.igdb.com/igdb/image/upload/t_720p/" + game.cover.image_id + ".jpg"}
          alt=""
        />
      {/if}
      <div class="vignette" />

      <div class="details-container">
        <img
          class="poster"
          src={"https://images.igdb.com/igdb/image/upload/t_cover_big/" +
            game.cover.image_id +
            ".jpg"}
          alt=""
        />

        <div class="details">
          <Title
            title={game.name}
            homepage={game.websites?.find((w) => w.category == GameWebsiteCategory.Official)?.url}
            releaseYear={new Date(game.first_release_date).getFullYear()}
            voteAverage={game.rating}
            voteCount={game.rating_count}
          />

          <span class="quick-info">
            {#if game.genres?.length > 0}
              <div>
                {#each game.genres as g, i}
                  <span>{g.name}{i !== game.genres.length - 1 ? ", " : ""}</span>
                {/each}
              </div>
            {:else}
              <span>Unknown Genres</span>
            {/if}
            <span></span>
            <div>
              {#if game.game_modes?.length > 0}
                {#each game.game_modes as g, i}
                  <span>{g.name}{i !== game.game_modes.length - 1 ? ", " : ""}</span>
                {/each}
              {:else}
                <span>Unknown Game Modes</span>
              {/if}
            </div>
          </span>

          <span style="font-weight: bold; font-size: 14px;">Overview</span>
          <p>{game.summary}</p>

          <div class="btns">
            {#if trailer}
              <button on:click={() => (trailerShown = !trailerShown)}>View Trailer</button>
              {#if trailerShown}
                <VideoEmbedModal embed={trailer} closed={() => (trailerShown = false)} />
              {/if}
            {/if}
            {#if wListItem}
              <div class="other-side">
                <AddToTagButton watchedItem={wListItem} />
                <button
                  on:click={() => {
                    if (wListItem?.pinned) {
                      contentChanged(undefined, undefined, undefined, false);
                    } else {
                      contentChanged(undefined, undefined, undefined, true);
                    }
                  }}
                  use:tooltip={{
                    text: `${wListItem?.pinned ? "Unpin from" : "Pin to"} top of list`,
                    pos: "bot"
                  }}
                >
                  <Icon i={wListItem?.pinned ? "unpin" : "pin"} wh={19} />
                </button>
                <button
                  class="delete-btn"
                  on:click={() =>
                    wListItem
                      ? removeWatched(wListItem.id)
                      : console.error("no wlistItem.. can't delete")}
                  use:tooltip={{ text: "Delete", pos: "bot" }}
                >
                  <Icon i="trash" wh={19} />
                </button>
              </div>
            {/if}
          </div>

          <!-- <ProvidersList providers={game["watch/providers"]} /> -->
        </div>
      </div>
    </div>

    <div class="page">
      <div class="review">
        <Rating rating={wListItem?.rating} onChange={(n) => contentChanged(undefined, n)} />
        <Status status={wListItem?.status} isForGame={true} onChange={(n) => contentChanged(n)} />
        {#if wListItem}
          <MyThoughts
            contentTitle={game.name}
            thoughts={wListItem?.thoughts}
            onChange={(newThoughts) => {
              return contentChanged(undefined, undefined, newThoughts);
            }}
          />
        {/if}
      </div>

      {#if gameId}
        <FollowedThoughts mediaType="game" mediaId={gameId} />
      {/if}

      {#if game.similar_games?.length > 0}
        <HorizontalList title="Similar">
          {#each game.similar_games as g}
            <GamePoster
              media={{
                id: g.id,
                coverId: g.cover.image_id,
                name: g.name,
                summary: g.summary,
                firstReleaseDate: g.first_release_date
              }}
              {...getPlayedDependedProps(g.id, wList)}
              small={true}
            />
          {/each}
        </HorizontalList>
      {/if}

      {#if wListItem}
        <Activity wListId={wListItem.id} activity={wListItem.activity} />
      {/if}
    </div>
  </div>
{:else}
  <Error error="Game not found" pretty="Game not found" />
{/if}

<style lang="scss">
  .content {
    position: relative;
    color: white;

    img.backdrop {
      position: absolute;
      left: 0;
      top: 0;
      z-index: -2;
      width: 100%;
      height: 100%;
      object-fit: cover;
      filter: $backdrop-filter;
      mix-blend-mode: $backdrop-mix-blend-mode;
      mask-image: $backdrop-mask-image;
    }

    .vignette {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background-color: rgba($color: #000000, $alpha: 0.7);
      z-index: -1;
      mask-image: $backdrop-mask-image;
    }

    .details-container {
      display: flex;
      flex-flow: row;
      gap: 35px;
      max-width: 1100px;
      padding: 40px 80px;
      margin-left: auto;
      margin-right: auto;

      img.poster {
        width: 235px;
        height: 100%;
        box-shadow: 0px 0px 14px -4px #9c8080;
        border-radius: 12px;
      }

      .details {
        display: flex;
        flex-flow: column;
        gap: 5px;

        .quick-info {
          display: flex;
          gap: 10px;
          margin-bottom: 8px;
        }

        p {
          font-size: 14px;
          margin-bottom: 18px;
        }

        .btns {
          display: flex;
          flex-flow: row;
          flex-wrap: wrap;
          gap: 8px;
          margin-top: auto;

          a.btn,
          button {
            max-width: fit-content;
            overflow: hidden;
            animation: 50ms cubic-bezier(0.86, 0, 0.07, 1) forwards otherbtn;
            white-space: nowrap;
            gap: 6px;
            justify-content: flex-start;
            font-size: 14px;

            @keyframes otherbtn {
              from {
                width: 0px;
              }
              to {
                width: 100%;
              }
            }
          }

          .other-side {
            display: flex;
            flex-flow: row;
            gap: 8px;

            @media screen and (min-width: 900px) {
              margin-left: auto;
            }
          }

          .delete-btn {
            &:hover {
              color: $error;
            }
          }
        }
      }

      @media screen and (max-width: 700px) {
        padding: 40px;
      }

      @media screen and (max-width: 590px) {
        flex-flow: column;
        align-items: center;
      }
    }
  }

  .page {
    display: flex;
    flex-flow: column;
    align-items: center;
    margin-left: auto;
    margin-right: auto;
    gap: 30px;
    padding: 20px 50px;
    max-width: 1200px;

    @media screen and (max-width: 500px) {
      padding: 20px;
    }
  }

  .review {
    display: flex;
    flex-flow: column;
    gap: 10px;
    max-width: 380px;

    @media screen and (max-width: 420px) {
      max-width: 340px;
    }
  }

  .creators {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 35px;
    margin: 10px 60px;

    div {
      display: flex;
      flex-flow: column;
      min-width: 150px;

      span:first-child {
        font-weight: bold;
      }
    }
  }
</style>
