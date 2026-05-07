package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"amd-report/internal/eventbuild"
	"amd-report/internal/runconfig"
	"amd-report/internal/teamguess"
)

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "eventbuild: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stderr io.Writer) error {
	options, err := parseOptions(args)
	if err != nil {
		return err
	}

	config, err := runconfig.LoadFile(options.ConfigPath)
	if err != nil {
		return err
	}

	paths := resolvePaths(options.ConfigPath, config, options.TeamConfigPath, options.OutputDir)
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
	defer admFile.Close()

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
		fmt.Fprintln(stderr, warning)
	}

	return eventbuild.WriteBundle(bundle, paths.OutputDir)
}

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

func parseOptions(args []string) (options, error) {
	var parsed options
	flags := flag.NewFlagSet("eventbuild", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&parsed.ConfigPath, "config", "", "Path to config.yaml")
	flags.StringVar(&parsed.TeamConfigPath, "teams", "", "Optional path to teams.yaml")
	flags.StringVar(&parsed.OutputDir, "out", "", "Optional output directory override")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
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

func loadTeamConfig(path string) (eventbuild.TeamConfig, error) {
	if path == "" {
		return eventbuild.TeamConfig{}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return eventbuild.TeamConfig{}, fmt.Errorf("open team config: %w", err)
	}
	defer file.Close()

	teams, err := eventbuild.LoadTeamConfig(file)
	if err != nil {
		return eventbuild.TeamConfig{}, err
	}

	return teams, nil
}
