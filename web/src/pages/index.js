import React from "react";
import classnames from "classnames";
import Layout from "@theme/Layout";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import CodeBlock from "@theme/CodeBlock";
import useBaseUrl from "@docusaurus/useBaseUrl";
import styles from "./styles.module.css";

function Feature({ imageUrl, title, description }) {
  const imgUrl = useBaseUrl(imageUrl);
  return (
    <div className={classnames("col col--4", styles.feature)}>
      {imgUrl && (
        <div className="text--center">
          <img className={styles.featureImage} src={imgUrl} alt={title} />
        </div>
      )}
      <h2>{title}</h2>
      {description}
    </div>
  );
}

function Home() {
  const context = useDocusaurusContext();
  const { siteConfig = {} } = context;
  return (
    <Layout title="" description="">
      <div
        style={{
          padding: "10px",
          backgroundColor: "#EEE",
          textAlign: "center",
        }}
      >
        <strong>Welcome to the Replicate private beta!</strong> It's ready to
        use for real work, but expect the odd rough edge.
        <br />
        If you know somebody else who would like to use it, email us:{" "}
        <a href="mailto:team@replicate.ai">team@replicate.ai</a>
      </div>
      <header className={classnames("hero", styles.heroBanner)}>
        <div className="container">
          <h1 className="hero__title">{siteConfig.title}</h1>
          <p className="hero__subtitle">
            Version control for machine learning experiments.
          </p>
          <div className={styles.buttons}>
            <Link
              className={classnames(
                "button button--secondary button--lg",
                styles.getStarted
              )}
              to={useBaseUrl("docs/tutorial")}
            >
              Get Started
            </Link>
          </div>
        </div>
      </header>
      <main>
        {/* Value propositions */}
        <section className={styles.features}>
          <div className="container">
            <div className="row">
              <Feature
                title={<>Never lose anything again</>}
                description={
                  <p>
                    On every training run, Replicate automatically saves
                    hyperparameters, code, training data, weights, metrics,
                    Python version, Python dependencies — <em>everything</em>.
                  </p>
                }
              />

              <Feature
                title={<>Go back in time</>}
                description={
                  <p>
                    You can get back the code and weights from any checkpoint if
                    you need to figure out how a model was trained or commit to
                    Git after the fact.
                  </p>
                }
              />

              <Feature
                title={<>Version your models</>}
                description={
                  <p>
                    All the model weights are versioned and stored on your own
                    Amazon S3 or Google Cloud bucket, so it's really easy to
                    feed them into production systems.
                  </p>
                }
              />
            </div>
          </div>
        </section>

        <div className="container padding-vert--lg">
          <div className="row padding-vert--lg">
            <div className={classnames(`col col--8`)}>
              <h1 className="" style={{ textAlign: "left" }}>
                How it works
              </h1>
              <CodeBlock className="python">
                {`import torch
import replicate

def train():
    # highlight-next-line
    # Save training code and hyperparameters
    # highlight-next-line
    experiment = replicate.init(path=".", params={...})
    model = Model()

    for epoch in range(params["num_epochs"]):
        # ...

        torch.save(model, "model.pth")
        # highlight-next-line
        # Save model weights and the metrics
        # highlight-next-line
        experiment.checkpoint(path="model.pth", metrics={...})`}
              </CodeBlock>
            </div>
          </div>
          {/* Differentiators and "yeah but" rebuttals */}
          <div className="row  padding-vert--lg">
            <Feature
              title={<>Don't change how you work</>}
              description={
                <p>
                  Just add two lines of code and Replicate will start keeping
                  track of everything.
                </p>
              }
            />
            <Feature
              title={<>You're in control of your data</>}
              description={
                <p>
                  All the data is stored on your own Amazon S3 or Google Cloud
                  Bucket as plain old files.
                </p>
              }
            />
            <Feature
              title={<>It's open source</>}
              description={
                <p>
                  It's not going to stop working if a startup goes out of
                  business. {/* lol we can work on this */}
                </p>
              }
            />
          </div>
        </div>
        {/* Use cases & "wow" moments */}
        <div className="container padding-vert--lg">
          <div className="row padding-vert--lg">
            <div className={classnames(`col col--12`)}>
              <h1 className="" style={{ textAlign: "center" }}>
                Now you can...
              </h1>
            </div>
          </div>
          <div className="row padding-vert--lg">
            <div className={classnames(`col col--6`)}>
              <h2 className="">Throw away your spreadsheet</h2>

              <p>
                Your experiments are all in one place, with filter and sort. The
                data's stored on S3, you can even see experiments that were run
                on other machines.
              </p>

              {/* TODO: animate this code block with filtering and sorting! */}

              <CodeBlock className="shell-session">
                {`$ replicate ls --filter "val_loss<0.2"
EXPERIMENT   HOST         STATUS    BEST CHECKPOINT
e510303      10.52.2.23   stopped   49668cb (val_loss=0.1484)
9e97e07      10.52.7.11   running   41f0c60 (val_loss=0.1989)`}
              </CodeBlock>
            </div>
            <div className={classnames(`col col--6`)}>
              <h2 className="">Compare experiments</h2>
              <p>
                It diffs everything, all the way down to versions of
                dependencies, just in case that latest Tensorflow version did
                something weird.
              </p>
              <CodeBlock className="shell-session">
                {`$ replicate diff 49668cb 41f0c60
Checkpoint:       49668cb     41f0c60
Experiment:       e510303     9e97e07

Params
learning_rate:    0.001       0.002

Metrics
train_loss:       0.4626      0.8155
train_accuracy:   0.7909      0.7254
val_loss:         0.1484      0.1989
val_accuracy:     0.9607      0.9411`}
              </CodeBlock>
            </div>
          </div>
          <div className="row padding-vert--lg">
            <div className={classnames(`col col--6`)}>
              <h2 className="">Commit to Git, after the fact</h2>
              <p>
                There's no need to carefully commit everything to Git. Replicate
                lets you get back to any point you called{" "}
                <code>experiment.checkpoint()</code>, so you can commit to Git
                once you've found something that works.
              </p>
              <CodeBlock className="shell-session">
                {`$ replicate checkout f81069d
Copying code and weights to working directory...

# save the code to git
$ git commit -am "Use hinge loss"`}
              </CodeBlock>
            </div>
            <div className={classnames(`col col--6`)}>
              <h2 className="">Run an experiment on a remote machine</h2>
              <p>
                Replicate has a handy shortcut for this. It'll SSH in, check the
                GPUs are set up correctly, copy over your code, then start your
                experiment inside Docker. All the results are stored on S3 so
                you can access them from your laptop.
              </p>
              <CodeBlock className="shell-session">
                {`$ replicate run -H 10.52.63.24 python train.py
═══╡ Checking 10.52.63.24 is set up correctly...
✔️ NVIDIA drivers 361.93 installed.
✔️ 2x NVIDIA K80 GPUs found.
✔️ CUDA 10.1 installed.
═══╡ Copying code...
═══╡ Building Docker image...
═══╡ Running 'python train.py' as experiment 35354d3...
`}
              </CodeBlock>
            </div>
          </div>
        </div>

        <div className="container padding-vert--lg">
          <div className="row padding-vert--lg">
            <div
              className={classnames(`col col--12`)}
              style={{ textAlign: "center" }}
            >
              <h3>Sound good?</h3>
              <Link
                className={classnames(
                  "button button--secondary button--lg margin-vert--md",
                  styles.getStarted
                )}
                to={useBaseUrl("docs/tutorial")}
              >
                Get Started
              </Link>
              <p>(It only takes 10 minutes.)</p>
            </div>
          </div>
        </div>
      </main>
    </Layout>
  );
}

export default Home;
