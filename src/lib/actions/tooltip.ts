export interface ToolTipOptions {
	text: string;
	pos?: "left" | "top" | "bot";

	/**
	 * Only show tooltip if this condition is true.
	 */
	condition?: boolean;

	/**
	 * Disable the text capitalization style that is on by default.
	 */
	styleNoCapitalize?: boolean;
}

export default function tooltip(node: HTMLElement, opts: ToolTipOptions) {
	let { text, pos = "left", condition = true, styleNoCapitalize } = opts;
	const tooltip = document.getElementById("tooltip");

	const show = () => {
		if (!tooltip) {
			console.error("tooltip element wasn't found!");
			return;
		}
		if (!condition) {
			return;
		}

		const trimmedText = text?.trim();
		if (!trimmedText) {
			return;
		}
		tooltip.innerHTML = trimmedText;

		const nrect = node.getBoundingClientRect();
		const trect = tooltip.getBoundingClientRect();
		nrect.y += window.scrollY; // Add scrollY to node dom rect so tooltip shows correcting when page is scrolled down
		if (pos === "left") {
			tooltip.style.left = `${nrect.x - trect.width - 10}px`;
			tooltip.style.top = `${nrect.y + trect.height / 2 - 19.5}px`;
		} else if (pos === "top") {
			tooltip.style.left = `${nrect.x - trect.width / 2 + nrect.width / 2}px`;
			tooltip.style.top = `${nrect.y - trect.height - 5}px`;
		} else if (pos === "bot") {
			tooltip.style.left = `${nrect.x - trect.width / 2 + nrect.width / 2}px`;
			tooltip.style.top = `${nrect.y + trect.height + 5}px`;
		}

		if (styleNoCapitalize) {
			tooltip.classList.add("no-capitalize");
		} else {
			tooltip.classList.remove("no-capitalize");
		}

		tooltip.style.visibility = "visible";
	};

	const hide = () => {
		if (tooltip) {
			tooltip.style.visibility = "hidden";
		}
	};

	node.addEventListener("mouseover", show);
	node.addEventListener("touchstart", show);
	node.addEventListener("mouseout", hide);
	node.addEventListener("touchend", hide);
	node.addEventListener("click", hide);

	return {
		update(opts: ToolTipOptions) {
			text = opts.text;
			pos = opts.pos || "left";
			condition = opts.condition ?? true;
			styleNoCapitalize = opts.styleNoCapitalize;
		},
		destroy() {
			node.removeEventListener("mouseover", show);
			node.removeEventListener("touchstart", show);
			node.removeEventListener("mouseout", hide);
			node.removeEventListener("touchend", hide);
			node.removeEventListener("click", hide);
			// Running hide() here fixes scenario where we (eg) hide `node`
			// after it is pressed, but tooltip persists and there is no way
			// to remove it.
			hide();
		},
	};
}
