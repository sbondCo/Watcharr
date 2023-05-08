export interface ToolTipOptions {
  text: string;
  pos?: "left" | "top";
}

export default function tooltip(node: HTMLElement, opts: ToolTipOptions) {
  const { text, pos = "left" } = opts;
  const tooltip = document.getElementById("tooltip");

  const show = () => {
    if (tooltip) {
      tooltip.innerHTML = text;
      const nrect = node.getBoundingClientRect();
      const trect = tooltip.getBoundingClientRect();
      nrect.y += window.scrollY; // Add scrollY to node dom rect so tooltip shows correcting when page is scrolled down
      if (pos === "left") {
        tooltip.style.left = `${nrect.x - trect.width - 10}px`;
        tooltip.style.top = `${nrect.y + trect.height / 2 - 10}px`;
      } else if (pos === "top") {
        tooltip.style.left = `${nrect.x - trect.width / 2 + nrect.width / 2}px`;
        tooltip.style.top = `${nrect.y - trect.height - 5}px`;
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
  node.addEventListener("touchcancel", hide);
  node.addEventListener("click", hide);

  return {
    destroy() {
      console.log("el destroyed");
      node.removeEventListener("mouseover", show);
      node.removeEventListener("touchstart", show);
      node.removeEventListener("mouseout", hide);
      node.removeEventListener("touchcancel", hide);
      node.removeEventListener("click", hide);
    }
  };
}
