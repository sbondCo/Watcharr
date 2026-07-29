import { store } from "@/store.svelte";
import { RatingStep, RatingSystem } from "@/types";

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
	if (
		!store.userSettings ||
		(!store.userSettings.ratingSystem && !store.userSettings.ratingStep)
	) {
		return Math.round(r);
	}
	if (store.userSettings.ratingSystem === RatingSystem.OutOf100) {
		return r * 10;
	}
	if (store.userSettings.ratingSystem === RatingSystem.OutOf5) {
		if (store.userSettings.ratingStep === RatingStep.Point5) {
			return Math.ceil((r / 2) * 2) / 2;
		}
		if (store.userSettings.ratingStep === RatingStep.Point1) {
			return Math.round((r / 2) * 10) / 10;
		}
		return Math.round(r / 2);
	}
	if (store.userSettings.ratingSystem === RatingSystem.OutOf10) {
		if (store.userSettings.ratingStep === RatingStep.Point5) {
			return Math.ceil(r * 2) / 2;
		}
		if (store.userSettings.ratingStep === RatingStep.Point1) {
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
