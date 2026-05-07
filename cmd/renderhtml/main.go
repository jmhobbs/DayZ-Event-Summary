package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jmhobbs/dayz-event-summary/internal/renderhtml"
	"github.com/jmhobbs/dayz-event-summary/internal/runconfig"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "renderhtml: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	options, err := parseOptions(args)
	if err != nil {
		return err
	}

	config, err := runconfig.LoadFile(options.ConfigPath)
	if err != nil {
		return err
	}

	dataDir, outputDir := resolvePaths(options.ConfigPath, config, options.DataDir, options.OutputDir)
	bundle, err := renderhtml.LoadBundle(dataDir)
	if err != nil {
		return err
	}

	return renderhtml.Render(bundle, outputDir)
}

type options struct {
	ConfigPath string
	DataDir    string
	OutputDir  string
}

func parseOptions(args []string) (options, error) {
	var parsed options
	flags := flag.NewFlagSet("renderhtml", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&parsed.ConfigPath, "config", "", "Path to config.yaml")
	flags.StringVar(&parsed.DataDir, "data-dir", "", "Optional event data directory override")
	flags.StringVar(&parsed.OutputDir, "out", "", "Optional output directory override")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if parsed.ConfigPath == "" {
		return options{}, fmt.Errorf("--config is required")
	}

	return parsed, nil
}

func resolvePaths(configPath string, config runconfig.Config, dataOverride string, outputOverride string) (string, string) {
	dataDir := runconfig.ResolvePath(configPath, config.DataDir)
	if dataOverride != "" {
		dataDir = dataOverride
	}

	outputDir := runconfig.ResolvePath(configPath, config.HTMLDir)
	if outputOverride != "" {
		outputDir = outputOverride
	}

	return dataDir, outputDir
}
