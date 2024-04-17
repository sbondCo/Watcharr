<script lang="ts">
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import RequestMovie from "@/lib/request/RequestMovie.svelte";
  import RequestShow from "@/lib/request/RequestShow.svelte";
  import { baseURL } from "@/lib/util/api";
  import { toRelativeDate } from "@/lib/util/helpers";
  import { notify } from "@/lib/util/notify";
  import { type ArrRequestResponse, type TMDBMovieDetails, type TMDBShowDetails } from "@/types";
  import axios from "axios";

  let allRequests: ArrRequestResponse[];
  let showBeingApproved: TMDBShowDetails | undefined;
  let movieBeingApproved: TMDBMovieDetails | undefined;
  let beingApprovedOriginalRequest: ArrRequestResponse | undefined;

  async function getRequests() {
    try {
      allRequests = (await axios.get(`/arr/request/`)).data as ArrRequestResponse[];
      if (allRequests?.length > 0) {
        allRequests = allRequests?.sort((a, b) => {
          if (b.status === "PENDING") return 1;
          if (a.status === "PENDING") return -1;
          return Date.parse(b.updatedAt) - Date.parse(a.updatedAt);
        });
      }
    } catch (err) {
      console.error("Failed to get requests!", err);
      notify({ type: "error", text: "Failed when getting all requests!" });
    }
  }

  async function deny(r: ArrRequestResponse) {
    try {
      await axios.post(`/arr/request/deny/${r.id}`);
      getRequests();
    } catch (err) {
      console.error("Failed to deny request!", err);
      notify({ type: "error", text: "Failed when denying request!" });
    }
  }

  async function approve(r: ArrRequestResponse) {
    console.debug("Approving request:", r);
    if (r.content.type === "tv") {
      showBeingApproved = (await axios.get(`/content/tv/${r.content.tmdbId}`))
        .data as TMDBShowDetails;
    } else if (r.content.type === "movie") {
      movieBeingApproved = (await axios.get(`/content/movie/${r.content.tmdbId}`))
        .data as TMDBMovieDetails;
    } else {
      notify({ type: "error", text: "Unknown content type, can't continue approval!" });
      return;
    }
    beingApprovedOriginalRequest = r;
  }
</script>

<div class="content">
  <div class="inner">
    <h2>Media Requests</h2>
    <h5 class="norm">Manage and view all of your media requests.</h5>

    {#await getRequests()}
      <Spinner />
    {:then}
      <div class="request-container">
        {#each allRequests as r}
          <div class={`request ${r.content.type}`}>
            <div class="poster">
              <img src={`${baseURL}/img${r.content?.poster_path}`} alt="" loading="lazy" />
              <span title={r.serverName}>{r.serverName}</span>
            </div>
            <img
              class="backdrop"
              src={`${baseURL}/img${r.content?.poster_path}`}
              alt=""
              loading="lazy"
            />
            <div class="wordsnstuff">
              <h2 class="norm">
                <a
                  data-sveltekit-preload-data="tap"
                  href={`/${r.content.type}/${r.content.tmdbId}`}
                  class="plain"
                >
                  {r.content.title}
                </a>
                {#if r.content.release_date}
                  <span>{new Date(r.content.release_date).getFullYear()}</span>
                {/if}
              </h2>
              <p>{r.content.overview}</p>
              <div class="btns">
                {#if r.username}
                  <span
                    style="font-size: 12px; margin-top: auto; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;"
                  >
                    By {r.username} &#160;&#8226;&#160; {toRelativeDate(new Date(r.createdAt))}
                  </span>
                {/if}
                {#if r.status === "PENDING"}
                  <button class="decline" on:click={() => deny(r)}>Decline</button>
                  <button class="approve" on:click={() => approve(r)}>Approve</button>
                {:else}
                  <button disabled>{r.status}</button>
                {/if}
              </div>
            </div>
          </div>
        {/each}
      </div>
    {:catch err}
      <PageError error={err} pretty="Failed to fetch requests!" />
    {/await}

    {#if showBeingApproved}
      <RequestShow
        content={showBeingApproved}
        approveMode={true}
        originalRequest={beingApprovedOriginalRequest}
        onClose={() => {
          showBeingApproved = undefined;
          // HACK
          getRequests();
        }}
      />
    {:else if movieBeingApproved}
      <RequestMovie
        content={movieBeingApproved}
        approveMode={true}
        originalRequest={beingApprovedOriginalRequest}
        onClose={() => {
          movieBeingApproved = undefined;
          // HACK
          getRequests();
        }}
      />
    {/if}
  </div>
</div>

<style lang="scss">
  .request-container {
    display: flex;
    flex-flow: column;
    gap: 10px;
    margin-top: 20px;
  }

  .content .request {
    display: flex;
    flex-flow: row;
    position: relative;
    gap: 10px;
    max-height: 245px;
    padding: 10px;
    border-radius: 10px;
    background-color: $img-blend-multiply-bg-col;
    color: white;
    overflow: hidden;

    img {
      width: 150px;
      height: 225px;
      border-radius: 6px;

      &.backdrop {
        position: absolute;
        left: 0;
        top: 0;
        width: 100%;
        height: auto;
        transform: translateY(-25%);
        filter: blur(4px) grayscale(80%);
        mix-blend-mode: multiply;
        z-index: -2;

        @media screen and (max-width: 400px) {
          transform: unset;
          height: 100%;
        }
      }
    }

    &.tv {
      .poster > span {
        background-color: #35c5f4;
      }
    }

    .poster {
      position: relative;

      > span {
        position: absolute;
        bottom: 3px;
        left: 3px;
        color: black;
        background-color: #ffc230;
        padding: 3px 5px;
        font-size: 11px;
        border-radius: 5px;
        max-width: calc(100% - 6px);
        overflow: hidden;
        text-overflow: ellipsis;
      }
    }

    & > div {
      display: flex;
      flex-flow: column;
      gap: 5px;

      &.wordsnstuff {
        overflow: hidden;
        width: 100%;
      }

      & > h2 {
        font-size: 22px;
        white-space: wrap;
        overflow: visible;

        a {
          color: white;
        }

        & span:last-of-type {
          font-size: 16px;
          font-weight: 400;
          color: rgba(255, 255, 255, 0.7);
        }

        @media screen and (max-width: 600px) {
          text-align: center;
        }
      }

      & > p {
        font-size: 14px;
        margin-bottom: 5px;
        text-overflow: ellipsis;
        overflow: hidden;
      }

      .btns {
        display: flex;
        flex-flow: row;
        gap: 5px;
        margin-top: auto;
        padding-top: 3px;

        & button {
          width: fit-content;

          &:first-of-type {
            margin-left: auto;
          }

          &.approve {
            background-color: $success;
            color: white;

            &:hover {
              background-color: $text-color;
              color: $bg-color;
            }
          }
        }

        @media screen and (max-width: 400px) {
          flex-flow: column;
          align-items: center;
          gap: 10px;

          & button:first-of-type {
            margin-left: unset;
          }
        }
      }
    }

    @media screen and (max-width: 600px) {
      flex-flow: column;
      max-height: unset;
      align-items: center;

      .wordsnstuff p {
        display: none;
      }
    }
  }

  .content {
    display: flex;
    width: 100%;
    justify-content: center;
    padding: 0 30px 30px 30px;

    .inner {
      min-width: 700px;
      max-width: 700px;
      overflow: hidden;

      h2 {
        overflow: hidden;
        text-overflow: ellipsis;
      }

      & > div:not(:first-of-type) {
        margin-top: 30px;
      }

      @media screen and (max-width: 740px) {
        width: 100%;
        min-width: unset;
      }
    }
  }
</style>
