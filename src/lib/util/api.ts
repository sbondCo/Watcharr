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
	type PlayedAddRequest,
	type ActivityUpdateRequest,
	type WatchedAddedToContent,
} from "@/types";
import axios from "axios";
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

interface UpdateWatchedOptions {
	/**
	 * TMDB ID.
	 */
	contentId: number;
	contentType: MediaType;
	status?: WatchedStatus;
	rating?: number;
	thoughts?: string;
	pinned?: boolean;
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
	const resp = await axios.put<WatchedUpdateResponse>(
		`/watched/${wEntry.id}`,
		obj,
	);
	if (status) wEntry.status = status;
	if (rating) wEntry.rating = rating;
	if (typeof thoughts !== "undefined") wEntry.thoughts = thoughts;
	if (typeof pinned !== "undefined") wEntry.pinned = pinned;
	if (resp?.data?.newActivity && resp?.data?.newActivity?.id) {
		if (wEntry.activity?.length > 0) {
			wEntry.activity.push(resp.data.newActivity);
		} else {
			wEntry.activity = [resp.data.newActivity];
		}
		// We want to update the updatedAt field too (so
		// change is reflected when filtering modified at)
		// We can piggy back from this data for now.
		wEntry.updatedAt = resp.data.newActivity.createdAt;
	}
}

/**
 * Add or update watched show/movie.
 * @param wEntry The watched entry (movie or tv only) we are updating.
 * @param opts Update options.
 * @returns Always returns Watched obj unless failed to add.
 * If updating fails, existing Watched obj will always return.
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
				);
				notify({ id: nid, text: `Saved!`, type: "success" });
			} catch (err) {
				console.error("updateWatched: Failed to update!", err);
				notify({ id: nid, text: `Saving Failed!`, type: "error" });
			}
			// We are updating, so a wEntry exists here.
			// So we will always return the existing entry,
			// regardless of if we fail above.
			return wEntry;
		}
		try {
			// Add new watched item
			notify({ id: nid, text: `Adding`, type: "loading" });
			const resp = await axios.post<Watched>("/watched", {
				contentId: opts.contentId,
				contentType: opts.contentType,
				rating: opts.rating,
				status: opts.status,
			} as WatchedAddRequest);
			console.log("Added watched:", resp.data);
			notify({ id: nid, text: `Added!`, type: "success" });
			return resp.data;
		} catch (err) {
			console.error("updateWatched: Failed to add!", err);
			notify({ id: nid, text: `Adding Failed!`, type: "error" });
			// Watched entry not added so returning undefined is fine,
			// that will be the current value everywhere anyways.
			return undefined;
		}
	} catch (err) {
		console.error("updateWatched: Failed!", err);
		notify({ id: nid, text: `Failed!`, type: "error" });
		return wEntry;
	}
}

/**
 * Delete an item from watched list.
 * @param id Watched Entry ID
 * @returns Deleted?
 */
export async function removeWatched(id: number): Promise<boolean> {
	console.log("removeWatched: Removing:", id);
	const nid = notify({ text: `Removing`, type: "loading" });
	try {
		const resp = await axios.delete(`/watched/${id}`);
		console.log("removeWatched: Removed resp:", resp.data);
		notify({ id: nid, text: "Removed!", type: "success" });
		return true;
	} catch (err) {
		console.error("removeWatched: Failed!", err);
		notify({ id: nid, text: "Failed To Remove!", type: "error" });
	}
	return false;
}

export async function updatePlayed(
	igdbId: number,
	status?: WatchedStatus,
	rating?: number,
	thoughts?: string,
	pinned?: boolean,
): Promise<boolean> {
	// If item is already in watched store, run update request instead
	const wEntry = store.watchedList.find((w) => w.game?.igdbId === igdbId);
	if (wEntry?.id) {
		return await _updateWatched(wEntry, status, rating, thoughts, pinned);
	}
	// Add new played item
	const nid = notify({ text: `Adding`, type: "loading" });
	return await axios
		.post("/game/played", {
			igdbId,
			rating,
			status,
		} as PlayedAddRequest)
		.then((resp) => {
			console.log("Added watched(played) game:", resp.data);
			store.watchedList.push(resp.data as Watched);
			notify({ id: nid, text: `Added!`, type: "success" });
			return true;
		})
		.catch((err) => {
			console.error(err);
			notify({ id: nid, text: "Failed To Add!", type: "error" });
			return false;
		});
}

export function updateActivity(
	watchedId: number,
	activityId: number,
	date: Date,
) {
	const nid = notify({ text: "Updating", type: "loading" });
	console.debug("updateActivity called", watchedId, activityId, date);
	try {
		axios
			.put("/activity/" + activityId, {
				customDate: date.toISOString(),
			} as ActivityUpdateRequest)
			.then((resp) => {
				console.log("Updated activity timestamp:", resp.status);
				const activity = store.watchedList
					.find((w) => w.id === watchedId)
					?.activity.find((a) => a.id === activityId);
				if (activity) {
					activity.customDate = date.toISOString();
				}
				notify({ id: nid, text: "Updated!", type: "success" });
			})
			.catch((err) => {
				console.error(err);
				notify({ id: nid, text: "Failed to Update!", type: "error" });
			});
	} catch (err) {
		console.error("updateActivity failed!", err);
		notify({ id: nid, text: "Failed!", type: "error" });
	}
}

export function removeActivity(watchedId: number, activityId: number) {
	const nid = notify({ text: "Deleting", type: "loading" });
	axios
		.delete("/activity/" + activityId)
		.then((resp) => {
			const wListItem = store.watchedList.find((w) => w.id === watchedId);
			if (wListItem) {
				wListItem.activity = wListItem.activity.filter(
					(i) => i.id !== activityId,
				);
			}
			notify({ id: nid, text: "Deleted!", type: "success" });
		})
		.catch((err) => {
			console.error(err);
			notify({ id: nid, text: "Failed to Delete!", type: "error" });
		});
}

export async function contentExistsOnJellyfin(
	type: MediaType,
	name: string,
	tmdbId: number,
): Promise<JellyfinFoundContent | undefined> {
	try {
		if (Number(store.userInfo?.type) == UserType.Jellyfin) {
			const resp = await axios.get(`/jellyfin/${type}/${name}/${tmdbId}`);
			console.log("contentExistsOnJellyfin response:", resp.data);
			return resp.data as JellyfinFoundContent;
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
	axios
		.post("/user/update", { [name]: value })
		.then((r) => {
			if (r.status === 200) {
				if (store.userSettings) store.userSettings[name] = value;
				notify({ id: nid, type: "success", text: "Updated" });
				if (typeof done !== "undefined") done();
			}
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
	axios
		.post("/auth/change_password", { oldPassword, newPassword })
		.then((r) => {
			if (r.status === 200) {
				notify({ id: nid, type: "success", text: "Password Changed" });
				if (typeof done !== "undefined") done();
			}
		})
		.catch((err) => {
			const errMsg = err?.response?.data?.error
				? err.response.data.error
				: "Couldn't Change Password";
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
		const f = await axios.get("/features");
		if (f?.data) {
			store.serverFeatures = f.data;
		}
	} catch (err) {
		console.error("getServerFeatures failed!", err);
	}
}

export async function followUser(id: number) {
	const nid = notify({ text: `Following`, type: "loading" });
	axios
		.post(`/follow/${id}`)
		.then((resp) => {
			console.log("Followed:", resp.data);
			store.follows.push(resp.data as Follow);
			notify({ id: nid, text: `Followed!`, type: "success" });
		})
		.catch((err) => {
			console.error(err);
			notify({ id: nid, text: "Failed To Follow!", type: "error" });
		});
}

export async function unfollowUser(id: number) {
	const nid = notify({ text: `Unfollowing`, type: "loading" });
	axios
		.delete(`/follow/${id}`)
		.then((resp) => {
			console.log("Unfollowed:", resp.data);
			store.follows = store.follows.filter((fo) => fo.followedUser.id != id);
			notify({ id: nid, text: `Unfollowed!`, type: "success" });
		})
		.catch((err) => {
			console.error(err);
			notify({ id: nid, text: "Failed To Unfollow!", type: "error" });
		});
}

/**
 * For use with routes that don't require authentication (eg login/register)
 */
export const noAuthAxios = axios.create({
	baseURL: baseURL,
});
