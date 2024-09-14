import { userSettings } from "@/store";
import { RatingStep, RatingSystem } from "@/types";
import { get } from "svelte/store";

/**
 * Used for scaling users 'actual' rating we store in db
 * into one we can show that takes into account their
 * settings for how they want stars displayed.
 * Only used for star ratings, not thumbs.
 */
export function toShowableRating(r?: number) {
  if (!r) {
    return 0;
  }
  const settings = get(userSettings);
  if (!settings || (!settings.ratingSystem && !settings.ratingStep)) {
    return Math.round(r);
  }
  if (settings.ratingSystem === RatingSystem.OutOf100) {
    return r * 10;
  }
  if (settings.ratingSystem === RatingSystem.OutOf5) {
    if (settings.ratingStep === RatingStep.Point5) {
      return Math.ceil((r / 2) * 2) / 2;
    }
    if (settings.ratingStep === RatingStep.Point1) {
      return r / 2;
    }
    return Math.round(r / 2);
  }
  if (settings.ratingSystem === RatingSystem.OutOf10) {
    if (settings.ratingStep === RatingStep.Point5) {
      return Math.ceil(r * 2) / 2;
    }
    if (settings.ratingStep === RatingStep.Point1) {
      return r;
    }
    return Math.round(r);
  }
  return Math.round(r);
}

export function toWhichThumb(r?: number) {
  if (!r) {
    return;
  }
  const rr = Math.round(r);
  if (rr > 0 && rr <= 4) {
    return -1;
  } else if (r >= 4 && r <= 7) {
    return 0;
  } else if (r >= 8) {
    return 1;
  }
}
