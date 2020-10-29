import Head from "next/head";
import Link from "next/link";
import {
  faGithub,
  faTwitter,
  faSoundcloud,
} from "@fortawesome/free-brands-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

function Layout({ title, children }) {
  return (
    <>
      <Head>
        <meta name="viewport" content="initial-scale=1.0, width=device-width" />
        <title>{title || "Replicate"}</title>
      </Head>

      <div className="global-banner">
        <p>
          Welcome to the Replicate private beta! If you have any feedback, or
          know somebody who might like it, email us:{" "}
          <a href="mailto:team@replicate.ai">team@replicate.ai</a>
        </p>
      </div>

      <div className="layout">
        {children}
        <footer>
          <h2>
            <div>
              <Link href="/docs">
                <a className="button">Get started</a>
              </Link>
            </div>
            <div> or, </div>
            <div>
              <Link href="/docs/learn/how-it-works">
                <a>learn more about how Replicate works</a>
              </Link>
            </div>
          </h2>
          <div id="contributors">
            <h3>Made by</h3>
            <div>
              <figure>
                <div
                  style={{ backgroundImage: "url(" + "/images/ben.jpg" + ")" }}
                ></div>
                <figcaption>
                  <h4>Ben Firshman</h4>
                  <p>Product at Docker, creator of Docker&nbsp;Compose.</p>
                  <p>
                    <a href="https://github.com/bfirsh" className="link">
                      <FontAwesomeIcon icon={faGithub} />
                    </a>
                    <a href="https://twitter.com/bfirsh" className="link">
                      <FontAwesomeIcon icon={faTwitter} />
                    </a>
                  </p>
                </figcaption>
              </figure>
              <figure>
                <div
                  style={{
                    backgroundImage: "url(" + "/images/andreas.jpg" + ")",
                  }}
                ></div>
                <figcaption>
                  <h4>Andreas Jansson</h4>
                  <p>ML infrastructure and research at Spotify.</p>
                  <p>
                    <a
                      href="https://github.com/andreasjansson"
                      className="link"
                    >
                      <FontAwesomeIcon icon={faGithub} />
                    </a>
                  </p>
                </figcaption>
              </figure>
            </div>
            <p>
              We want to make machine learning tools that are less terrible. We
              also built{" "}
              <a href="https://www.arxiv-vanity.com/">arXiv Vanity</a>, which
              lets you read arXiv papers as responsive web pages.
            </p>
          </div>
          {/* 
          <div id="get-involved">
            <h3>Get involved</h3>
            <p>
              Placeholder diffs everything, all the way down to versions of
              dependencies, just in case that latest Tensorflow version did
              something weird.
            </p>
            <a className="button" href="#">
              Report a bug
            </a>
          </div> */}
          <nav>
            <Link href="/docs">
              <a>Docs</a>
            </Link>
            <a href="https://github.com/replicate/replicate">GitHub</a>
            <a href="mailto:team@replicate.ai">team@replicate.ai</a>
          </nav>
          <p className="tagline">
            <strong>Replicate</strong> Version control for machine&nbsp;learning
          </p>
        </footer>
      </div>
    </>
  );
}

export default Layout;
