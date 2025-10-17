import type { ReactNode } from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import Heading from '@theme/Heading';

import styles from './index.module.css';


function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className={styles.heroBackground}>
        <img
          src="/previews/demo.gif"
          alt="HyprDynamicMonitors TUI Demo"
          className={styles.heroGif}
        />
      </div>
      <div className={styles.heroOverlay}></div>
      <div className="container" style={{ position: 'relative', zIndex: 2 }}>
        <div className="row">
          <div className="col col--12 text--center">
            <div className={styles.heroTitleContainer}>
              {['Dynamic', 'Monitor', 'Management'].map((word, index) => (
                <span key={index} className={styles.fadeWord} style={{ animationDelay: `${index * 0.1}s` }}>
                  {word}{' '}
                </span>
              ))}
            </div>
            <Heading as="h1" className={clsx("hero__title", styles.heroTitle)}>
              {siteConfig.title.split('').map((letter, index) => {
                const breatheClass = letter === 'H' ? styles.breatheLetterH :
                  letter === 'D' ? styles.breatheLetterD :
                    letter === 'M' ? styles.breatheLetterM : '';
                return (
                  <span key={index} className={breatheClass}>
                    {letter}
                  </span>
                );
              })}
            </Heading>
            <p className={clsx("hero__subtitle", styles.heroSubtitle)}>
              {siteConfig.tagline.split(' ').map((word, index) => (
                <span key={index} className={styles.fadeSubtitleWord} style={{ animationDelay: `${index * 0.08}s` }}>
                  {word}{' '}
                </span>
              ))}
            </p>
            <div className={styles.buttons}>
              <Link
                className="button button--secondary button--lg"
                to="docs/category/quick-start">
                Get Started
              </Link>
              <Link
                className="button button--outline button--secondary button--lg"
                to="/docs/">
                Learn More
              </Link>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}

export default function Home(): ReactNode {
  const { siteConfig } = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title}`}
      description="Event-driven monitor configuration service for Hyprland with interactive TUI and profile-based management">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
