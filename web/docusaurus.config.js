module.exports = {
  title: "Replicate",
  url: "https://your-docusaurus-test-site.com",
  baseUrl: "/",
  favicon: "img/favicon.ico",
  organizationName: "facebook", // Usually your GitHub org/user name.
  projectName: "docusaurus", // Usually your repo name.
  themeConfig: {
    disableDarkMode: true,
    sidebarCollapsible: false, // Perhaps enable this when we have more content
    navbar: {
      title: "Replicate",
      logo: {
        alt: "",
        src: "img/logo.svg",
      },
      links: [
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
          title: "",
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
      copyright: ``,
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
          customCss: require.resolve("./src/css/custom.css"),
        },
      },
    ],
  ],
};
