<script lang="ts">
  import type { WatchedStatus } from "@/types";
  import Icon from "./Icon.svelte";
  import { isTouch, watchedStatuses } from "./helpers";
  import { goto } from "$app/navigation";
  import tooltip from "./actions/tooltip";

  export let poster: string | undefined;
  export let title: string | undefined;
  export let desc: string | undefined;
  export let rating: number | undefined = undefined;
  export let status: WatchedStatus | undefined = undefined;
  export let link: string | undefined = undefined;
  export let onStatusChanged: (type: WatchedStatus) => void = () => {};
  export let onRatingChanged: (rating: number) => void = () => {};

  // If poster is active (scaled up)
  let posterActive = false;
  // If ratings are shown
  let ratingsShown = false;
  // If statuses are shown
  let statusesShown = false;

  function handleStarClick(r: number) {
    if (r == rating) return;
    onRatingChanged(r);
    ratingsShown = false;
  }

  function handleStatusClick(type: WatchedStatus) {
    onStatusChanged(type);
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
      <img loading="lazy" src={poster} alt="poster" />
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
      <span>{desc}</span>

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
              -webkit-text-stroke: 1.5px black;

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
              -webkit-text-stroke: 1.5px gold;
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

            button {
              width: 100%;
              color: black;
              fill: black;
              font-size: 20px;

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
