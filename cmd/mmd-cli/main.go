package main

import (
	"fmt"
	"os"

	"github.com/coolamit/mermaid-cli/internal/cli"
)

func main() {
	cmd := cli.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\033[31m\n%s\n\033[0m", err.Error())
		os.Exit(1)
	}
}
