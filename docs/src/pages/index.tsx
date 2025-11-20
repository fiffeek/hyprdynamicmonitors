import type { ReactNode } from 'react';
import { useState, useEffect } from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Head from '@docusaurus/Head';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import Heading from '@theme/Heading';

import styles from './index.module.css';


function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  const [fontLoaded, setFontLoaded] = useState(false);

  useEffect(() => {
    if (typeof document !== 'undefined' && 'fonts' in document) {
      document.fonts.load('1em TheFont').then(() => {
        setFontLoaded(true);
      }).catch(() => {
        setFontLoaded(true);
      });
    } else {
      setFontLoaded(true);
    }
  }, []);

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
            <Heading as="h1" className={clsx("hero__title", styles.heroTitle)} style={{ visibility: fontLoaded ? 'visible' : 'hidden' }}>
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
              <iframe
                src="https://ghbtns.com/github-btn.html?user=fiffeek&repo=hyprdynamicmonitors&type=star&count=true&size=large"
                frameBorder="0"
                scrolling="0"
                width="130"
                height="30"
                title="GitHub Stars"
                style={{ border: 'none', marginLeft: '0.5rem', alignSelf: 'center', display: 'block' }}
              />
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
      <Head>
        <link
          rel="preload"
          href="https://garet.typeforward.com/assets/fonts/shared/TFMixVF.woff2"
          as="font"
          type="font/woff2"
          crossOrigin="anonymous"
        />
      </Head>
      <HomepageHeader />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
