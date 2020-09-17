import { MDXProvider } from "@mdx-js/react";
import { useRouter } from "next/router";
import { useEffect } from "react";

import CodeBlock from "../components/code-block";
import * as gtag from "../lib/gtag";

import "../styles/global.scss";

const mdxComponents = {
  pre: (props) => <div {...props} />,
  code: (props) => <CodeBlock {...props} />,
};

function MyApp({ Component, pageProps }) {
  const router = useRouter();
  useEffect(() => {
    const handleRouteChange = (url) => {
      gtag.pageview(url);
    };
    router.events.on("routeChangeComplete", handleRouteChange);
    return () => {
      router.events.off("routeChangeComplete", handleRouteChange);
    };
  }, [router.events]);

  return (
    <MDXProvider components={mdxComponents}>
      <Component {...pageProps} />{" "}
    </MDXProvider>
  );
}

export default MyApp;
