import Link from "next/link";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faGithub, faTwitter } from "@fortawesome/free-brands-svg-icons";

function Footer({ children }) {
  return (
    <footer>
      {children}
      <div id="manifesto" className="body">
        <h2>Let’s build this together</h2>
        <div className="content">
          <p>
            Everyone uses version control for software, but it’s much less
            common in machine learning.
          </p>
          <p>
            This causes all sorts of problems: people are manually keeping track
            of things in spreadsheets, model weights are scattered on S3, and
            results can’t be reproduced. Somebody who wrote a model has left the
            team? Bad luck – nothing’s written down and you’ve probably got to
            start from scratch.
          </p>
          <p>
            So why isn’t everyone using Git?{" "}
            <strong>Git doesn’t work well with machine learning.</strong> It
            can’t handle large files, it can’t handle key/value metadata like
            metrics, and it can’t record information automatically from inside a
            training script. There are some solutions for these things, but they
            feel like band-aids.
          </p>
          <p>
            We spent a year talking to people in the ML community about this,
            and this is what we found out:
          </p>
          <ul>
            <li>
              <strong>We need a native version control system for ML.</strong>{" "}
              It’s sufficiently different to normal software that we can’t just
              put band-aids on existing systems.
            </li>
            <li>
              <strong>
                It needs to be small, easy to use, and extensible.
              </strong>{" "}
              We found people struggling to migrate to “AI Platforms”. We
              believe tools should do one thing well and combine with other
              tools to produce the system you need.
            </li>
            <li>
              <strong>It needs to be open source.</strong> There are a number of
              proprietary solutions, but something so foundational needs to be
              built by and for the ML community.
            </li>
          </ul>
          <p>
            We need your help to make this a reality. If you’ve built this for
            yourself, or are just interested in this problem, join us to help
            build a better system for everyone.
          </p>
          <p className="buttons">
            <a className="button" href="https://discord.gg/QmzJApGjyE">
              <strong>Join our Discord chat</strong>
            </a>
            &nbsp;&nbsp;or&nbsp;&nbsp;
            <a
              className="button"
              href="https://github.com/replicate/replicate#get-involved"
            >
              <strong>Get involved on GitHub</strong>
            </a>
          </p>
          <hr />
          <form
            action="https://works.us5.list-manage.com/subscribe/post?u=78112dc49af57b49c0d96fd04&amp;id=8c9d6302ee"
            method="post"
            target="_blank"
          >
            <h3>
              Sign up for occasional email updates about the project and the
              community:
            </h3>
            <fieldset>
              <input
                type="email"
                name="EMAIL"
                placeholder="Enter your email address"
              />
              <div
                style={{ position: "absolute", left: "-5000px" }}
                aria-hidden="true"
              >
                <input
                  type="text"
                  name="b_78112dc49af57b49c0d96fd04_8c9d6302ee"
                  tabIndex="-1"
                  value=""
                  readOnly
                />
              </div>
              <button type="submit">Sign up</button>
            </fieldset>
          </form>
        </div>
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
        <Link href="/">
          <a>
            <strong>Replicate</strong> Version control for machine&nbsp;learning
          </a>
        </Link>
      </p>
    </footer>
  );
}

export default Footer;
