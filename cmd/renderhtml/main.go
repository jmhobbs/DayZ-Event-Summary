package main

import (
	"flag"
	"fmt"
	"os"

	"amd-report/internal/renderhtml"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "renderhtml: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var dataDir string
	var outputDir string

	flag.StringVar(&dataDir, "data-dir", "", "Path to the event data directory")
	flag.StringVar(&outputDir, "out", "", "Output directory for rendered HTML")
	flag.Parse()

	if dataDir == "" || outputDir == "" {
		return fmt.Errorf("--data-dir and --out are required")
	}

	bundle, err := renderhtml.LoadBundle(dataDir)
	if err != nil {
		return err
	}

	return renderhtml.Render(bundle, outputDir)
}
