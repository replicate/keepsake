package main

import (
	"github.com/replicate/keepsake/go/pkg/cli"
	"github.com/replicate/keepsake/go/pkg/console"
)

func main() {
	cmd := cli.NewDaemonCommand()
	if err := cmd.Execute(); err != nil {
		console.Fatal("%s", err)
	}
}
