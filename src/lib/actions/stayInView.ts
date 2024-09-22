export interface StayInViewOptions {
  /**
   * If the `node` contains an element (like an arrow for menus),
   * that should be shifted over to account for any shifting
   * of the `node` itself, then pass a selector for it here.
   */
  elToShiftSelector?: string;
}

export default function stayInView(node: HTMLElement, opts: StayInViewOptions) {
  console.debug("stayInView: Initial opts:", opts);
  let { elToShiftSelector } = opts;
  let viewDeb: ReturnType<typeof setTimeout>;

  /**
   * Move element to in view, if it isn't.
   *
   * Currently only supports moving `node` back into view if oob
   * on the left. May need to support other sides and/or resizing
   * when `node` still wont fit after shifting it in the future.
   */
  const getInView = () => {
    const nrect = node.getBoundingClientRect();
    const brect = document.body.getBoundingClientRect();
    console.debug("stayInView->getInView: Called.", nrect, brect);
    if (nrect.x <= brect.x) {
      const diff = nrect.x - brect.x + 10;
      console.debug(
        "stayInView->getInView: Node is out of bounds on the left, shifting forwards to:",
        diff
      );
      node.style.left = `${diff}px`;
      if (elToShiftSelector) {
        const elToShift = node.querySelector(elToShiftSelector) as HTMLElement;
        if (elToShift) {
          console.debug("stayInView->getInView: Shifting elToShift.");
          const nrectNew = node.getBoundingClientRect();
          const arrowDiff = nrectNew.left - nrect.left;
          elToShift.style.left = `${elToShift.offsetLeft - arrowDiff}px`;
        } else {
          console.warn("elToShift not found.", elToShiftSelector);
        }
      }
    }
  };

  const getInViewDeb = () => {
    clearTimeout(viewDeb);
    viewDeb = setTimeout(getInView, 200);
  };

  window.addEventListener("resize", getInViewDeb);
  getInView();

  return {
    update(opts: StayInViewOptions) {
      console.debug("stayInView: Opts updated", opts);
      elToShiftSelector = opts.elToShiftSelector;
      getInView();
    },
    destroy() {
      window.removeEventListener("resize", getInViewDeb);
    }
  };
}
