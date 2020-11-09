import Link from "next/link";

function Footer({ children }) {
  return (
    <footer>
      {children}

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
  );
}

export default Footer;
