package main

import (
	"github.com/replicate/replicate/go/pkg/cli"
	"github.com/replicate/replicate/go/pkg/console"
)

func main() {
	cmd := cli.NewDaemonCommand()
	if err := cmd.Execute(); err != nil {
		console.Fatal("%s", err)
	}
}
