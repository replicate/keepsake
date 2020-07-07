package main

import (
	"replicate.ai/cli/pkg/cli"
	"replicate.ai/cli/pkg/console"
)

func main() {
	cmd, err := cli.NewRootCommand()
	if err != nil {
		console.Fatal("%s", err)
	}

	if err = cmd.Execute(); err != nil {
		console.Fatal("%s", err)
	}
}
