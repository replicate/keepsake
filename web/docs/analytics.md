---
id: analytics
title: Analytics
---

The Replicate CLI sends anonymous analytics about commands you run. You will be notified the first time you run `replicate`.

## Why we do this

It helps us:

1. Understand what commands people use and what's failing so we can prioritize work.
2. Understand what operating systems, Python, and Tensorflow/PyTorch versions people use so we can figure out what ones to support.
3. Figure out roughly how many people use Replicate and what its growth rate is. We intend to build a business to support the open-source project, and this helps us get investment and so on. (This is the unspoken truth as to why most things gather analytics, but we want to be honest with you.)

## What data is sent

- A random token for the machine (e.g. `9f3027bb-0eb8-917d-e5bf-c6c1bdb1fd0a`)
- The subcommand you ran, without any options or arguments (e.g. `replicate run`, not `replicate run python secretproject.py`)
- The Replicate version (e.g. `1.0.0`)
- Your CPU architecture (e.g. `amd64`)
- Your operating system (e.g. `linux`)

## Where it's sent

Data is sent to Segment, then Mixpanel. We're figuring out what our retention policy will be.

## Opting out

These analytics really help us, and we'd really appreciate it if you left it on. But, if you want to opt out, you can set this environment variable:

```
export REPLICATE_NO_ANALYTICS=1
```

Or, run this command:

```
replicate analytics off
```
