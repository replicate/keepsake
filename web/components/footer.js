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
            nothing is reproducible. It's hard enough getting your own model
            from a month ago running, let alone somebody else's.
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
            <strong>
              We want to make a small, lightweight, native version control
              system for ML.
            </strong>{" "}
            Something that does one thing well and combines with other tools to
            produce the system you need.
          </p>
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
              href="https://github.com/replicate/keepsake#get-involved"
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
      <p className="project-from">
        A project from <a href="https://replicate.ai">Replicate</a>.
      </p>
      <nav>
        <a href="https://github.com/replicate/keepsake">GitHub</a>
        <a href="https://twitter.com/replicateai">Twitter</a>
        <a href="mailto:team@replicate.ai">team@replicate.ai</a>
      </nav>
    </footer>
  );
}

export default Footer;
