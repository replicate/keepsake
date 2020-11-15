import Head from "next/head";

function Layout({ title, wholeTitle, description = "", children }) {
  if (!wholeTitle) {
    wholeTitle = `${title} | Replicate` || "Replicate";
  }
  return (
    <>
      <Head>
        <meta name="viewport" content="initial-scale=1.0, width=device-width" />
        <title>{wholeTitle}</title>
        <meta name="title" content={wholeTitle} />
        <meta name="description" content={description} />

        {/* https://metatags.io/ */}
        {/* Open Graph / Facebook */}
        <meta property="og:type" content="website" />
        <meta property="og:title" content={wholeTitle} />
        <meta property="og:description" content={description} />
        <meta property="og:image" content="" />

        {/* Twitter */}
        <meta property="twitter:card" content="summary" />
        <meta
          property="twitter:title"
          key="twitter:title"
          content={wholeTitle}
        />
        <meta property="twitter:description" content={description} />
        <meta property="twitter:image" content="" />
      </Head>

      <div className="layout">{children}</div>
    </>
  );
}

export default Layout;
