import React from "react";
import clsx from "clsx";
import styles from "./styles.module.css";

const FeatureList = [
  {
    title: "Free and Open",
    Svg: require("@site/static/img/icon/code-slash.svg").default,
    description: (
      <>
        Watcharr is free and open source software distributed under the MIT license. Feel free to
        browse, modify or contribute!
      </>
    )
  },
  {
    title: "Simple",
    Svg: require("@site/static/img/icon/happy.svg").default,
    description: (
      <>
        Watcharr is incredibly easy to dive into. Install without hassle and get started with the
        seamless and intuitive user experience.
      </>
    )
  },
  {
    title: "Built With Go and Svelte",
    Svg: require("@site/static/img/icon/go.svg").default,
    description: (
      <>Don't know what else to add here, so now you know Watcharr is built with Go and Svelte.</>
    )
  }
];

function Feature({ Svg, title, description }) {
  return (
    <div className={clsx("col col--4")}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
