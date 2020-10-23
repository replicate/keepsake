import Header from "../components/header";
import Layout from "./default";

function DocsLayout({ title, children, ...props }) {
  return (
    <Layout {...props}>
      <Header className="documentation">
        <div className="breadcrumb">
          <a href="/">Home</a>&nbsp;
          {title ? (
            <>
              <a href="/docs">
                <span>Documentation</span>
              </a>
              &nbsp;<h2>{title}</h2>
            </>
          ) : (
            <h2>Documentation</h2>
          )}
        </div>
      </Header>

      <section className="docs documentation">
        <nav>
          <ol>
            <li>
              <h2>
                <a href="/docs/tutorial">Tutorial</a>
              </h2>
              <ol>
                <li>
                  <a href="/docs/tutorial">First steps</a>
                </li>
                <li>
                  <a href="/docs/tutorial/cloud-storage">
                    Store data in the cloud
                  </a>
                </li>
              </ol>
            </li>
            <li>
              <h2>Guides</h2>
              <ol>
                <li>
                  <a href="/docs/guides/notebooks">Working in notebooks</a>
                </li>
                <li>
                  <a href="/docs/guides/keras-integration">Keras integration</a>
                </li>
              </ol>
            </li>
            <li>
              <h2>Learning</h2>
              <ol>
                <li>
                  <a href="/docs/learn/how-it-works">How it works</a>
                </li>
                <li>
                  <a href="/docs/learn/analytics">Analytics</a>
                </li>
              </ol>
            </li>
            <li>
              <h2>Reference</h2>
              <ol>
                <li>
                  <a href="/docs/reference/python">Python library</a>
                </li>
                <li>
                  <a href="/docs/reference/yaml">replicate.yaml</a>
                </li>
                <li>
                  <a href="/docs/reference/cli">Command-line interface</a>
                </li>
              </ol>
            </li>
          </ol>
        </nav>
        <div className="body">{children}</div>
      </section>
    </Layout>
  );
}
export default DocsLayout;
