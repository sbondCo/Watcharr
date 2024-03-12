import React from "react";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import Layout from "@theme/Layout";
import HomepageFeatures from "@site/src/components/HomepageFeatures";
import LogoCol from "@site/static/img/logo-col.png";

import styles from "./index.module.css";
import clsx from "clsx";
import { HomepageTwoColFeats } from "../components/TwoColFeat";

function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className="hero shadow--lw">
      <div className={styles.heroContainer}>
        <div className={styles.heroWrapper}>
          <img src={LogoCol} />
          <h1 className="hero__title">{siteConfig.title}</h1>
          <p className="hero__subtitle">{siteConfig.tagline}</p>
          <div className={styles.buttons}>
            <Link className="button button--secondary button--lg" to="/docs/introduction">
              Get Started
            </Link>
            <Link
              className="button button--secondary button--outline button--lg"
              to="https://beta.watcharr.app"
              target="_blank"
            >
              Demo
            </Link>
          </div>
        </div>
      </div>
    </header>
  );
}

export default function Home() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title}`}
      description="Description will go into a meta tag in <head />"
    >
      <HomepageHeader />
      <main>
        <HomepageFeatures />
        <HomepageTwoColFeats />
      </main>
    </Layout>
  );
}
