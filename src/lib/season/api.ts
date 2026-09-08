import { req } from "../util/api";
import { notify } from "../util/notify";
import type {
	Activity,
	Watched,
	WatchedEpisodeSetResponse,
	WatchedSeasonSetResponse,
	WatchedStatus,
} from "@/types";

interface UpdateWatchedSeasonOptions {
	status?: WatchedStatus;
	rating?: number;
}

interface UpdateWatchedEpisodeOptions {
	status?: WatchedStatus;
	rating?: number;
}

export async function updateWatchedSeason(
	watchedItem: Watched,
	seasonNumber: number,
	opts: UpdateWatchedSeasonOptions,
) {
	if (!watchedItem) {
		console.error("updateWatchedSeason: No watched item.");
		return;
	}
	const nid = notify({ text: `Saving`, type: "loading" });
	try {
		const r = await req.post<WatchedSeasonSetResponse>(`/watched/season`, {
			watchedId: watchedItem.id,
			seasonNumber: seasonNumber,
			status: opts.status,
			rating: opts.rating,
		});

		if (!watchedItem.watchedSeasons) {
			watchedItem.watchedSeasons = [r.watchedSeason];
		} else {
			if (r.update) {
				const idx = watchedItem.watchedSeasons.findIndex(
					(e) => e.id == r.watchedSeason.id,
				);
				if (idx !== -1) {
					watchedItem.watchedSeasons[idx] = r.watchedSeason;
				} else {
					console.error(
						"updateWatchedSeason: Failed to find index for season to update.",
						"r.watchedSeason:",
						r.watchedSeason,
					);
					notify({
						id: nid,
						text: "Updated but failed to reflect in UI, please reload the page.",
						type: "error",
					});
				}
			} else {
				watchedItem.watchedSeasons.push(r.watchedSeason);
			}
		}

		if (!watchedItem.activity) {
			watchedItem.activity = [];
		}
		if (r.addedActivities && r.addedActivities.length > 0) {
			watchedItem.activity.push(...r.addedActivities);
		}

		notify({ id: nid, text: `Saved!`, type: "success" });
	} catch (err) {
		console.error("updateWatchedSeason: Failed!", err);
		notify({ id: nid, text: "Failed To Update!", type: "error" });
	}
}

export async function removeWatchedSeason(watchedItem: Watched, id: number) {
	const nid = notify({ text: `Removing`, type: "loading" });
	try {
		const r = await req.delete<Activity>(`/watched/season/${id}`);
		watchedItem.watchedSeasons = watchedItem.watchedSeasons?.filter(
			(s) => s.id !== id,
		);
		if (r) {
			if (watchedItem.activity && watchedItem.activity?.length > 0) {
				watchedItem.activity.push(r);
			} else {
				watchedItem.activity = [r];
			}
		}
		notify({ id: nid, text: `Removed!`, type: "success" });
	} catch (err) {
		console.error("removeWatchedSeason: Failed!", err);
		notify({ id: nid, text: "Failed To Remove!", type: "error" });
	}
}

export async function updateWatchedEpisode(
	watchedItem: Watched,
	seasonNumber: number,
	episodeNumber: number,
	opts: UpdateWatchedEpisodeOptions,
) {
	if (!watchedItem) {
		console.error("SeasonListEpisode: updateWatchedEpisode: No watched item.");
		return;
	}
	const nid = notify({ text: `Saving`, type: "loading" });
	try {
		const r = await req.post<WatchedEpisodeSetResponse>(`/watched/episode`, {
			watchedId: watchedItem.id,
			seasonNumber,
			episodeNumber,
			status: opts.status,
			rating: opts.rating,
		});

		if (!watchedItem.watchedEpisodes) {
			watchedItem.watchedEpisodes = [r.watchedEpisode];
		} else {
			if (r.update) {
				const idx = watchedItem.watchedEpisodes.findIndex(
					(e) => e.id == r.watchedEpisode.id,
				);
				if (idx !== -1) {
					watchedItem.watchedEpisodes[idx] = r.watchedEpisode;
				} else {
					console.error(
						"updateWatchedEpisode: Failed to find index for episode to update.",
						"r.watchedEpisode:",
						r.watchedEpisode,
					);
					notify({
						id: nid,
						text: "Updated but failed to reflect in UI, please reload the page.",
						type: "error",
					});
				}
			} else {
				watchedItem.watchedEpisodes.push(r.watchedEpisode);
			}
		}

		if (!watchedItem.activity) {
			watchedItem.activity = [];
		}
		if (r.addedActivities && r.addedActivities.length > 0) {
			watchedItem.activity.push(...r.addedActivities);
		}

		try {
			const epHookResp = r?.episodeStatusChangedHookResponse;
			if (epHookResp && Object.keys(epHookResp).length > 0) {
				if (epHookResp.errors && epHookResp.errors.length > 0) {
					console.error(
						"episodeStatusChangedHookResponse contained errors! All possible automations may not have been completed.",
						epHookResp.errors,
					);
					notify({
						type: "error",
						text: "Some automations have failed, check console for more info.",
					});
				}
				if (
					epHookResp.addedActivities &&
					epHookResp.addedActivities.length > 0
				) {
					watchedItem.activity.push(...epHookResp.addedActivities);
				}
				if (epHookResp.watchedSeason) {
					if (!watchedItem.watchedSeasons) {
						watchedItem.watchedSeasons = [epHookResp.watchedSeason];
					} else {
						const watchedSeasonIdx = watchedItem.watchedSeasons.findIndex(
							(s) => s.id === epHookResp.watchedSeason?.id,
						);
						if (watchedSeasonIdx === -1) {
							watchedItem.watchedSeasons.push(epHookResp.watchedSeason);
						} else {
							watchedItem.watchedSeasons[watchedSeasonIdx] =
								epHookResp.watchedSeason;
						}
					}
				}
				if (epHookResp.newShowStatus) {
					watchedItem.status = epHookResp.newShowStatus;
				}
			}
		} catch (err) {
			console.error("Failed to process episodeStatusChangedHookResponse", err);
			notify({
				type: "error",
				text: "Failed to process automation response, check console for more info.",
			});
		}
		notify({ id: nid, text: `Saved!`, type: "success" });
	} catch (err) {
		console.error("updateWatchedEpisode: Failed!", err);
		notify({ id: nid, text: "Failed To Update!", type: "error" });
	}
}

export async function removeWatchedEpisode(watchedItem: Watched, id: number) {
	const nid = notify({ text: `Removing`, type: "loading" });
	try {
		const r = await req.delete<Activity>(`/watched/episode/${id}`);
		watchedItem.watchedEpisodes = watchedItem.watchedEpisodes?.filter(
			(s) => s.id !== id,
		);
		if (r.data) {
			if (watchedItem.activity && watchedItem.activity?.length > 0) {
				watchedItem.activity.push(r);
			} else {
				watchedItem.activity = [r];
			}
		}
		notify({ id: nid, text: `Removed!`, type: "success" });
	} catch (err) {
		console.error("removeWatchedEpisode: Failed!", err);
		notify({ id: nid, text: "Failed To Remove!", type: "error" });
	}
}
