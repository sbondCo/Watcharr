import type { Icon, MediaType, Watched, WatchedStatus } from "@/types";

export const watchedStatuses: {
  [key in WatchedStatus]: Icon;
} = {
  PLANNED: "calendar",
  WATCHING: "clock",
  FINISHED: "check",
  HOLD: "pause",
  DROPPED: "thumb-down"
};

export function isTouch() {
  return "ontouchstart" in window;
}

// Not passing wList from #each loop caused it not to have reactivity.
// Passing it through must allow it to recognize it as a dependency?
export function getWatchedDependedProps(wid: number, wtype: MediaType, list: Watched[]) {
  const wel = list.find((wl) => wl.content.tmdbId === wid && wl.content.type === wtype);
  if (!wel) return {};
  console.log(wid, wtype, wel?.content.title, wel?.status, wel?.rating);
  return {
    status: wel.status,
    rating: wel.rating
  };
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
