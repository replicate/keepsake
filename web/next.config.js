const remarkSlug = require("remark-slug");
const withMDX = require("@next/mdx")({
  options: {
    remarkPlugins: [
      // Add ids to headings
      remarkSlug,
    ],
  },
});

module.exports = withMDX({
  pageExtensions: ["js", "jsx", "mdx"],
});
