const mdx = require("@next/mdx");
const remarkSlug = require("remark-slug");

let config = {
  pageExtensions: ["js", "jsx", "mdx"],
  env: {
    TUTORIAL_COLAB_URL:
      "https://colab.research.google.com/drive/1iypzEOysdACpIIXrM0RoGz9kg6z93XMP",
    ANALYSIS_COLAB_URL:
      "https://colab.research.google.com/drive/18sVRE4Zi484G2rBeOYjobE3zek2gDBvy",
    INFERENCE_COLAB_URL:
      "https://colab.research.google.com/drive/18sVRE4Zi484G2rBeOYjobE3zek2gDBvy#scrollTo=J_Z02qBik8c9",
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
