import Head from "next/head";

function Layout({ title, children }) {
  return (
    <>
      <Head>
        <meta name="viewport" content="initial-scale=1.0, width=device-width" />
        <title>{title || "Replicate"}</title>
      </Head>

      <div className="global-banner">
        <p>Welcome to the Replicate private beta! If you have any feedback, or know somebody who might like it, email us:{" "}<a href="mailto:team@replicate.ai">team@replicate.ai</a></p>
        <nav>
          <a href="/docs">Docs</a>
          <a href="https://github.com/replicate/replicate">GitHub</a>
        </nav>

      </div>

      <div className="layout">
        {children}
        <footer>
          <h2>
            <div>
              <a className="button" href="/docs">
                Get started
              </a>
            </div>
            <div> or, </div>
            <div>
              <a href="/docs">learn more about how Replicate works</a>
            </div>
          </h2>
          {/* <div id="contributors">
            <h3>Contributors</h3>
            <ul>
              <li>
                <img src="assets/img/photo.png" />
                <h4>Name Surname</h4>
                <p>
                  A brief biog or some details about what they have done on the
                  project. <a href="#">More</a>
                </p>
              </li>
              <li>
                <img src="assets/img/photo.png" />
                <h4>Name Surname</h4>
                <p>
                  A brief biog or some details about what they have done on the
                  project. <a href="#">More</a>
                </p>
              </li>
              <li>
                <img src="assets/img/photo.png" />
                <h4>Name Surname</h4>
                <p>
                  A brief biog or some details about what they have done on the
                  project. <a href="#">More</a>
                </p>
              </li>
            </ul>
          </div>
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
            <a href="/docs">Docs</a>
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
