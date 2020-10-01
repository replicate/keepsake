import CodeBlock from "../components/code-block";
import Header from "../components/header";
import Layout from "../layouts/default";

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
        {/* <p>
          Lightweight and open source. <a href="/docs">Get started</a> or{" "}
          <a href="https://github.com/replicate/replicate">view on GitHub</a>
        </p> */}
        <p>
          Lightweight and open source.
          <a className="button" href="/docs">
            Get started
          </a>
        </p>
        <section className="info">
          <div>
            <h2>
              <span>{num()}</span> Never lose anything
            </h2>
            <p>
              On every training run, Replicate automatically saves
              hyperparameters, code, training data, weights, metrics, Python
              version, Python dependencies — <em>everything</em>.
            </p>
          </div>
          <div>
            <h2>
              <span>{num()}</span> Go back in time
            </h2>
            <p>
              You can get back the code and weights from any checkpoint if you
              need to figure out how a model was trained or commit to Git after
              the fact.
            </p>
          </div>
          <div>
            <h2>
              <span>{num()}</span> Version your models
            </h2>
            <p>
              All the model weights are versioned and stored on your own Amazon
              S3 or Google Cloud bucket, so it's really easy to feed them into
              production systems.
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
        # Save a model weights and the metrics
        experiment.checkpoint(path="model.pth", metrics={...})
        #highlight-end`}</CodeBlock>
      </section>

      <section className="info">
        <div>
          <h2>
            <span>{num()}</span> You’re in control of your data
          </h2>
          <p>
            All the data is stored on your own Amazon S3 or Google Cloud bucket
            as plain old files.
          </p>
        </div>
        <div>
          <h2>
            <span>{num()}</span> It’s open source
          </h2>
          <p>
            It's not going to stop working if a startup goes out of business.
          </p>
        </div>
      </section>

      <section className="features">
        <nav>
          <ol>
            <li>
              <h2>
                <span>{num()}</span> Features
              </h2>
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

Metrics
train_loss:       0.4626      0.8155
train_accuracy:   0.7909      0.7254
val_loss:         0.1484      0.1989
val_accuracy:     0.9607      0.9411`}
          </CodeBlock>

          <h3 id="anchor-3">Commit to Git, after the fact</h3>
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

          <h3 id="anchor-4">Run an experiment on a remote machine</h3>
          <p>
            Replicate has a handy shortcut for this. It'll SSH in, check the
            GPUs are set up correctly, copy over your code, then start your
            experiment inside Docker. All the results are stored on S3 so you
            can access them from your laptop.
          </p>
          <CodeBlock language="shell-session">
            {`$ replicate run -H 10.52.63.24 python train.py
═══╡ Checking 10.52.63.24 is set up correctly...
✔️ NVIDIA drivers 361.93 installed.
✔️ 2x NVIDIA K80 GPUs found.
✔️ CUDA 10.1 installed.
═══╡ Copying code...
═══╡ Building Docker image...
═══╡ Running 'python train.py' as experiment 35354d3...`}
          </CodeBlock>
        </div>
      </section>
      <footer>
        <h2>
          <div>
            <a className="button" href="/docs">
              Get started
            </a>
          </div>
          <div> or, </div>
          <div>
            <a href="/docs/learn/how-it-works">
              learn more about how Replicate works
            </a>
          </div>
        </h2>
      </footer>
    </Layout>
  );
}
