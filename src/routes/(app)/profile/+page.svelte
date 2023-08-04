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

  function toggleTheme(theme: "light" | "dark") {
    if (theme === "dark") {
      document.documentElement.classList.add("theme-dark");
    } else {
      document.documentElement.classList.remove("theme-dark");
    }
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

    <div class="settings">
      <h3 class="norm">Settings</h3>

      <div class="theme">
        <button class="plain" id="light" on:click={() => toggleTheme("light")} />
        <button class="plain" id="dark" on:click={() => toggleTheme("dark")} />
      </div>
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
      margin: 0 30px;

      & > div:not(:first-of-type) {
        margin-top: 30px;
      }
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
      background-color: $accent-color;
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

  .settings {
    display: flex;
    flex-flow: column;
    width: 100%;

    .theme {
      display: flex;
      gap: 10px;
      margin: 20px;
      width: 100%;

      & > button {
        width: 50%;
        height: 80px;
        border-radius: 10px;
        outline: 3px solid $text-color;

        &#light {
          background-color: gray;
        }

        &#dark {
          background-color: black;
        }

        &.selected {
          border: 3px solid gray;
        }
      }
    }
  }
</style>
