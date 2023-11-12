<script lang="ts">
  import { goto } from "$app/navigation";
  import Checkbox from "@/lib/Checkbox.svelte";
  import Error from "@/lib/Error.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import Setting from "@/lib/settings/Setting.svelte";
  import { updateUserSetting } from "@/lib/util/api";
  import { getOrdinalSuffix, monthsShort, toggleTheme } from "@/lib/util/helpers";
  import { appTheme, userInfo, userSettings } from "@/store";
  import type { Profile } from "@/types";
  import axios from "axios";

  $: user = $userInfo;
  $: settings = $userSettings;
  $: selectedTheme = $appTheme;

  let privateDisabled = false;
  let hideSpoilersDisabled = false;

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
    <h2 title={user?.username}>Hey {user?.username}</h2>

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
        <h4 class="norm">Theme</h4>
        <div class="row">
          <button
            class={`plain${selectedTheme === "light" ? " selected" : ""}`}
            id="light"
            on:click={() => toggleTheme("light")}
          >
            light
          </button>
          <button
            class={`plain${selectedTheme === "dark" ? " selected" : ""}`}
            id="dark"
            on:click={() => toggleTheme("dark")}
          >
            dark
          </button>
        </div>
      </div>

      <Setting title="Private" desc="Hide your profile from others?" row>
        <Checkbox
          name="private"
          disabled={privateDisabled}
          value={settings?.private}
          toggled={(on) => {
            privateDisabled = true;
            updateUserSetting("private", on, () => {
              privateDisabled = false;
            });
          }}
        />
      </Setting>

      <Setting title="Hide Spoilers" desc="Do you want to hide episode info?" row>
        <Checkbox
          name="hideSpoilers"
          disabled={hideSpoilersDisabled}
          value={settings?.hideSpoilers}
          toggled={(on) => {
            hideSpoilersDisabled = true;
            updateUserSetting("hideSpoilers", on, () => {
              hideSpoilersDisabled = false;
            });
          }}
        />
      </Setting>

      <div class="row btns">
        <button on:click={() => goto("/import")}>Import</button>
      </div>
    </div>
  </div>
</div>

<style lang="scss">
  .content {
    display: flex;
    width: 100%;
    justify-content: center;
    padding: 0 30px 0 30px;

    .inner {
      min-width: 400px;
      max-width: 400px;
      overflow: hidden;

      h2 {
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
      }

      & > div:not(:first-of-type) {
        margin-top: 30px;
      }

      @media screen and (max-width: 440px) {
        width: 100%;
        min-width: unset;
      }
    }
  }

  .stats {
    display: flex;
    flex-flow: row;
    gap: 12px;
    margin-top: 15px;

    @media screen and (max-width: 440px) {
      flex-wrap: wrap;
    }

    > div {
      display: flex;
      flex-flow: column;
      flex-grow: 1;
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
    gap: 20px;
    width: 100%;

    h3 {
      font-variant: small-caps;
    }

    & > div {
      margin: 0 15px;
    }

    div {
      &.row {
        display: flex;
        flex-flow: row;
        gap: 10px;
        align-items: center;

        &.btns button {
          width: min-content;
        }
      }
    }

    .theme {
      display: flex;
      flex-flow: column;
      gap: 10px;

      & button {
        width: 50%;
        height: 80px;
        border-radius: 10px;
        outline: 3px solid;
        font-size: 20px;
        text-transform: uppercase;
        font-family: "Rampart One";
        color: transparent;
        transition: all 200ms ease-in;

        &#light {
          background-color: white;
          outline-color: $accent-color;
          &:hover {
            color: black;
            -webkit-text-stroke: 0.5px black;
          }
        }

        &#dark {
          background-color: black;
          outline-color: white;
          &:hover {
            color: white;
            -webkit-text-stroke: 0.5px white;
          }
        }

        &.selected {
          outline-color: gold !important;
        }
      }
    }
  }
</style>
