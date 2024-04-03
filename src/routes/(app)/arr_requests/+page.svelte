<script lang="ts">
  import Icon from "@/lib/Icon.svelte";
  import PageError from "@/lib/PageError.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { baseURL } from "@/lib/util/api";
  import { getOrdinalSuffix, monthsShort, userHasPermission } from "@/lib/util/helpers";
  import { UserPermission, type ManagedUser, type ArrRequestResponse } from "@/types";
  import axios from "axios";

  let allRequests: ArrRequestResponse[];
  let editingUser: ManagedUser | undefined;

  async function getRequests() {
    allRequests = (await axios.get(`/arr/request/`)).data as ArrRequestResponse[];
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
          <div class="request">
            <img src={`${baseURL}/img${r.content?.poster_path}`} alt="" />
            <img src={`${baseURL}/img${r.content?.poster_path}`} alt="" />
            <div>
              <h2 class="norm">
                <span>{r.content.title}</span>
                {#if r.content.release_date}
                  <span>{new Date(r.content.release_date).getFullYear()}</span>
                {/if}
              </h2>
              <p>{r.content.overview}</p>
              <div class="btns">
                <button class="decline">Decline</button>
                <button class="approve">Approve</button>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {:catch err}
      <PageError error={err} pretty="Failed to fetch requests!" />
    {/await}
  </div>
</div>

<style lang="scss">
  .request-container {
    display: flex;
    flex-flow: column;
    gap: 10px;
    margin-top: 20px;
  }

  .request {
    display: flex;
    flex-flow: row;
    position: relative;
    gap: 10px;
    max-height: 333px;
    padding: 10px;
    border-radius: 10px;
    background-color: $img-blend-multiply-bg-col;
    color: white;
    overflow: hidden;

    img {
      width: 150px;
      height: 225px;
      border-radius: 6px;

      &:first-of-type {
        position: absolute;
        left: 0;
        top: 0;
        width: 100%;
        height: auto;
        transform: translateY(-25%);
        filter: blur(4px) grayscale(80%);
        mix-blend-mode: multiply;
        z-index: -2;
      }
    }

    & > div {
      display: flex;
      flex-flow: column;

      & > h2 {
        font-size: 22px;

        & span:last-of-type {
          font-size: 16px;
          font-weight: 400;
          color: rgba(255, 255, 255, 0.7);
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
        white-space: nowrap;
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
