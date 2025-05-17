/**
 * Helper function to execute a callback when the user clicks outside of a specified element.
 *
 * @param {HTMLElement} node The element within which click events are ignored.
 * @param {() => void} callback The function to be called when a click occurs outside the specified element.
 * @returns {{ destroy: () => void }} An object with a `destroy` method to remove event listeners and stop listening for clicks outside the element.
 */
export function clickOutside(
	node: HTMLElement,
	callback: (() => void) | undefined,
) {
	if (callback === undefined) {
		return;
	}

	let pointerDownStartedInside = false;

	function isClickOnExcludedTarget(
		target: (PointerEvent | KeyboardEvent)["target"],
	): boolean {
		if (!(target instanceof Node)) return false;
		let clicked = false;

		const elements = document.querySelectorAll("#exclude-outclick");
		for (const element of elements) {
			if (element && element.contains(target)) {
				clicked = true;
				break;
			}
		}

		return clicked;
	}

	const handlePointerDown = (event: PointerEvent) => {
		if (
			(!event.defaultPrevented &&
				node?.contains(event.target as HTMLElement)) ||
			isClickOnExcludedTarget(event.target)
		) {
			pointerDownStartedInside = true;
			return;
		}
		pointerDownStartedInside = false;
	};

	const handlePointerUp = (event: PointerEvent) => {
		if (
			!event.defaultPrevented &&
			!pointerDownStartedInside &&
			node &&
			!node.contains(event.target as HTMLElement)
		) {
			callback();
		}

		pointerDownStartedInside = false;
	};

	document.addEventListener("pointerup", handlePointerUp, true);
	document.addEventListener("pointerdown", handlePointerDown, true);

	return {
		destroy() {
			document.removeEventListener("pointerup", handlePointerUp, true);
			document.removeEventListener("pointerdown", handlePointerDown, true);
		},
	};
}
