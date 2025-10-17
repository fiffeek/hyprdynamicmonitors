import { themes as prismThemes } from 'prism-react-renderer';
import type { Config } from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';


const config: Config = {
  title: 'HyprDynamicMonitors',
  tagline: 'Event-driven monitor configuration for Hyprland',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://hyprdynamicmonitors.filipmikina.com',
  baseUrl: '',

  organizationName: 'fiffeek',
  projectName: 'hyprdynamicmonitors',

  onBrokenLinks: 'throw',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  plugins: [
    [
      '@easyops-cn/docusaurus-search-local',
      {
        hashed: true,
        language: ['en'],
        highlightSearchTermsOnTargetPage: true,
        explicitSearchResultPath: true,
        indexDocs: true,
        indexBlog: false,
        indexPages: false,
      },
    ],
  ],

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl:
            'https://github.com/fiffeek/hyprdynamicmonitors/tree/main/docs/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    // Replace with your project's social card
    image: 'img/docusaurus-social-card.jpg',
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'HyprDynamicMonitors',
      logo: {
        alt: 'hyprdynamicmonitors',
        src: 'img/logo.png',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'tutorialSidebar',
          position: 'left',
          label: 'Documentation',
        },
        {
          href: 'https://github.com/fiffeek/hyprdynamicmonitors',
          label: 'GitHub',
          position: 'right',
        },
        {
          type: 'docsVersionDropdown',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {
              label: 'Getting Started',
              to: '/docs/',
            },
            {
              label: 'Quick Start',
              to: '/docs/category/quick-start',
            },
            {
              label: 'Configuration',
              to: '/docs/category/configuration',
            },
          ],
        },
        {
          title: 'Resources',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/fiffeek/hyprdynamicmonitors',
            },
            {
              label: 'Examples',
              href: 'https://github.com/fiffeek/hyprdynamicmonitors/tree/main/examples',
            },
          ],
        },
      ],
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'toml', 'go', 'ini'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
