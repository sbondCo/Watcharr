<script lang="ts">
  import { userSettings } from "@/store";
  import Setting from "../settings/Setting.svelte";
  import { RatingSystem } from "@/types";

  $: settings = $userSettings;

  function update(v: RatingSystem) {
    if (!settings) {
      console.error("No settings.");
      return;
    }
    settings.ratingSystem = v;
    // userSettings.update((u) => u);
  }

  function updateStep(v: number) {
    if (!settings) {
      console.error("No settings.");
      return;
    }
    settings.ratingStep = v;
  }
</script>

<Setting title="Rating System" desc="How would you like to rate content?">
  <div class="rat-wrap">
    <button
      class={["plain", settings?.ratingSystem === RatingSystem.OutOf5 ? "active" : ""].join(" ")}
      on:click={() => update(RatingSystem.OutOf5)}
    >
      0-5
    </button>
    <button
      class={[
        "plain",
        settings?.ratingSystem === RatingSystem.OutOf10 || !settings?.ratingSystem ? "active" : ""
      ].join(" ")}
      on:click={() => update(RatingSystem.OutOf10)}
    >
      0-10
    </button>
    <button
      class={["plain", settings?.ratingSystem === RatingSystem.OutOf100 ? "active" : ""].join(" ")}
      on:click={() => update(RatingSystem.OutOf100)}
    >
      0-100
    </button>
    <button
      class={["plain", settings?.ratingSystem === RatingSystem.Thumbs ? "active" : ""].join(" ")}
      on:click={() => update(RatingSystem.Thumbs)}
    >
      Thumbs
    </button>
  </div>
</Setting>

{#if settings?.ratingSystem === RatingSystem.OutOf10 || settings?.ratingSystem === RatingSystem.OutOf5 || !settings?.ratingSystem}
  <Setting title="Rating Step" desc="How would you like to increment through the stars?">
    <div class="rat-wrap">
      <button
        class={["plain", settings?.ratingStep === 0.1 ? "active" : ""].join(" ")}
        on:click={() => updateStep(0.1)}
      >
        0.1
      </button>
      <button
        class={["plain", settings?.ratingStep === 0.5 ? "active" : ""].join(" ")}
        on:click={() => updateStep(0.5)}
      >
        0.5
      </button>
      <button
        class={["plain", !settings?.ratingStep || settings?.ratingStep === 1 ? "active" : ""].join(
          " "
        )}
        on:click={() => updateStep(1)}
      >
        1
      </button>
    </div>
  </Setting>
{/if}

<style lang="scss">
  .rat-wrap {
    display: flex;
    flex-flow: row;
    /* gap: 5px; */
    border-radius: 10px;
    overflow: hidden;

    button {
      display: flex;
      flex-flow: row;
      align-items: center;
      justify-content: center;
      gap: 3px;
      width: 100%;
      padding: 15px 8px;
      color: $text-color;
      background-color: $accent-color;
      /* font-family: "Shrikhand", sans-serif; */
      font-size: 16px;
      transition:
        color 150ms ease-in-out,
        background-color 150ms ease-in-out;

      &:hover,
      &.active {
        color: $bg-color;
        background-color: $accent-color-hover;
        font-weight: bold;
      }

      &:not(:last-of-type) {
        /* border-right: 1px solid $placeholder-color; */
      }

      :global(svg) {
        width: 24px;
      }
    }
  }
</style>
