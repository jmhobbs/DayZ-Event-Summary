package generateteamscmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/jmhobbs/dayz-event-summary/internal/generateteams"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type options struct {
	ConfigPath string
	OutputPath string
}

func New(stderr io.Writer) *ffcli.Command {
	var parsed options
	flags := newFlagSet(stderr, &parsed)

	return &ffcli.Command{
		Name:       "generate-teams",
		ShortUsage: "dayz-event-summary generate-teams --config <path/to/config.yaml>",
		ShortHelp:  "Generate teams.yaml suggestions from the event log and config window.",
		FlagSet:    flags,
		Exec: func(_ context.Context, _ []string) error {
			validated, err := validateOptions(parsed)
			if err != nil {
				return err
			}

			return generateteams.Run(validated.ConfigPath, validated.OutputPath)
		},
	}
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
	flags := flag.NewFlagSet("generate-teams", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&parsed.ConfigPath, "config", "", "Path to config.yaml")
	flags.StringVar(&parsed.OutputPath, "out", "", "Optional path to teams.yaml")
	return flags
}

func validateOptions(parsed options) (options, error) {
	if parsed.ConfigPath == "" {
		return options{}, fmt.Errorf("--config is required")
	}

	return parsed, nil
}
