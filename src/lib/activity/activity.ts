import type { Activity, Watched } from "@/types";

/**
 * Activity removed hook for all content pages to use
 * since the code will be the same.
 */
export function activityRemovedHook(w?: Watched, a?: Activity) {
	if (!w || !w.activity || w.activity?.length == 0 || !a) {
		console.warn("activityRemovedHook: Watched or activity not defined!", w, a);
		return;
	}
	activityRemovedUpdateWatchedPlays(w, a.countAsPlay);
	w.activity = w.activity.filter((ac) => ac.id !== a.id);
}

/**
 * Update local watched state with new `plays` when removing activity.
 */
function activityRemovedUpdateWatchedPlays(
	w: Watched,
	removedCountedAsPlay: boolean,
) {
	// Only if the removed actitivty counted as a play
	// will we update the state.
	if (!removedCountedAsPlay) {
		return;
	}
	if (w.plays && w.plays > 0) {
		w.plays--;
	}
}
