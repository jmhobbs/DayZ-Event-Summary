package rendercmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/jmhobbs/dayz-event-summary/internal/renderhtml"
	"github.com/jmhobbs/dayz-event-summary/internal/runconfig"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type options struct {
	ConfigPath string
	DataDir    string
	OutputDir  string
}

func New(stderr io.Writer) *ffcli.Command {
	var parsed options
	flags := newFlagSet(stderr, &parsed)

	return &ffcli.Command{
		Name:       "render",
		ShortUsage: "dayz-event-summary render --config <path/to/config.yaml>",
		ShortHelp:  "Render a static HTML report from the JSON bundle.",
		FlagSet:    flags,
		Exec: func(_ context.Context, _ []string) error {
			validated, err := validateOptions(parsed)
			if err != nil {
				return err
			}

			return run(validated)
		},
	}
}

func run(parsed options) error {
	config, err := runconfig.LoadFile(parsed.ConfigPath)
	if err != nil {
		return err
	}

	dataDir, outputDir := resolvePaths(parsed.ConfigPath, config, parsed.DataDir, parsed.OutputDir)
	bundle, err := renderhtml.LoadBundle(dataDir)
	if err != nil {
		return err
	}

	return renderhtml.Render(bundle, outputDir)
}

func parseOptions(args []string) (options, error) {
	var parsed options
	flags := newFlagSet(os.Stderr, &parsed)
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}

	return validateOptions(parsed)
}

func newFlagSet(stderr io.Writer, parsed *options) *flag.FlagSet {
	flags := flag.NewFlagSet("render", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&parsed.ConfigPath, "config", "", "Path to config.yaml")
	flags.StringVar(&parsed.DataDir, "data-dir", "", "Optional event data directory override")
	flags.StringVar(&parsed.OutputDir, "out", "", "Optional output directory override")
	return flags
}

func validateOptions(parsed options) (options, error) {
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
