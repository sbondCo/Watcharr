import {
	UserPermission,
	type Icon,
	type TMDBContentCreditsCrew,
	type TokenClaims,
	type WatchedStatus,
} from "@/types";
import { getLanguage, t } from "@/lib/i18n/index.svelte";
import { formatShortDate } from "./date";

export const watchedStatuses: {
	[key in WatchedStatus]: Icon;
} = {
	PLANNED: "calendar",
	WATCHING: "clock",
	FINISHED: "check",
	HOLD: "pause",
	DROPPED: "thumb-down",
};

export function isTouch() {
	return "ontouchstart" in window;
}

/**
 * Turns a WatchedStatus into readable and context aware text.
 * Watched statuses can be used normally for movies/tv, but
 * for games, we want to transform the status to make more sense.
 * ex: 'finished' would become 'played' for games, but remain
 *     unmodified for series/movies.
 *
 * This is only for use when displaying a status in ui for a user
 * to read, should **never** be involved in logic (comparing
 * statuses for example).
 *
 * @param s The watched status.
 * @param isForGame If this status is being displayed for a game or not.
 */
export function toUnderstandableStatus(s: WatchedStatus, isForGame: boolean) {
	return statusLabel(s, isForGame).toLocaleLowerCase(getLanguage());
}

/**
 * Get the translated, title cased name of a WatchedStatus
 * (eg: `On Hold`).
 *
 * Like `toUnderstandableStatus`, only for displaying in the ui.
 *
 * @param s The watched status.
 * @param isForGame If this status is being displayed for a game or not.
 */
export function statusLabel(s: WatchedStatus, isForGame = false): string {
	if (isForGame) {
		if (s === "FINISHED") {
			return t("status.played");
		} else if (s === "WATCHING") {
			return t("status.playing");
		}
	}
	switch (s) {
		case "PLANNED":
			return t("status.planned");
		case "WATCHING":
			return t("status.watching");
		case "FINISHED":
			return t("status.finished");
		case "HOLD":
			return t("status.hold");
		case "DROPPED":
			return t("status.dropped");
		default:
			return s ?? t("common.unknown");
	}
}

/**
 * Add a class to the parent node of a clicked element.
 * @param e Event with currentTarget.
 * @param c Class to add to parent.
 */
export function addClassToParent(
	e: Event & {
		currentTarget: EventTarget & Element;
	},
	c: string,
) {
	(e.currentTarget?.parentNode as HTMLDivElement)?.classList.add(c);
}

/**
 * Gets "main" crew members from list of crew.
 * @param crew Crew
 * @returns Top Crew
 */
export function getTopCrew(crew: TMDBContentCreditsCrew[]) {
	return crew.filter(
		(c) =>
			c.job === "Director" ||
			c.job === "Writer" ||
			c.job === "Characters" ||
			c.job === "Story",
	);
}

/**
 * Calculates what the transform-origin property should be
 * depending on where the scaled (poster) element will be
 * in the viewport to keep it in view.
 * @param e
 */
export function calculateTransformOrigin(
	e: Event & {
		currentTarget: EventTarget & HTMLLIElement;
	},
) {
	const magicNumber = 26;
	const ctr = e.currentTarget.querySelector(".container") as HTMLElement;
	const pb = ctr.getBoundingClientRect();
	const sx = pb.x;
	const sw = pb.width;
	const wb = document.body.getBoundingClientRect();

	if (ctr) {
		ctr.style.transformOrigin = "unset";
		const origins = [];
		// Overflow on right
		if (sx + sw + magicNumber > wb.x + wb.width) {
			origins.push("right");
		}
		// Overflow on left
		if (sx - magicNumber < wb.x) {
			origins.push("left");
		}
		// Overflow on bottom
		const ppb = e.currentTarget.getBoundingClientRect();
		if (ppb.bottom + magicNumber > window.innerHeight) {
			origins.push("bottom");
		}
		ctr.style.transformOrigin = `${origins.join(" ")}`;
	}
}

export function parseTokenPayload(): TokenClaims | undefined {
	try {
		const token = localStorage.getItem("token");
		if (!token) return;
		return JSON.parse(atob(token.split(".")[1])) as TokenClaims;
	} catch (err) {
		console.error("parseTokenPayload: Failed.", err);
		return;
	}
}

export async function sleep(ms: number) {
	return new Promise<void>((r) =>
		setTimeout(() => {
			r();
		}, ms),
	);
}

export function userHasPermission(
	perms: UserPermission,
	reqPerm: UserPermission,
): boolean {
	// Admins have permission for everything.
	if (perms & UserPermission.PERM_ADMIN) {
		return true;
	}
	return (perms & reqPerm) == reqPerm;
}

/**
 * Takes season and episode number to make a quickly understandable
 * string in this format S1E1.
 */
export function seasonAndEpToReadable(
	season: number | undefined,
	episode: number | undefined,
) {
	const unknown = `(${t("common.unknown").toLocaleLowerCase(getLanguage())})`;
	return t("season.seasonAndEpisodeShort", {
		season: typeof season === "number" ? String(season) : unknown,
		episode: typeof episode === "number" ? String(episode) : unknown,
	});
}

export function msToAmountsOfTime(ms: number) {
	const seconds = Math.floor((ms / 1000) % 60);
	const minutes = Math.floor((ms / (1000 * 60)) % 60);
	const hours = Math.floor((ms / (1000 * 60 * 60)) % 24);
	const days = Math.floor(ms / (1000 * 60 * 60 * 24));
	return { days, hours, minutes, seconds };
}

/** I'm no warden of time, this is as relative as it's getting. */
export function toRelativeDate(d: Date): string {
	if (!d) {
		return t("common.unknown");
	}
	const dn = new Date(Date.now());
	if (d.getFullYear() === dn.getFullYear()) {
		if (d.getMonth() === dn.getMonth()) {
			if (d.getDate() === dn.getDate()) {
				return t("date.today");
			}
		}
		return formatShortDate(d, false);
	}
	return formatShortDate(d);
}

/**
 * To relative time (seconds, mins, hours).
 * @param s Seconds.
 */
export function toRelativeTime(s: number) {
	const totalMinutes = Math.floor(s / 60);
	const hours = Math.floor(totalMinutes / 60);
	const mins = totalMinutes % 60;
	const secs = Math.floor(s % 60);
	if (hours > 0) {
		return t("time.hours", { count: hours });
	}
	if (mins > 0) {
		return t("time.minutes", { count: mins });
	}
	if (secs > 0) {
		return t("time.seconds", { count: secs });
	}
	return t("time.now");
}

/**
 * Tests whether the mouse is hovering over `el`.
 */
export function mouseOverEl(el?: HTMLElement) {
	if (!el) {
		return false;
	}
	return el.matches(":hover");
}
