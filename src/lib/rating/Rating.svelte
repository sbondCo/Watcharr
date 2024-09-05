<script lang="ts">
  import { userSettings } from "@/store";
  import { RatingSystem } from "@/types";
  import StarRating from "./StarRating.svelte";
  import ThumbRating from "./ThumbRating.svelte";

  $: settings = $userSettings;

  export let rating: number | undefined;
  export let onChange: (newRating: number) => void;
</script>

<!-- TODO make sure stars work good on mobile, should be able to hold down and adjust like a slider -->
<!-- TODO 0.1 increments need to work better - eg: 5.8 looks the same as 6.0 -->
<!-- TODO 1.0 increments need to work better - eg: rating of 8 shows like 1px of yellow onto the 9th star.. JARRING -->

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
    /* TODO responsivise for smol screenz */
    width: 377px;
  }
</style>
