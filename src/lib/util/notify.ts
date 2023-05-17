import { notifications } from "@/store";
import { get } from "svelte/store";

export interface Notification {
  /**
   * Text shown in popup;
   */
  text: string;

  /**
   * Type of notification, controls the style.
   */
  type?: "error" | "success";

  /**
   * How long in milliseconds the popup will stay shown for.
   */
  time?: number;
}

export function notify(n: Notification) {
  const notifs = get(notifications);
  const id = Math.random();
  notifs.push({ id, ...n });
  notifications.update(() => notifs);
  setTimeout(() => unNotify(id), n.time ?? 2500);
  return id;
}

export function unNotify(id: number) {
  const ns = get(notifications);
  const dn = ns.filter((e) => e.id !== id);
  notifications.update(() => dn);
}
