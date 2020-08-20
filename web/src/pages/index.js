import React from "react";
import classnames from "classnames";
import Layout from "@theme/Layout";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import CodeBlock from "@theme/CodeBlock";
import useBaseUrl from "@docusaurus/useBaseUrl";
import styles from "./styles.module.css";

const features = [
  {
    title: (
      <>
        Track <em>everything</em>
      </>
    ),
    imageUrl: "",
    description: (
      <>
        <p>
          Code, training data, hyperparameters, weights, metrics, Tensorboard
          logs, Python version, Python dependencies – <em>everything</em>.
        </p>
        <p>
          You'll never forget how a model was trained, and you'll always be able
          to retrain it in the future.
        </p>
      </>
    ),
  },
  {
    title: <>Don't change how you work</>,
    imageUrl: "",
    description: (
      <>
        <p>
          You don't need to learn a new workflow or remember to commit anything.
          Just work the way you want, and Replicate keeps track of it.
        </p>
        {/* <p>
          Add <code>experiment = replicate.init()</code> to your code to start
          an experiment, then call <code>experiment.commit()</code> to save the
          exact state of your training at that point. That's it.
        </p> */}
      </>
    ),
  },
  {
    title: <>Works remotely out of the box</>,
    imageUrl: "",
    description: (
      <>
        <p>
          We figure you'll do most of your training on GPU machines or a cloud
          service. Replicate makes that really simple, all controlled from your
          laptop.
        </p>
        {/* <p>
          Just point Replicate at any machine with SSH installed, prefix your
          training script with `replicate run`, and Replicate handles the rest.
          All your data is stored on S3.
        </p> */}
      </>
    ),
  },
];

const replicateYamlSnippet = `
`;

function Feature({ imageUrl, title, description }) {
  const imgUrl = useBaseUrl(imageUrl);
  return (
    <div className={classnames("col col--4", styles.feature)}>
      {imgUrl && (
        <div className="text--center">
          <img className={styles.featureImage} src={imgUrl} alt={title} />
        </div>
      )}
      <h3>{title}</h3>
      {description}
    </div>
  );
}

function Home() {
  const context = useDocusaurusContext();
  const { siteConfig = {} } = context;
  return (
    <Layout title="" description="">
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
        {features && features.length && (
          <section className={styles.features}>
            <div className="container">
              <div className="row">
                {features.map((props, idx) => (
                  <Feature key={idx} {...props} />
                ))}
              </div>
            </div>
          </section>
        )}

        <div className="container padding-vert--lg">
          <div className="row padding-vert--lg">
            <div className={classnames(`col col--12`)}>
              <h1 className="" style={{ textAlign: "center" }}>
                How it works
              </h1>
            </div>
          </div>
          <div className="row padding-vert--lg">
            <div className={classnames(`col col--6`)}>
              <h2>Step 1: Add Replicate to your training code</h2>
              <CodeBlock className="python">
                {`import torch
import replicate

def train(**params):
    # highlight-next-line
    experiment = replicate.init(**params)
    model = Model()

    for epoch in range(params["num_epochs"]):
        # ...

        torch.save(model, "model.torch")
        # highlight-next-line
        experiment.commit(path="model.torch", **metrics)`}
              </CodeBlock>
            </div>
            <div className={classnames(`col col--6`)}>
              <h2>Step 2: Get back to work</h2>
              <p>
                That's it. Everything about your experiments is now saved
                forever:
              </p>
              <ul>
                <li>Hyperparameters</li>
                <li>Metrics</li>
                <li>Weight files, Tensorboard logs, etc</li>
                <li>Code, even if you didn't commit to Git</li>
                <li>Training data (with a bit of extra configuration)</li>
                <li>
                  Python version, Python requirements, PyTorch/Tensorflow
                  version
                </li>
              </ul>
              <p>
                By default this data is stored on your local filesystem, but you
                can also store it centrally on S3 or Google Cloud Storage.
              </p>
            </div>
          </div>
        </div>
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
              <h2 className="">See all your experiments in one place</h2>

              <p>
                If you store your data on S3, you can even see experiments that
                were run on other machines.
              </p>

              <CodeBlock className="shell-session">
                {`$ replicate ls
EXPERIMENT   HOST         STATUS    BEST COMMIT
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
Commit:           49668cb     41f0c60
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
                <code>experiment.commit()</code>, so you can commit to Git once
                you've found something that works.
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
---| Checking 10.52.63.24 is set up correctly...
✔️ NVIDIA drivers 361.93 installed.
✔️ 2x NVIDIA K80 GPUs found.
✔️ CUDA 10.1 installed.
---| Copying code...
---| Building Docker image...
---| Running 'python train.py' as experiment 35354d3...
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
