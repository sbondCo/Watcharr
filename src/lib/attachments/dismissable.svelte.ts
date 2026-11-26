import { page } from "$app/state";
import { untrack } from "svelte";
import type { Attachment } from "svelte/attachments";

/**
 * Calls passed in `cb` callback on dismiss guesture (clicking elsewhere,
 * pressing ESC) by user.
 *
 * NOTE: Best to use this with a condition on the @attach when possible (ie,
 * only attach when the Menu we want to support dismissing on is open) so the
 * events stay registered only while necessary.
 */
export default function dismissable(cb: () => void): Attachment<HTMLElement> {
	return (node) => {
		// Dismiss when: Clicking anywhere but the `node`.
		const onClick = (e: MouseEvent) => {
			if (!node.contains(e.target as Node)) {
				console.debug("dismissable: from click elsewhere");
				cb();
			}
		};

		// Dismiss when: Pressing `ESC`.
		const onKeydown = (e: KeyboardEvent) => {
			if (e.key === "Escape") {
				console.debug("dismissable: from ESC");
				cb();
			}
		};

		// Dismiss when: Scrolling page.
		const onScroll = () => {
			console.debug("dismissable: from scroll");
			cb();
		};

		// Dismiss when: User navigates.
		const initialUrl = untrack(() => page.url.href);
		$effect(() => {
			if (page.url.href !== initialUrl) {
				console.debug("dismissable: from navigate");
				untrack(() => cb());
			}
		});

		document.addEventListener("click", onClick, true);
		document.addEventListener("keydown", onKeydown);
		document.addEventListener("scroll", onScroll);

		return () => {
			document.removeEventListener("click", onClick, true);
			document.removeEventListener("keydown", onKeydown);
			document.removeEventListener("scroll", onScroll);
		};
	};
}
