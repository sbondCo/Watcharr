// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const { themes } = require("prism-react-renderer");
const lightTheme = themes.github;
const darkTheme = themes.dracula;

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "Watcharr Docs",
  tagline:
    "Open source, self-hostable watched list for all your content with user authentication, modern and clean UI and a very simple setup. ",
  favicon: "img/favicon.png",

  // Set the production url of your site here
  url: "https://watcharr.app",
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: "/",

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: "sbondCo", // Usually your GitHub org/user name.
  projectName: "Watcharr", // Usually your repo name.

  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",

  trailingSlash: false,

  // Even if you don't use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is Chinese, you may want
  // to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: "en",
    locales: ["en"]
  },

  presets: [
    [
      "classic",
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve("./sidebars.js"),
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl: "https://github.com/sbondCo/Watcharr/tree/dev/doc"
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl: "https://github.com/sbondCo/Watcharr/tree/dev/doc"
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css")
        }
      })
    ]
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      image: "img/social-card.png",
      navbar: {
        title: "Watcharr",
        logo: {
          alt: "Watcharr Logo",
          src: "img/favicon.png"
        },
        items: [
          {
            type: "docSidebar",
            sidebarId: "tutorialSidebar",
            position: "left",
            label: "Docs"
          },
          {
            type: "docsVersionDropdown"
          },
          {
            href: "https://beta.watcharr.app",
            label: "Demo",
            position: "right"
          },
          {
            href: "https://github.com/sbondCo/Watcharr",
            label: "GitHub",
            position: "right"
          }
        ]
      },
      prism: {
        theme: lightTheme,
        darkTheme: darkTheme
      },
      colorMode: {
        defaultMode: "light",
        // Dark theme currently disabled.. no time to fix the icons etc.
        disableSwitch: true
      }
    })
};

module.exports = config;
