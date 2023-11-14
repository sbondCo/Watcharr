export interface ToolTipOptions {
  text: string;
  pos?: "left" | "top" | "bot";

  /**
   * Only show tooltip if this condition is true.
   */
  condition?: boolean;
}

export default function tooltip(node: HTMLElement, opts: ToolTipOptions) {
  let { text, pos = "left", condition = true } = opts;
  const tooltip = document.getElementById("tooltip");

  const show = () => {
    if (!condition) return;
    if (tooltip) {
      tooltip.innerHTML = text;
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
      tooltip.style.visibility = "visible";
    }
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
      condition = opts.condition ?? true;
    },
    destroy() {
      node.removeEventListener("mouseover", show);
      node.removeEventListener("touchstart", show);
      node.removeEventListener("mouseout", hide);
      node.removeEventListener("touchend", hide);
      node.removeEventListener("click", hide);
    }
  };
}
