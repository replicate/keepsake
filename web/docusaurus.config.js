module.exports = {
  title: "Replicate",
  url: "https://beta.replicate.ai",
  customFields: {
    version: "0.1.21",
  },
  baseUrl: "/",
  favicon: "img/favicon.ico",
  organizationName: "facebook", // Usually your GitHub org/user name.
  projectName: "docusaurus", // Usually your repo name.
  plugins: ["docusaurus-plugin-sass"],
  themeConfig: {
    colorMode: {
      defaultMode: "light",
      disableSwitch: true,
      respectPrefersColorScheme: false,
    },
    sidebarCollapsible: false, // Perhaps enable this when we have more content
    navbar: {
      title: "Replicate",
      logo: {
        alt: " ",
        src: "img/logo.svg",
      },
      items: [
        {
          to: "docs/tutorial",
          activeBasePath: "docs",
          label: "Docs",
          position: "left",
        },
        {
          href: "https://github.com/replicate/replicate",
          label: "GitHub",
          position: "right",
        },
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          title: " ",
          items: [
            {
              label: "GitHub",
              to: "https://github.com/replicate/replicate",
            },
            {
              label: "team@replicate.ai",
              to: "mailto:team@replicate.ai",
            },
          ],
        },
      ],
      copyright: ` `,
    },
    prism: {
      additionalLanguages: ["shell-session"],
      theme: require("prism-react-renderer/themes/oceanicNext"),
    },
    googleAnalytics: {
      // Development ID by default, set this in production
      trackingID: process.env.GOOGLE_ANALYTICS_TRACKING_ID || "UA-107304984-6",
    },
  },
  presets: [
    [
      "@docusaurus/preset-classic",
      {
        docs: {
          sidebarPath: require.resolve("./sidebars.js"),
          editUrl: "https://github.com/replicate/replicate/edit/master/web/",
        },
        theme: {
          customCss: require.resolve("./src/css/custom.scss"),
        },
      },
    ],
  ],
};
