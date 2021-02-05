const mdx = require("@next/mdx");
const remarkSlug = require("remark-slug");

let config = {
  pageExtensions: ["js", "jsx", "mdx"],
  env: {
    TUTORIAL_COLAB_URL:
      "https://colab.research.google.com/drive/1zzcQHd5ZtIgLA_vQjdQ0PUcAfYwcc43H",
    ANALYSIS_COLAB_URL:
      "https://colab.research.google.com/drive/11maIJO1C1yKSagTRyCJkPMAj6EjgLa73",
    INFERENCE_COLAB_URL:
      "https://colab.research.google.com/drive/11maIJO1C1yKSagTRyCJkPMAj6EjgLa73#scrollTo=J_Z02qBik8c9",
  },
  async redirects() {
    return [
      {
        source: "/docs/guides/production",
        destination: "/docs/guides/inference",
        permanent: true,
      },
    ];
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
