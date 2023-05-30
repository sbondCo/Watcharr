import type { Icon, MediaType, TMDBContentCreditsCrew, Watched, WatchedStatus } from "@/types";

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
