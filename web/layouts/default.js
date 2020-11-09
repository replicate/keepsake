import Head from "next/head";

function Layout({ title, children }) {
  return (
    <>
      <Head>
        <meta name="viewport" content="initial-scale=1.0, width=device-width" />
        <title>{title || "Replicate"}</title>
      </Head>

      <div className="layout">{children}</div>
    </>
  );
}

export default Layout;
