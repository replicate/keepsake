import Head from "next/head";

function Layout({ title, wholeTitle, description = "", children }) {
  if (!wholeTitle) {
    wholeTitle = `${title} | Keepsake` || "Keepsake";
  }
  return (
    <>
      <Head>
        <meta name="viewport" content="initial-scale=1.0, width=device-width" />
        <title>{wholeTitle}</title>
        <meta name="title" content={wholeTitle} />
        <meta name="description" content={description} />
        <link
          rel="apple-touch-icon"
          sizes="180x180"
          href="/apple-touch-icon.png"
        />
        <link rel="icon" type="image/png" href="/favicon.png" />
        <link rel="manifest" href="/site.webmanifest" />
        {/* https://metatags.io/ */}
        {/* Open Graph / Facebook */}
        <meta property="og:type" content="website" />
        <meta property="og:title" content={wholeTitle} />
        <meta property="og:description" content={description} />
        <meta
          property="og:image"
          content={process.env.NEXT_PUBLIC_URL + "/apple-touch-icon.png"}
        />
        {/* Twitter */}
        <meta property="twitter:card" content="summary" />
        <meta
          property="twitter:title"
          key="twitter:title"
          content={wholeTitle}
        />
        <meta property="twitter:description" content={description} />
        <meta
          property="twitter:image"
          content={process.env.NEXT_PUBLIC_URL + "/apple-touch-icon.png"}
        />
      </Head>

      <div className="layout">{children}</div>
    </>
  );
}

export default Layout;
