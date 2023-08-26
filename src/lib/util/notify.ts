import { notifications } from "@/store";
import { get } from "svelte/store";

export interface Notification {
  /**
   * Notification ID.
   * Used to reference an exiting notification.
   */
  id?: number;

  /**
   * Text shown in popup;
   */
  text: string;

  /**
   * Type of notification, controls the style.
   */
  type?: "error" | "success" | "loading";

  /**
   * How long in milliseconds the popup will stay shown for.
   */
  time?: number;
}

export function notify(n: Notification) {
  const notifs = get(notifications);
  if (n.id) {
    const notif = notifs.find((not) => not.id === n.id);
    if (notif) {
      notif.type = n.type;
      notif.text = n.text;
      notifications.update(() => notifs);
    } else {
      console.error("Can't update notif that doesnt exist", n);
    }
  } else {
    n.id = Math.random();
    notifs.push({ ...n });
    notifications.update(() => notifs);
  }
  if (n.type !== "loading") setTimeout(() => unNotify(n.id!), n.time ?? 2500);
  return n.id;
}

export function unNotify(id: number) {
  const ns = get(notifications);
  const dn = ns.filter((e) => e.id !== id);
  notifications.update(() => dn);
}
