import Link from "next/link";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faGithub, faTwitter } from "@fortawesome/free-brands-svg-icons";

function Footer({ children }) {
  return (
    <footer>
      {children}
      <div id="manifesto" className="body">
        <h2>Let's build together</h2>
        <p>
          Everyone uses version control for software, but it's much less
          common in machine learning. Why is this?
        </p>
        <p>
          We spent a year talking to people in the ML community
          and this is what we found out:
        </p>
        <ul>
          <li>
            <strong>Git doesn’t work well with machine learning.</strong> It
            can’t handle large files, it can’t handle key/value metadata like
            metrics, and it can’t commit automatically in your training
            script. There are some solutions for this, but they feel like
            band-aids.
          </li>
          <li>
            <strong>It should be open source.</strong> There are a number of
            proprietary solutions, but something so foundational needs to be
            built by and for the ML community.
          </li>
          <li>
            <strong>
              It needs to be small, easy to use, and extensible.
            </strong>{" "}
            We found people struggling to integrate with “AI Platforms”. We
            want to make a tool that does one thing well and can be combined
            with other tools to produce the system you need.
          </li>
        </ul>
        <p>
          We think the ML community needs a good version control system. But,
          version control systems are complex, and to make this a reality we
          need your help.
        </p>
        <p>
          Have you strung together some shell scripts to build this for
          yourself? Are you interested in the problem of making machine
          learning reproducible?
        </p>
        <p>
          <a class="button" href="https://discord.gg/QmzJApGjyE">
            <strong>Join our Discord</strong>
          </a>
          &nbsp;&nbsp;or&nbsp;&nbsp;
          <a class="button" href="https://github.com/replicate/replicate#get-involved">
            <strong>Get involved on GitHub</strong>
          </a>
        </p>
      </div>
      <div id="contributors">
        <h3>Core team</h3>
        <div className="us">
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
                <a href="https://github.com/andreasjansson" className="link">
                  <FontAwesomeIcon icon={faGithub} />
                </a>
              </p>
            </figcaption>
          </figure>
        </div>
        <div className="more">
          <p>
            We also built{" "}
            <a href="https://www.arxiv-vanity.com/" target="_blank">
              arXiv Vanity
            </a>
            , which lets you read arXiv papers as responsive web pages.
          </p>
        </div>
      </div>
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
