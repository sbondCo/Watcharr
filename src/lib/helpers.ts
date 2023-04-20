import type { Icon, WatchedStatus } from "@/types";

export const watchedStatuses: {
  [key in WatchedStatus]: Icon;
} = {
  PLANNED: "calendar",
  WATCHING: "clock",
  FINISHED: "check",
  HOLD: "pause",
  DROPPED: "thumb-down"
};
