<script lang="ts">
  import { userSettings } from "@/store";
  import { RatingSystem } from "@/types";
  import StarRating from "./StarRating.svelte";
  import ThumbRating from "./ThumbRating.svelte";

  $: settings = $userSettings;

  export let rating: number | undefined;
  export let onChange: (newRating: number) => void;
</script>

DO THE TODO COMMENTS<br />

{settings?.ratingSystem}<br />
RATING: {rating}<br />

<div class="wrap">
  {#if settings?.ratingSystem === RatingSystem.Thumbs}
    <ThumbRating {rating} {onChange} />
  {:else}
    <!-- All other systems work with the stars -->
    <StarRating {rating} {onChange} />
  {/if}
</div>

<style lang="scss">
  .wrap {
    width: 377px;

    @media screen and (max-width: 420px) {
      width: 100%;
    }
  }
</style>
