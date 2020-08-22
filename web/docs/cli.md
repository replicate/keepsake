---
id: cli
title: CLI reference
---

This is the reference for the Replicate CLI commands. You can also see this in the terminal by running `replicate --help` or `replicate command --help`.

## Commands

* [`replicate checkout`](#replicate-checkout) – Copy files from a commit into the project directory
* [`replicate diff`](#replicate-diff) – Compare two experiments or commits
* [`replicate feedback`](#replicate-feedback) – Submit feedback to the team!
* [`replicate ls`](#replicate-ls) – List experiments in this project
* [`replicate ps`](#replicate-ps) – List running experiments in this project
* [`replicate rm`](#replicate-rm) – Remove experiments or commits
* [`replicate run`](#replicate-run) – Run a command on a remote machine
* [`replicate show`](#replicate-show) – View information about an experiment or commit

## `replicate checkout`

Copy files from a commit into the project directory

### Usage

```
replicate checkout <commit-id> [flags]
```

### Flags

```
  -f, --force                     Force checkout without prompt, even if the directory is not empty
  -h, --help                      help for checkout
  -o, --output-directory string   Output directory (defaults to working directory or directory with replicate.yaml in it)
  -S, --storage-url string        Storage URL (e.g. 's3://my-replicate-bucket' (if omitted, uses storage URL from replicate.yaml)

      --color                     Display color in output (default true)
  -D, --source-directory string   Local source directory
  -v, --verbose                   Verbose output
```
## `replicate diff`

Compare two experiments or commits.

If an experiment ID is passed, it will pick the best commit from that experiment. If a primary metric is not defined in replicate.yaml, it will use the latest commit.

### Usage

```
replicate diff <id> <id> [flags]
```

### Flags

```
  -h, --help                 help for diff
  -S, --storage-url string   Storage URL (e.g. 's3://my-replicate-bucket' (if omitted, uses storage URL from replicate.yaml)

      --color                     Display color in output (default true)
  -D, --source-directory string   Local source directory
  -v, --verbose                   Verbose output
```
## `replicate feedback`

Submit feedback to the team!

### Usage

```
replicate feedback [flags]
```

### Flags

```
  -h, --help   help for feedback

      --color                     Display color in output (default true)
  -D, --source-directory string   Local source directory
  -v, --verbose                   Verbose output
```
## `replicate ls`

List experiments in this project

### Usage

```
replicate ls [flags]
```

### Flags

```
  -p, --all-params           Output all experiment params (by default, outputs only parameters that change between experiments)
  -f, --filter stringArray   Filters (format: "<name> <operator> <value>")
  -h, --help                 help for ls
      --json                 Print output in JSON format
  -q, --quiet                Only print experiment IDs
  -s, --sort string          Sort key. Suffix with '-desc' for descending sort, e.g. --sort=started-desc (default "started")
  -S, --storage-url string   Storage URL (e.g. 's3://my-replicate-bucket' (if omitted, uses storage URL from replicate.yaml)

      --color                     Display color in output (default true)
  -D, --source-directory string   Local source directory
  -v, --verbose                   Verbose output
```
## `replicate ps`

List running experiments in this project

### Usage

```
replicate ps [flags]
```

### Flags

```
  -p, --all-params           Output all experiment params (by default, outputs only parameters that change between experiments)
  -f, --filter stringArray   Filters (format: "<name> <operator> <value>")
  -h, --help                 help for ps
      --json                 Print output in JSON format
  -q, --quiet                Only print experiment IDs
  -s, --sort string          Sort key. Suffix with '-desc' for descending sort, e.g. --sort=started-desc (default "started")
  -S, --storage-url string   Storage URL (e.g. 's3://my-replicate-bucket' (if omitted, uses storage URL from replicate.yaml)

      --color                     Display color in output (default true)
  -D, --source-directory string   Local source directory
  -v, --verbose                   Verbose output
```
## `replicate rm`

Remove experiments or commits.

To remove experiments or commits, pass any number of IDs (or prefixes).


### Usage

```
replicate rm <experiment-or-commit-id> [experiment-or-commit-id...] [flags]
```

### Flags

```
  -h, --help                 help for rm
  -S, --storage-url string   Storage URL (e.g. 's3://my-replicate-bucket' (if omitted, uses storage URL from replicate.yaml)

      --color                     Display color in output (default true)
  -D, --source-directory string   Local source directory
  -v, --verbose                   Verbose output
```
## `replicate run`

Run a command on a remote machine

### Usage

```
replicate run [flags] <command> [arg...]
```

### Flags

```
  -h, --help                   help for run
  -H, --host string            SSH host to run command on, in form [username@]hostname[:port]
  -i, --identity-file string   SSH private key path
  -m, --mount stringArray      Mount directories from the host to Replicate's Docker container. Format: --mount <host-directory>:<container-directory>

      --color                     Display color in output (default true)
  -D, --source-directory string   Local source directory
  -v, --verbose                   Verbose output
```
## `replicate show`

View information about an experiment or commit

### Usage

```
replicate show <experiment-or-commit-id> [flags]
```

### Flags

```
  -h, --help                 help for show
  -S, --storage-url string   Storage URL (e.g. 's3://my-replicate-bucket' (if omitted, uses storage URL from replicate.yaml)

      --color                     Display color in output (default true)
  -D, --source-directory string   Local source directory
  -v, --verbose                   Verbose output
```
