// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
async function createConfig() {
  const mdxMermaid = await import('mdx-mermaid')

  return {
    title: 'Constellation',
    tagline: 'Constellation: The world\'s most secure Kubernetes',
    url: 'https://constellation-docs.netlify.app',
    baseUrl: '/constellation/',
    onBrokenLinks: 'throw',
    onBrokenMarkdownLinks: 'warn',
    favicon: 'img/favicon.ico',

    // GitHub pages deployment config.
    // If you aren't using GitHub pages, you don't need these.
    organizationName: 'Edgeless Systems', // Usually your GitHub org/user name.
    projectName: 'Constellation', // Usually your repo name.

    // Even if you don't use internalization, you can use this field to set useful
    // metadata like html lang. For example, if your site is Chinese, you may want
    // to replace "en" with "zh-Hans".
    i18n: {
      defaultLocale: 'en',
      locales: ['en'],
    },

    presets: [
      [
        'classic',
        /** @type {import('@docusaurus/preset-classic').Options} */
        ({
          docs: {
            remarkPlugins: [[mdxMermaid.default, { mermaid: {
              theme: 'base',
              themeVariables: {
                // general
                'fontFamily': '"Open Sans", sans-serif',
                'primaryColor': '#90FF99', // edgeless green
                'primaryTextColor': '#000000',
                'secondaryColor': '#A5A5A5', // edgeless grey
                'secondaryTextColor': '#000000',
                'tertiaryColor': '#E7E6E6', // edgeless light grey
                'tertiaryTextColor': '#000000',
                // flowchart
                'clusterBorder': '#A5A5A5',
                'clusterBkg': '#ffffff',
                'edgeLabelBackground': '#ffffff',
                // sequence diagram
                'activationBorderColor': '#000000',
                'actorBorder': '#A5A5A5',
                'actorFontFamily': '"Open Sans", sans-serif', // not released by mermaid yet
                'noteBkgColor': '#8B04DD', // edgeless purple
                'noteTextColor': '#ffffff',
              },
              startOnLoad: true
            }}]],
            sidebarPath: require.resolve('./sidebars.js'),
            // sidebarPath: 'sidebars.js',
            // Please change this to your repo.
            // Remove this to remove the "edit this page" links.
            editUrl: ({ locale, docPath }) => {
              return `https://github.com/edgelesssys/constellation-docs/edit/ref/docusarus/docs/${docPath}`;
            },
            routeBasePath: "/"
          },
          blog: false,
          theme: {
            customCss: require.resolve('./src/css/custom.css'),
          },
        }),
      ],
    ],

    themeConfig:
      /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
      ({
        navbar: {
          hideOnScroll: false,
          logo: {
            alt: 'Constellation Logo',
            src: 'img/logos/constellation_oneline.svg',
          },
          items: [
            // left
            // Running docs only mode no need for a link here
            // {
            //   type: 'doc',
            //   docId: 'intro',
            //   position: 'left',
            //   label: 'Docs',
            // },
            // right
            {
              type: 'docsVersionDropdown',
              position: 'right',
            },
            {
              href: 'https://github.com/edgelesssys/constellation',
              position: 'right',
              className: 'header-github-link',
            },
          ],
        },
        colorMode: {
          defaultMode: 'light',
          disableSwitch: true,
          respectPrefersColorScheme: false,
        },
        announcementBar: {
          content:
            '⭐️ If you like Constellation, give it a star on <a target="_blank" rel="noopener noreferrer" href="https://github.com/edgelesssys/constellation">GitHub</a>! ⭐️',
        },
        footer: {
          style: 'dark',
          links: [
            {
              title: 'Learn',
              items: [
                {
                  label: 'Confidential Kubernetes',
                  to: '/overview/confidential-kubernetes',
                },
                {
                  label: 'Install',
                  to: '/getting-started/install',
                },
                {
                  label: 'First steps',
                  to: '/getting-started/first-steps',
                },
              ],
            },
            {
              title: 'Community',
              items: [
                {
                  label: 'GitHub',
                  href: 'https://github.com/edgelesssys/constellation',
                },
                {
                  label: 'Discord',
                  href: 'https://discord.gg/rH8QTH56JN',
                },
                {
                  label: 'Newsletter',
                  href: 'https://www.edgeless.systems/#newsletter-signup'
                },
              ],
            },
            {
              title: 'Social',
              items: [
                {
                  label: 'Blog',
                  to: 'https://blog.edgeless.systems/',
                },
                {
                  label: 'Twitter',
                  href: 'https://twitter.com/EdgelessSystems',
                },
                {
                  label: 'LinkedIn',
                  href: 'https://www.linkedin.com/company/edgeless-systems/',
                },

                {
                  label: 'Youtube',
                  href: 'https://www.youtube.com/channel/UCOOInN0sCv6icUesisYIDeA',
                },
              ],
            },
          ],
          copyright: `Copyright © ${new Date().getFullYear()} Edgeless Systems. Built with Docusaurus.`,
        },
        prism: {
          theme: lightCodeTheme,
          darkTheme: darkCodeTheme,
        },
      }),
    }
};

module.exports = createConfig;
