import { MDXProvider } from "@mdx-js/react";
import { useRouter } from "next/router";
import { useEffect } from "react";
import Link from "next/link";

import CodeBlock from "../components/code-block";
import * as gtag from "../lib/gtag";

import "../styles/global.scss";

// https://github.com/vercel/next.js/discussions/11110?sort_order=relevance
const CustomLink = (props) => {
  const href = props.href;
  const isInternalLink = href && (href.startsWith("/") || href.startsWith("#"));

  if (isInternalLink) {
    return (
      <Link href={href}>
        <a {...props} />
      </Link>
    );
  }
  return <a target="_blank" {...props} />;
};

const mdxComponents = {
  a: CustomLink,
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
