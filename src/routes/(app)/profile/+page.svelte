<script lang="ts">
  import Error from "@/lib/Error.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import { getOrdinalSuffix, monthsShort } from "@/lib/util/helpers";
  import type { Profile } from "@/types";
  import axios from "axios";

  async function getProfile() {
    return (await axios.get(`/profile`)).data as Profile;
  }

  function formatDate(d: Date) {
    return `${d.getDate()}${getOrdinalSuffix(d.getDate())} ${
      monthsShort[d.getMonth()]
    } ${d.getFullYear()}`;
  }
</script>

<div class="content">
  <div class="inner">
    <h2>Hey {localStorage.getItem("username")}</h2>

    <div class="stats">
      {#await getProfile()}
        <Spinner />
      {:then profile}
        <div>
          <span>{formatDate(new Date(profile.joined))}</span>
          <span>Joined</span>
        </div>
        <div>
          <span class="large">{profile.moviesWatched}</span>
          <span>Movies Watched</span>
        </div>
        <div>
          <span class="large">{profile.showsWatched}</span>
          <span>Shows Watched</span>
        </div>
      {:catch err}
        <Error error={err} pretty="Failed to get stats!" />
      {/await}
    </div>
  </div>
</div>

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;

    .inner {
      min-width: 400px;
      max-width: 400px;
    }
  }

  .stats {
    display: flex;
    flex-flow: row;
    gap: 12px;
    margin-top: 15px;

    > div {
      display: flex;
      flex-flow: column;
      padding: 20px 15px;
      background-color: rgba(128, 128, 128, 0.226);
      border-radius: 8px;

      > span:first-child {
        font-weight: bold;
        font-size: 20px;

        &.large {
          font-size: 32px;
        }
      }

      > span:last-child {
        margin-top: auto;
      }
    }
  }
</style>
