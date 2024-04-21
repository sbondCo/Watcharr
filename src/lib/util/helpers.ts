import { appTheme } from "@/store";
import {
  UserPermission,
  type Icon,
  type MediaType,
  type TMDBContentCreditsCrew,
  type Theme,
  type TokenClaims,
  type Watched,
  type WatchedStatus,
  type WatchedEpisode,
  type WatchedSeason
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
    rating: wel.rating,
    extraDetails: {
      dateAdded: wel.createdAt,
      dateModified: wel.updatedAt,
      lastWatched: getLatestWatchedInTv(wel.watchedSeasons, wel.watchedEpisodes)
    }
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

// Get biggest season watching or biggest season watched.
// This could probably be simpler but -_-
export function getLatestWatchedInTv(
  ws: WatchedSeason[] | undefined,
  we: WatchedEpisode[] | undefined
): string {
  if ((!ws || ws.length <= 0) && (!we || we.length <= 0)) {
    return "";
  }

  let biggestSeasonWatched = -1;
  let biggestSeasonWatching = -1;
  if (ws && ws.length > 0) {
    for (let i = 0; i < ws.length; i++) {
      const s = ws[i];
      if (s.status === "WATCHING") {
        if (s.seasonNumber > biggestSeasonWatching) {
          biggestSeasonWatching = s.seasonNumber;
        }
      } else if (s.status === "FINISHED") {
        if (s.seasonNumber > biggestSeasonWatched) {
          biggestSeasonWatched = s.seasonNumber;
        }
      }
    }
  }
  const season = biggestSeasonWatching >= 0 ? biggestSeasonWatching : biggestSeasonWatched;

  // Look for biggest watched/watching episode in season if any.
  // Does same thing as above.
  let episode: WatchedEpisode | undefined;
  if (we && we.length > 0) {
    let biggestEpisodeWatched: WatchedEpisode | undefined;
    let biggestEpisodeWatching: WatchedEpisode | undefined;
    for (let i = 0; i < we.length; i++) {
      const s = we[i];
      if (season >= 0 && s.seasonNumber !== season) continue;
      if (s.status === "WATCHING") {
        if (!biggestEpisodeWatching) {
          biggestEpisodeWatching = s;
        }
        if (
          s.episodeNumber > biggestEpisodeWatching.episodeNumber ||
          s.seasonNumber > biggestEpisodeWatching.seasonNumber
        ) {
          biggestEpisodeWatching = s;
        }
      } else if (s.status === "FINISHED") {
        if (!biggestEpisodeWatched) {
          biggestEpisodeWatched = s;
        }
        if (
          s.episodeNumber > biggestEpisodeWatched.episodeNumber ||
          s.seasonNumber > biggestEpisodeWatched.seasonNumber
        ) {
          biggestEpisodeWatched = s;
        }
      }
    }
    if (biggestEpisodeWatched || biggestEpisodeWatching) {
      episode =
        biggestEpisodeWatching !== undefined ? biggestEpisodeWatching : biggestEpisodeWatched;
    }
  }

  if (season >= 0 && episode) {
    return seasonAndEpToReadable(season, episode.episodeNumber);
  } else if (season >= 0) {
    return `Season ${season}`;
  } else if (episode) {
    return seasonAndEpToReadable(episode.seasonNumber, episode.episodeNumber);
  } else {
    return "";
  }
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
 * string in this format S1E1.
 */
export function seasonAndEpToReadable(season: number | undefined, episode: number | undefined) {
  return `S${typeof season === "number" ? String(season) : "(unknown)"}E${typeof episode === "number" ? String(episode) : "(unknown)"}`;
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
    return "Unknown";
  }
  const dn = new Date(Date.now());
  if (d.getFullYear() === dn.getFullYear()) {
    if (d.getMonth() === dn.getMonth()) {
      if (d.getDate() === dn.getDate()) {
        return "Today";
      }
    }
    return `${d.getDate()}${getOrdinalSuffix(d.getDate())} ${monthsShort[d.getMonth()]}`;
  }
  return `${d.getDate()}${getOrdinalSuffix(d.getDate())} ${monthsShort[d.getMonth()]} ${d.getFullYear()}`;
}
