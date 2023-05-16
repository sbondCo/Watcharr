<script lang="ts">
  import type { MediaType, WatchedStatus } from "@/types";
  import Icon from "./Icon.svelte";
  import { isTouch, watchedStatuses } from "./helpers";
  import { goto } from "$app/navigation";
  import tooltip from "./actions/tooltip";

  export let media: {
    poster_path?: string;
    title?: string;
    name?: string;
    overview?: string;
    id?: number;
    media_type?: MediaType;
  };
  export let rating: number | undefined = undefined;
  export let status: WatchedStatus | undefined = undefined;
  export let onStatusChanged: (type: WatchedStatus) => void = () => {};
  export let onRatingChanged: (rating: number) => void = () => {};
  export let onDeleteClicked: () => void = () => {};

  // If poster is active (scaled up)
  let posterActive = false;
  // If ratings are shown
  let ratingsShown = false;
  // If statuses are shown
  let statusesShown = false;

  const title = media.title || media.name;
  const poster = `https://image.tmdb.org/t/p/w500${media.poster_path}`;
  const link = media.id ? `/${media.media_type}/${media.id}` : undefined;

  function handleStarClick(r: number) {
    if (r == rating) return;
    onRatingChanged(r);
    ratingsShown = false;
  }

  function handleStatusClick(type: WatchedStatus | "DELETE") {
    if (type === "DELETE") {
      onDeleteClicked();
      return;
    }
    onStatusChanged(type);
  }

  function addClassToParent(
    e: Event & {
      currentTarget: EventTarget & Element;
    },
    c: string
  ) {
    (e.currentTarget?.parentNode as HTMLDivElement)?.classList.add(c);
  }
</script>

<li
  on:mouseenter={() => {
    if (!isTouch()) posterActive = true;
  }}
  on:mouseleave={() => (posterActive = false)}
  on:click={() => (posterActive = true)}
  on:keypress={() => console.log("on kpress")}
  class={posterActive ? "active" : ""}
>
  <div class={`container${!poster ? " details-shown" : ""}`}>
    {#if poster}
      <div class="img-loader" />
      <img
        loading="lazy"
        src={poster}
        alt=""
        on:load={(e) => {
          addClassToParent(e, "img-loaded");
        }}
        on:error={(e) => {
          addClassToParent(e, "details-shown");
        }}
      />
    {/if}
    <div
      on:click={() => {
        if (posterActive && link) goto(link);
      }}
      on:keypress={() => console.log("on kpress")}
      class="inner"
    >
      <h2>
        {#if link}
          <a data-sveltekit-preload-data="tap" href={link}>
            {title}
          </a>
        {:else}
          {title}
        {/if}
      </h2>
      <span>{media.overview}</span>

      <div class="buttons">
        <button
          class="rating"
          on:click={(ev) => {
            ev.stopPropagation();
            ratingsShown = !ratingsShown;
          }}
          on:mouseleave={(ev) => {
            ratingsShown = false;
            ev.currentTarget.blur();
          }}
        >
          <span>*</span>
          <span>{rating ? rating : "Rate"}</span>
          {#if ratingsShown}
            <div>
              {#each [10, 9, 8, 7, 6, 5, 4, 3, 2, 1] as v}
                <button
                  class="plain{rating === v ? ' active' : ''}"
                  on:click={(ev) => {
                    ev.stopPropagation();
                    handleStarClick(v);
                  }}
                >
                  {v}
                </button>
              {/each}
            </div>
          {/if}
        </button>
        <button
          class="status"
          on:click={(ev) => {
            ev.stopPropagation();
            statusesShown = !statusesShown;
          }}
          on:mouseleave={(ev) => {
            statusesShown = false;
            ev.currentTarget.blur();
          }}
        >
          {#if status}
            <Icon i={watchedStatuses[status]} />
          {:else}
            <span class="no-icon">+</span>
          {/if}
          {#if statusesShown}
            <div>
              {#each Object.entries(watchedStatuses) as [statusName, icon]}
                <button
                  class="plain{status && status !== statusName ? ' not-active' : ''}"
                  on:click={() => handleStatusClick(statusName)}
                  use:tooltip={{ text: statusName }}
                >
                  <Icon i={icon} />
                </button>
              {/each}
              {#if status}
                <button
                  class="plain not-active"
                  on:click={() => handleStatusClick("DELETE")}
                  use:tooltip={{ text: "Delete" }}
                >
                  <Icon i="trash" />
                </button>
              {/if}
            </div>
          {/if}
        </button>
      </div>
    </div>
  </div>
</li>

<style lang="scss">
  li.active {
    cursor: pointer;
  }

  .container {
    display: flex;
    flex-flow: column;
    background-color: rgb(48, 45, 45);
    overflow: hidden;
    flex: 1 1;
    border-radius: 5px;
    width: 170px;
    height: 100%;
    min-height: 256.367px;
    position: relative;
    // aspect-ratio: 2/3;
    transition: transform 150ms ease;

    img {
      width: 100%;
      height: 100%;
    }

    &.img-loaded .img-loader,
    &.details-shown .img-loader {
      display: none;
    }

    .img-loader {
      position: absolute;
      width: 100%;
      height: 100%;
      background-color: gray;
      background: linear-gradient(359deg, #5c5c5c, #2c2929, #2c2424);
      background-size: 400% 400%;
      animation: imgloader 4s ease infinite;

      @-webkit-keyframes imgloader {
        0% {
          background-position: 50% 0%;
        }
        50% {
          background-position: 50% 100%;
        }
        100% {
          background-position: 50% 0%;
        }
      }
    }

    .inner {
      position: absolute;
      opacity: 0;
      display: flex;
      flex-flow: column;
      top: 0;
      height: 100%;
      width: 100%;
      padding: 10px;
      background-color: transparent;
      transition: opacity 150ms cubic-bezier(0.19, 1, 0.22, 1);

      h2 {
        font-family: sans-serif, system-ui, -apple-system, BlinkMacSystemFont;
        font-size: 18px;
        color: white;
      }

      span {
        color: white;
        margin: 5px 0 5px 0;
        font-size: 9px;
        display: -webkit-box;
        -webkit-line-clamp: 5;
        -webkit-box-orient: vertical;
        hyphens: auto;
        overflow: hidden;
      }

      .buttons {
        display: flex;
        flex-flow: row;
        margin-top: auto;
        gap: 10px;
        height: 35px;

        button {
          padding: 3px;
          position: relative;
          font-family: "Rampart One";

          /** Rating */
          &.rating {
            span {
              color: black;
              /* -webkit-text-stroke: 1.5px black; */

              &:first-child {
                font-size: 39px;
                letter-spacing: 10px;
                line-height: 52px;
                height: 42px;
              }

              &:nth-child(2) {
                font-size: 22px;
                height: 35px; // quick fix to make the rating num look centered - text-stroke makes it look not centered
              }
            }

            &:hover span,
            &:focus-visible span {
              color: gold;
              /* -webkit-text-stroke: 1.5px gold; */
            }

            div button {
              font-size: 20px;
            }
          }

          /** Status */
          &.status {
            width: 40%;

            .no-icon {
              color: black;
              font-size: 30px;
              height: 52px;
            }

            &:hover .no-icon,
            &:focus-visible .no-icon {
              color: white;
            }
          }

          div {
            display: flex;
            flex-flow: column;
            position: absolute;
            width: 100%;
            height: 200px;
            background-color: white;
            top: calc(-100% - 170px);
            list-style: none;
            border-radius: 4px 4px 0 0;
            overflow: auto;
            scrollbar-width: thin;

            button {
              width: 100%;
              color: black;
              fill: black;

              & :global(svg) {
                width: 100%;
                padding: 0 2px;
              }

              &:hover,
              &:focus-visible {
                background-color: rgb(100, 100, 100, 0.25);
              }
            }
          }
        }
      }
    }

    &:hover,
    &:focus-within {
      transform: scale(1.3);
      z-index: 99;
    }

    &:hover,
    &:focus-within,
    &:global(.details-shown) {
      img {
        filter: blur(4px) grayscale(80%);
        // This makes the background very dark,
        // but atleast the text is visible.. may want to change later.
        mix-blend-mode: multiply;
      }

      .inner {
        color: white;
        opacity: 1;
      }
    }
  }
</style>
