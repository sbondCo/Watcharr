import type { ContentType, Watched, WatchedStatus } from "@/types";
import { getLatestWatchedInTv } from "../util/helpers";

export type PosterExtraDetails = {
	rating: number | undefined;
	status: WatchedStatus | undefined;
	dateAdded?: string;
	dateModified?: string;
	/**
	 * Only for shows.
	 */
	lastWatched?: string;
};

export function buildExtraDetails(
	t: ContentType,
	w: Watched,
): PosterExtraDetails {
	const obj = {
		rating: w.rating,
		status: w.status,
		dateAdded: w.createdAt,
		dateModified: w.updatedAt,
	} as PosterExtraDetails;
	if (t === "tv") {
		obj.lastWatched = getLatestWatchedInTv(w.watchedSeasons, w.watchedEpisodes);
	}
	return obj;
}
