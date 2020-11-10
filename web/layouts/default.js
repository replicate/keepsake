import Head from "next/head";

function Layout({ title, wholeTitle, children }) {
  return (
    <>
      <Head>
        <meta name="viewport" content="initial-scale=1.0, width=device-width" />
        <title>{wholeTitle || `${title} | Replicate` || "Replicate"}</title>
      </Head>

      <div className="layout">{children}</div>
    </>
  );
}

export default Layout;
