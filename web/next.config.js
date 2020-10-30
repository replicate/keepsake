const mdx = require("@next/mdx");
const remarkSlug = require("remark-slug");

let config = {
  pageExtensions: ["js", "jsx", "mdx"],
  env: {
    TUTORIAL_COLAB_URL:
      "https://colab.research.google.com/drive/1vjZReg--45P-NZ4j8TXAJFWuepamXc7K",
    ANALYSIS_COLAB_URL:
      "https://colab.research.google.com/drive/1o-LQI3LvIzO2SWPugHby1YvRW7-9ZOtH",
  },
};

// Add MDX support
config = mdx({
  options: {
    remarkPlugins: [
      // Add ids to headings
      remarkSlug,
    ],
  },
})(config);

module.exports = config;
