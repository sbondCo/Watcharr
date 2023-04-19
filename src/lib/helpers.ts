import type { WatchedStatus } from "@/types";

// TODO couldn't import Icon type from the component, figure it out later.
export function iconFromStatus(status: WatchedStatus): string {
  switch (status) {
    case "DROPPED":
      return "thumb-down";
    case "PLANNED":
      return "calendar";
    case "WATCHING":
      return "clock";
    case "FINISHED":
      return "check";
    case "HOLD":
      return "pause";
  }
}
