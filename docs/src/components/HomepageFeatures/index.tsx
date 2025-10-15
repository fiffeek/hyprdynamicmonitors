import type { ReactNode } from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  description: ReactNode;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Event-Driven Architecture',
    description: (
      <>
        Automatically responds to monitor connect/disconnect events and power state
        changes in real-time. Seamlessly switch between profiles as your setup changes.
      </>
    ),
  },
  {
    title: 'Interactive TUI',
    description: (
      <>
        Visual terminal interface for configuring monitors with real-time preview.
        Create and manage profiles without editing configuration files manually.
      </>
    ),
  },
  {
    title: 'Profile-Based Management',
    description: (
      <>
        Define different monitor configurations for different setups. Use templates
        for dynamic configuration based on power state, lid state, and connected monitors.
      </>
    ),
  },
];

function Feature({ title, description }: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        {/* <Svg className={styles.featureSvg} role="img" /> */}
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
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
