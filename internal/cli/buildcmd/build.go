package buildcmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jmhobbs/dayz-event-summary/internal/eventbuild"
	"github.com/jmhobbs/dayz-event-summary/internal/runconfig"
	"github.com/jmhobbs/dayz-event-summary/internal/teamguess"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type options struct {
	ConfigPath     string
	TeamConfigPath string
	OutputDir      string
}

type paths struct {
	LogFilePath    string
	TeamConfigPath string
	OutputDir      string
}

func New(stderr io.Writer) *ffcli.Command {
	var parsed options
	flags := newFlagSet(stderr, &parsed)

	return &ffcli.Command{
		Name:       "build",
		ShortUsage: "dayz-event-summary build --config <path/to/config.yaml>",
		ShortHelp:  "Generate the JSON event bundle from config.yaml.",
		FlagSet:    flags,
		Exec: func(_ context.Context, _ []string) error {
			validated, err := validateOptions(parsed)
			if err != nil {
				return err
			}

			return run(validated, stderr)
		},
	}
}

func run(parsed options, stderr io.Writer) (err error) {
	config, err := runconfig.LoadFile(parsed.ConfigPath)
	if err != nil {
		return err
	}

	paths := resolvePaths(parsed.ConfigPath, config, parsed.TeamConfigPath, parsed.OutputDir)
	start, err := teamguess.ParseClock(config.Start)
	if err != nil {
		return fmt.Errorf("parse start: %w", err)
	}
	end, err := teamguess.ParseClock(config.End)
	if err != nil {
		return fmt.Errorf("parse end: %w", err)
	}

	admFile, err := os.Open(paths.LogFilePath)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer func() {
		if closeErr := admFile.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close log file: %w", closeErr))
		}
	}()

	teams, err := loadTeamConfig(paths.TeamConfigPath)
	if err != nil {
		return err
	}

	bundle, err := eventbuild.Build(admFile, filepath.Base(paths.LogFilePath), eventbuild.BuildOptions{
		Window:      teamguess.Window{Start: start, End: end},
		WindowStart: config.Start,
		WindowEnd:   config.End,
		Settings: eventbuild.EventSettings{
			EventName:           config.EventName,
			AssistWindowSeconds: config.AssistWindowSeconds,
		},
		Teams: teams,
	})
	if err != nil {
		return err
	}

	for _, warning := range bundle.Warnings {
		if _, err := fmt.Fprintln(stderr, warning); err != nil {
			return fmt.Errorf("write warning: %w", err)
		}
	}

	return eventbuild.WriteBundle(bundle, paths.OutputDir)
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
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&parsed.ConfigPath, "config", "", "Path to config.yaml")
	flags.StringVar(&parsed.TeamConfigPath, "teams", "", "Optional path to teams.yaml")
	flags.StringVar(&parsed.OutputDir, "out", "", "Optional output directory override")
	return flags
}

func validateOptions(parsed options) (options, error) {
	if parsed.ConfigPath == "" {
		return options{}, fmt.Errorf("--config is required")
	}

	return parsed, nil
}

func resolvePaths(configPath string, config runconfig.Config, teamOverride string, outputOverride string) paths {
	resolved := paths{
		LogFilePath: runconfig.ResolvePath(configPath, config.LogFile),
		OutputDir:   runconfig.ResolvePath(configPath, config.DataDir),
	}
	if outputOverride != "" {
		resolved.OutputDir = outputOverride
	}
	if teamOverride != "" {
		resolved.TeamConfigPath = teamOverride
		return resolved
	}
	if config.TeamsEvent {
		resolved.TeamConfigPath = filepath.Join(filepath.Dir(configPath), "teams.yaml")
	}

	return resolved
}

func loadTeamConfig(path string) (_ eventbuild.TeamConfig, err error) {
	if path == "" {
		return eventbuild.TeamConfig{}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return eventbuild.TeamConfig{}, fmt.Errorf("open team config: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close team config: %w", closeErr))
		}
	}()

	teams, err := eventbuild.LoadTeamConfig(file)
	if err != nil {
		return eventbuild.TeamConfig{}, err
	}

	return teams, nil
}
