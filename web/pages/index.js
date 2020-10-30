import CodeBlock from "../components/code-block";
import Header from "../components/header";
import Layout from "../layouts/default";
import Link from "next/link";

export default function Home() {
  let _num = 1;

  function num() {
    let s = _num + "";
    _num += 1;
    if (s.length < 2) {
      s = "0" + s;
    }
    return s;
  }

  return (
    <Layout title="Replicate – Version control for machine learning">
      <Header className="homepage">
        <p>
          <Link href="/docs">
            <a className="button">Get Started</a>
          </Link>{" "}
          Lightweight and open source.{" "}
        </p>
        <section className="info">
          <div>
            <h2>
              <span>{num()}</span> Track experiments
            </h2>
            <p>
              Automatically track code, hyperparameters, training data, weights,
              metrics, Python dependencies — <em>everything</em>.
            </p>
          </div>
          <div>
            <h2>
              <span>{num()}</span> Go back in time
            </h2>
            <p>
              Get back the code and weights from any checkpoint if you need to
              replicate your results or commit to Git after the fact.
            </p>
          </div>
          <div>
            <h2>
              <span>{num()}</span> Version your models
            </h2>
            <p>
              Model weights are stored on your own Amazon S3 or Google Cloud
              bucket, so it's really easy to feed them into production systems.
            </p>
          </div>
        </section>
      </Header>

      <section className="terminal">
        <div>
          <h2>
            <span>{num()}</span> How it works
          </h2>
          <p>
            Just add two lines of code and Replicate will start keeping track of
            everything. You don't need to change how you work.
          </p>
        </div>
        <div className="windowChrome">
          <CodeBlock language="python">{`import torch
import replicate

def train():
    #highlight-start
    # Save training code and hyperparameters
    experiment = replicate.init(path=".", params={...})
    #highlight-end
    model = Model()

    for epoch in range(num_epochs):
        # ...

        torch.save(model, "model.pth")
        #highlight-start
        # Save model weights and metrics
        experiment.checkpoint(path="model.pth", metrics={...})
        #highlight-end`}</CodeBlock>
        </div>
      </section>

      <section className="control">
        <div>
          <h2>
            <span>{num()}</span> Open source
          </h2>
          <p>It won't stop working if a startup goes out of business.</p>
        </div>
        <div>
          <h2>
            <span>{num()}</span> You're in control of&nbsp;your&nbsp;data
          </h2>
          <p>
            All the data is stored on your own Amazon S3 or Google Cloud Storage
            as plain old files. There's no server to run.
          </p>
        </div>
        <div>
          <h2>
            <span>{num()}</span> Works with everything
          </h2>
          <p>
            Tensorflow, PyTorch, scikit-learn, XGBoost, you name it. It's just
            saving files and dictionaries – export however you want.
          </p>
        </div>
      </section>

      <section className="docs homepage">
        <nav>
          <ol>
            <li>
              <h2>
                <span>{num()}</span> Features
              </h2>
              <ol>
                <li>
                  <a href="#anchor-1">Throw away your spreadsheet</a>
                </li>
                <li>
                  <a href="#anchor-2">Compare experiments</a>
                </li>
                <li>
                  <a href="#anchor-3">Analyze in a notebook</a>
                </li>
                <li>
                  <a href="#anchor-4">Commit to Git, after the fact</a>
                </li>
                <li>
                  <a href="#anchor-5">Load models in production</a>
                </li>
                <li>
                  <a href="#anchor-6">A platform to build upon</a>
                </li>
              </ol>
            </li>
          </ol>
        </nav>
        <div className="body">
          <h3 id="anchor-1">Throw away your spreadsheet</h3>
          <p>
            Your experiments are all in one place, with filter and sort. Because
            the data's stored on S3, you can even see experiments that were run
            on other machines.
          </p>
          <CodeBlock language="shell-session">
            {`$ replicate ls --filter "val_loss<0.2"
EXPERIMENT   HOST         STATUS    BEST CHECKPOINT
e510303      10.52.2.23   stopped   49668cb (val_loss=0.1484)
9e97e07      10.52.7.11   running   41f0c60 (val_loss=0.1989)`}
          </CodeBlock>

          <h3 id="anchor-3">Analyze in a notebook</h3>
          <p>
            Don't like the CLI? No problem. You can retrieve, analyze, and plot
            your results from within a notebook. Think of it like a programmable
            Tensorboard.
          </p>
          <img src="images/notebook.png" width="900" />

          <h3 id="anchor-2">Compare experiments</h3>
          <p>
            It diffs everything, all the way down to versions of dependencies,
            just in case that latest Tensorflow version did something weird.
          </p>
          <CodeBlock language="shell-session">
            {`$ replicate diff 49668cb 41f0c60
Checkpoint:       49668cb     41f0c60
Experiment:       e510303     9e97e07

Params
learning_rate:    0.001       0.002

Python Packages
tensorflow:       2.3.0       2.3.1

Metrics
train_loss:       0.4626      0.8155
train_accuracy:   0.7909      0.7254
val_loss:         0.1484      0.1989
val_accuracy:     0.9607      0.9411`}
          </CodeBlock>

          <h3 id="anchor-4">Commit to Git, after the fact</h3>
          <p>
            There's no need to carefully commit everything to Git. Replicate
            lets you get back to any point you called{" "}
            <code>experiment.checkpoint()</code>, so you can commit to Git once
            you've found something that works.
          </p>
          <CodeBlock language="shell-session">
            {`$ replicate checkout f81069d
Copying code and weights to working directory...

# save the code to git
$ git commit -am "Use hinge loss"`}
          </CodeBlock>

          <h3 id="anchor-5">Load models in production</h3>
          <p>
            You can use Replicate to feed your models into production systems.
            Connect them back to how they were trained, who trained them, and
            what their metrics were.
          </p>
          <CodeBlock language="python">
            {`import replicate
model = torch.load(replicate.checkpoints.get("e45a203").open("model.pth"))`}
          </CodeBlock>

          <h3 id="anchor-6">A platform to build upon</h3>
          <p>
            Replicate is intentionally lightweight and doesn't try to do too
            much. Instead, we give you Python and command-line APIs so you can
            integrate it with your own tools and workflow.
          </p>
        </div>
      </section>
    </Layout>
  );
}
