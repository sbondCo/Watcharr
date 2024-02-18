import { appTheme } from "@/store";
import {
  UserPermission,
  type Icon,
  type MediaType,
  type TMDBContentCreditsCrew,
  type Theme,
  type TokenClaims,
  type Watched,
  type WatchedStatus
} from "@/types";

export const watchedStatuses: {
  [key in WatchedStatus]: Icon;
} = {
  PLANNED: "calendar",
  WATCHING: "clock",
  FINISHED: "check",
  HOLD: "pause",
  DROPPED: "thumb-down"
};

export const months = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December"
];

export const monthsShort = [
  "Jan",
  "Feb",
  "March",
  "Apr",
  "May",
  "June",
  "July",
  "Aug",
  "Sept",
  "Oct",
  "Nov",
  "Dec"
];

export function isTouch() {
  return "ontouchstart" in window;
}

// Not passing wList from #each loop caused it not to have reactivity.
// Passing it through must allow it to recognize it as a dependency?
export function getWatchedDependedProps(wid: number, wtype: MediaType, list: Watched[]) {
  const wel = list.find((wl) => wl.content?.tmdbId === wid && wl.content?.type === wtype);
  if (!wel) return {};
  console.log(wid, wtype, wel?.content?.title, wel?.status, wel?.rating);
  return {
    id: wel.id,
    status: wel.status,
    rating: wel.rating
  };
}

export function getPlayedDependedProps(wid: number, list: Watched[]) {
  const wel = list.find((wl) => wl.game?.igdbId === wid);
  if (!wel) return {};
  return {
    id: wel.id,
    status: wel.status,
    rating: wel.rating
  };
}

/**
 * Turn a watched status into understandable text
 * depending on if the status is for a game or not.
 */
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
  if (isForGame) {
    if (s === "FINISHED") {
      return "played";
    } else if (s === "WATCHING") {
      return "playing";
    }
  }
  if (s === "HOLD") {
    return "on hold";
  }
  return s?.toLowerCase();
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
  c: string
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
    (c) => c.job === "Director" || c.job === "Writer" || c.job === "Characters" || c.job === "Story"
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
  }
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

/**
 * Get ordinal suffix to use from day number (1`st`, 2`nd`, 3`rd`).
 * @param i Day number.
 */
export function getOrdinalSuffix(i: number) {
  const j = i % 10,
    k = i % 100;
  if (j == 1 && k != 11) {
    return "st";
  }
  if (j == 2 && k != 12) {
    return "nd";
  }
  if (j == 3 && k != 13) {
    return "rd";
  }
  return "th";
}

/**
 * Toggle site wide theme.
 * @param theme The theme to switch to.
 */
export function toggleTheme(theme: Theme) {
  if (theme === "dark") {
    document.documentElement.classList.add("theme-dark");
    appTheme.update((t) => (t = "dark"));
  } else {
    document.documentElement.classList.remove("theme-dark");
    appTheme.update((t) => (t = "light"));
  }
}

export function parseTokenPayload(): TokenClaims | undefined {
  try {
    const token = localStorage.getItem("token");
    if (!token) return;
    return JSON.parse(atob(token.split(".")[1])) as TokenClaims;
  } catch (err) {
    return;
  }
}

export async function sleep(ms: number) {
  return new Promise<void>((r) =>
    setTimeout(() => {
      r();
    }, ms)
  );
}

export function userHasPermission(perms: UserPermission, reqPerm: UserPermission): boolean {
  // Admins have permission for everything.
  if (perms & UserPermission.PERM_ADMIN) {
    return true;
  }
  return (perms & reqPerm) == reqPerm;
}

/**
 * Takes season and episode number to make a quickly understandable
 * string in this format S01E01.
 */
export function seasonAndEpToReadable(season: number, episode: number) {
  return `S${season ? String(season)?.padStart(2, "0") : "(unknown)"}E${episode ? String(episode)?.padStart(2, "0") : "(unknown)"}`;
}
