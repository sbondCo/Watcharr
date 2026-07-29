import { store } from "@/store.svelte";
import {
	UserType,
	type JellyfinFoundContent,
	type MediaType,
	type Watched,
	type WatchedAddRequest,
	type WatchedStatus,
	type WatchedUpdateRequest,
	type WatchedUpdateResponse,
	type UserSettings,
	type Follow,
	type ActivityUpdateRequest,
	type Activity,
	type SupportedMedia,
	type ServerFeatures,
} from "@/types";
import { Reqer, ReqerError } from "./fetch";
import { notify, unNotify } from "./notify";
import { browser } from "$app/environment";
const { MODE } = import.meta.env;

export const baseURL =
	MODE === "development"
		? browser
			? `${location.protocol}//${location.hostname}:3080/api`
			: "http://127.0.0.1:3080/api"
		: "/api";
console.log("api: baseURL constructed:", baseURL);

export const req = new Reqer(baseURL, true);
export const noAuthReq = new Reqer(baseURL, false);

/**
 * Options for our internal updateWatched func.
 */
export interface UpdateWatchedOptions extends Omit<
	WatchedUpdateRequest,
	"removeThoughts"
> {
	/**
	 * TMDB ID.
	 */
	contentId: number;
	contentType: SupportedMedia;
}

/**
 * Updates watched item with new status, rating or thoughts.
 * @param wEntry The watched entry to update. Updates properties in this object.
 */
async function _updateWatched(
	wEntry: Watched,
	status?: WatchedStatus,
	rating?: number,
	thoughts?: string,
	pinned?: boolean,
	letCountAsPlay?: boolean,
) {
	if (
		!status &&
		!rating &&
		typeof thoughts === "undefined" &&
		typeof pinned === "undefined"
	) {
		console.warn(
			"_updateWatched: Nothing was provided, so nothing can be updated!!!!",
		);
		throw new Error("no updated values provided");
	}
	const obj = {} as WatchedUpdateRequest;
	if (status) obj.status = status;
	if (rating) obj.rating = rating;
	if (typeof thoughts !== "undefined") obj.thoughts = thoughts;
	if (thoughts === "") obj.removeThoughts = true;
	if (typeof pinned !== "undefined") obj.pinned = pinned;
	if (typeof letCountAsPlay !== "undefined") {
		obj.letCountAsPlay = letCountAsPlay;
	}
	const resp = await req.put<WatchedUpdateResponse>(
		`/watched/${wEntry.id}`,
		obj,
	);
	if (status) wEntry.status = status;
	if (rating) wEntry.rating = rating;
	if (typeof thoughts !== "undefined") wEntry.thoughts = thoughts;
	if (typeof pinned !== "undefined") wEntry.pinned = pinned;
	if (resp?.newActivity && resp?.newActivity?.id) {
		if (wEntry.activity && wEntry.activity.length > 0) {
			wEntry.activity.push(resp.newActivity);
		} else {
			wEntry.activity = [resp.newActivity];
		}
		// If new activity counts as play, increment plays for local state.
		if (resp.newActivity.countAsPlay) {
			if (wEntry.plays) {
				wEntry.plays++;
			} else {
				wEntry.plays = 1;
			}
		}
	}
}

/**
 * Add or update watched media.
 * @param wEntry The watched entry we are updating.
 * @param opts Update options.
 * @returns Updated watched entry if request succeeded, otherwise will
 * throw error after displaying updating the notification to "failed".
 * We throw so callers can skip reassigning state when not necessary.
 */
export async function updateWatched(
	wEntry: Watched | undefined,
	opts: UpdateWatchedOptions,
): Promise<Watched | undefined> {
	const nid = notify({ text: `Saving`, type: "loading" });
	try {
		// If exists, run update request instead
		if (wEntry?.id) {
			try {
				await _updateWatched(
					wEntry,
					opts.status,
					opts.rating,
					opts.thoughts,
					opts.pinned,
					opts.letCountAsPlay,
				);
				notify({ id: nid, text: `Saved!`, type: "success" });
			} catch (err) {
				console.error("updateWatched: Failed to update!", err);
				throw err;
			}
			// We are updating, so a wEntry exists here.
			// So we will always return the existing entry.
			return wEntry;
		}

		// Add new watched item
		notify({ id: nid, text: `Adding`, type: "loading" });
		const reqBody: WatchedAddRequest = {
			contentType: opts.contentType,
			status: opts.status,
			rating: opts.rating,
		};
		if (opts.contentType === "movie" || opts.contentType === "tv") {
			reqBody.tmdbId = opts.contentId;
		} else if (opts.contentType === "game") {
			reqBody.igdbId = opts.contentId;
		} else {
			throw "invalid contentType";
		}
		const resp = await req.post<Watched>("/watched", reqBody);
		console.log("Added watched:", resp);
		notify({ id: nid, text: `Added!`, type: "success" });
		return resp;
	} catch (err) {
		console.error("updateWatched: Failed!", err);
		notify({ id: nid, text: `Failed!`, type: "error" });
		throw err;
	}
}

/**
 * Delete an item from watched list.
 * @param id Watched Entry ID
 * @returns Deleted?
 */
export async function removeWatched(id: number): Promise<boolean> {
	console.log("removeWatched: Removing:", id);
	const nid = notify({ text: "Removing", type: "loading" });
	try {
		const resp = await req.delete(`/watched/${id}`);
		console.log("removeWatched: Removed resp:", resp);
		notify({ id: nid, text: "Removed!", type: "success" });
		return true;
	} catch (err) {
		console.error("removeWatched: Failed!", err);
		notify({ id: nid, text: "Failed To Remove!", type: "error" });
	}
	return false;
}

export async function updateActivity(
	activity: Activity,
	date: Date,
): Promise<Activity | undefined> {
	const nid = notify({ text: "Updating", type: "loading" });
	console.debug("updateActivity:", activity, date);
	try {
		const resp = await req.putWhole(`/activity/${activity.id}`, {
			customDate: date.toISOString(),
		} as ActivityUpdateRequest);
		console.log("updateActivity: Response status:", resp.status);
		if (activity) {
			activity.customDate = date.toISOString();
		}
		notify({ id: nid, text: "Updated!", type: "success" });
		return activity;
	} catch (err) {
		console.error("updateActivity failed!", err);
		notify({ id: nid, text: "Failed to Update!", type: "error" });
	}
}

export async function removeActivity(activityId: number): Promise<boolean> {
	const nid = notify({ text: "Deleting", type: "loading" });
	try {
		await req.delete("/activity/" + activityId);
		console.log("removeActivity: Removed:", activityId);
		notify({ id: nid, text: "Deleted!", type: "success" });
		return true;
	} catch (err) {
		console.error("removeActivity: Failed!", err);
		notify({ id: nid, text: "Failed to Delete!", type: "error" });
	}
	return false;
}

export async function contentExistsOnJellyfin(
	type: MediaType,
	name: string,
	tmdbId: number,
): Promise<JellyfinFoundContent | undefined> {
	try {
		if (Number(store.userInfo?.type) == UserType.Jellyfin) {
			const resp = await req.get<JellyfinFoundContent>(
				`/jellyfin/${type}/${name}/${tmdbId}`,
			);
			console.log("contentExistsOnJellyfin response:", resp);
			return resp;
		}
	} catch (err) {
		console.error(err);
		// notify({ text: "Failed To Remove!", type: "error" });
	}
}

export function updateUserSetting<K extends keyof UserSettings>(
	name: K,
	value: UserSettings[K],
	done?: () => void,
) {
	console.log("Updating user setting", name, "to", value);
	if (!store.userSettings) {
		console.log("updateUserSetting: userSettings not set..");
		return;
	}
	const originalValue = store.userSettings[name];
	const nid = notify({ type: "loading", text: "Updating" });
	req
		.post("/user/update", { [name]: value })
		.then(() => {
			if (store.userSettings) store.userSettings[name] = value;
			notify({ id: nid, type: "success", text: "Updated" });
			if (typeof done !== "undefined") done();
		})
		.catch((err) => {
			console.error("Failed to update user setting", err);
			notify({ id: nid, type: "error", text: "Couldn't Update" });
			if (store.userSettings) store.userSettings[name] = originalValue;
			if (typeof done !== "undefined") done();
		});
}

export function changeUserPassword(
	oldPassword: string,
	newPassword: string,
	done?: (errMsg?: string) => void,
) {
	const nid = notify({ type: "loading", text: "Changing Password" });
	req
		.post("/auth/change_password", { oldPassword, newPassword })
		.then(() => {
			notify({ id: nid, type: "success", text: "Password Changed" });
			if (typeof done !== "undefined") done();
		})
		.catch((err) => {
			const errMsg = ReqerError.getMsg(err, "Couldn't Change Password");
			console.error(
				"Change Password Form - Failed to change password on the server",
				err,
			);
			unNotify(nid);
			if (typeof done !== "undefined") done(errMsg);
		});
}

/**
 * Update serverFeatues store with fresh data.
 */
export async function getServerFeatures() {
	try {
		const f = await req.get<ServerFeatures>("/features");
		if (f) {
			store.serverFeatures = f;
		}
	} catch (err) {
		console.error("getServerFeatures failed!", err);
	}
}

export async function followUser(id: number) {
	const nid = notify({ text: `Following`, type: "loading" });
	req
		.post<Follow>(`/follow/${id}`)
		.then((resp) => {
			console.log("Followed:", resp);
			store.follows.push(resp);
			notify({ id: nid, text: `Followed!`, type: "success" });
		})
		.catch((err) => {
			console.error(err);
			notify({ id: nid, text: "Failed To Follow!", type: "error" });
		});
}

export async function unfollowUser(id: number) {
	const nid = notify({ text: `Unfollowing`, type: "loading" });
	req
		.delete(`/follow/${id}`)
		.then((resp) => {
			console.log("Unfollowed:", resp);
			store.follows = store.follows.filter((fo) => fo.followedUser.id != id);
			notify({ id: nid, text: `Unfollowed!`, type: "success" });
		})
		.catch((err) => {
			console.error(err);
			notify({ id: nid, text: "Failed To Unfollow!", type: "error" });
		});
}
