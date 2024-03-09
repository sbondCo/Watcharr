import React, { useEffect } from "react";
import styles from "./styles.module.css";
import clsx from "clsx";
import Viewer from "viewerjs";
import "@site/node_modules/viewerjs/dist/viewer.min.css";

function TwoColFeat({ title, description, imgSrc }) {
  return (
    <div className={clsx(styles.twoCol)}>
      <div>
        <h2>{title}</h2>
        <p>{description}</p>
      </div>
      <div>
        <img src={imgSrc} alt="" />
      </div>
    </div>
  );
}

export function HomepageTwoColFeats() {
  useEffect(() => {
    const ctr = document.getElementById("homepage-two-col-feats");
    console.debug("Starting image viewer", ctr);
    new Viewer(ctr, {
      navbar: 0,
      toolbar: {
        zoomIn: 2,
        zoomOut: 2,
        next: 2,
        prev: 2
      },
      title: 0,
      className: "twoColViewer"
    });
  }, []);

  return (
    <section id="homepage-two-col-feats" className={styles.twoColFeats}>
      <TwoColFeat
        title="Neat UI"
        description="We just look lightweight."
        imgSrc="https://github.com/sbondCo/Watcharr/raw/dev/screenshot/homepage.png"
      />
      <TwoColFeat
        title="Easy To Use"
        description="Not much to it, other than ease of use."
        imgSrc="https://github.com/sbondCo/Watcharr/raw/dev/screenshot/homepage-poster-change-status.png"
      />
      <TwoColFeat
        title="In-Depth Details"
        description="About your favourite content."
        imgSrc="https://github.com/sbondCo/Watcharr/raw/dev/screenshot/content-details-page.png"
      />
      <TwoColFeat
        title="Personal Stats and Configuration"
        description="Make it how you like."
        imgSrc="https://github.com/sbondCo/Watcharr/raw/dev/screenshot/user-profile.png"
      />
      <TwoColFeat
        title="Discovery"
        description="Find the latest trending content."
        imgSrc="https://github.com/sbondCo/Watcharr/raw/dev/screenshot/discover-page.png"
      />
      <TwoColFeat
        title="Dark Theme"
        description="For you who are still unsatisfied. ;("
        imgSrc="https://github.com/sbondCo/Watcharr/raw/dev/screenshot/content-details-page-dark.png"
      />
    </section>
  );
}
