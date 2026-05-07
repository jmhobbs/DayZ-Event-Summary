package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jmhobbs/dayz-event-summary/internal/cli/root"
)

var version string = "v0.0.0-dev"

func main() {
	command := root.New(root.Options{
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Version: version,
	})

	if err := command.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "dayz-event-summary: %v\n", err)
		os.Exit(1)
	}
}
