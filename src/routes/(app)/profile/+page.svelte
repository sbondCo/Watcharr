<script lang="ts">
  import { goto } from "$app/navigation";
  import Checkbox from "@/lib/Checkbox.svelte";
  import Error from "@/lib/Error.svelte";
  import Icon from "@/lib/Icon.svelte";
  import Spinner from "@/lib/Spinner.svelte";
  import Setting from "@/lib/settings/Setting.svelte";
  import Stat from "@/lib/stats/Stat.svelte";
  import Stats from "@/lib/stats/Stats.svelte";
  import { updateUserSetting } from "@/lib/util/api";
  import { getOrdinalSuffix, monthsShort, toggleTheme } from "@/lib/util/helpers";
  import { appTheme, userInfo, userSettings } from "@/store";
  import type { Profile } from "@/types";
  import axios from "axios";
  import { onMount } from "svelte";

  $: user = $userInfo;
  $: settings = $userSettings;
  $: selectedTheme = $appTheme;

  let privateDisabled = false;
  let hideSpoilersDisabled = false;
  let avatarInput: HTMLInputElement;

  async function getProfile() {
    return (await axios.get(`/profile`)).data as Profile;
  }

  function formatDate(d: Date) {
    return `${d.getDate()}${getOrdinalSuffix(d.getDate())} ${
      monthsShort[d.getMonth()]
    } ${d.getFullYear()}`;
  }

  function avatarDropped(ev: Event) {
    console.log(avatarInput.files);
    if (!avatarInput?.files || avatarInput?.files?.length <= 0) {
      console.error("avatarDropped: no file found");
      return;
    }
    axios.postForm(
      "/user/avatar",
      { avatar: avatarInput.files[0] },
      {
        headers: {
          "Content-Type": "multipart/form-data"
        }
      }
    );
  }

  onMount(() => {
    avatarInput?.addEventListener("input", avatarDropped);

    return () => {
      avatarInput?.removeEventListener("input", avatarDropped);
    };
  });
</script>

<div class="content">
  <div class="inner">
    <div class="user-basic-info">
      <div class="img-ctr">
        <Icon i="person" wh="100%" />
        <input bind:this={avatarInput} type="file" title="" accept=".jpg,.png,.gif,.webp" />
      </div>
      <div>
        <h2 title={user?.username}>
          <span style="font-weight: normal; font-variant: all-small-caps;">Hey</span>
          {user?.username}
        </h2>
        <textarea name="" id="" cols="30" rows="1" placeholder="my bio"></textarea>
      </div>
    </div>

    <Stats>
      {#await getProfile()}
        <Spinner />
      {:then profile}
        <Stat name="Joined" value={formatDate(new Date(profile.joined))} />
        <Stat name="Movies Watched" value={profile.moviesWatched} large />
        <Stat name="Shows Watched" value={profile.showsWatched} large />
      {:catch err}
        <Error error={err} pretty="Failed to get stats!" />
      {/await}
    </Stats>

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

  .user-basic-info {
    display: flex;
    gap: 20px;

    .img-ctr {
      width: 80px;
      min-width: 80px;
      height: 80px;
      min-height: 80px;
      object-fit: cover;
      border-radius: 50%;
      position: relative;

      &:hover {
        opacity: 0.8;
      }

      input[type="file"] {
        opacity: 0;
        width: 100%;
        height: 100%;
        position: absolute;
        cursor: pointer;
      }
    }

    & > div {
      display: flex;
      flex-flow: column;
      gap: 5px;
      width: 100%;

      textarea {
        resize: none;
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
