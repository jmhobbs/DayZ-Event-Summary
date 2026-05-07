package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"amd-report/internal/eventbuild"
	"amd-report/internal/teamguess"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "eventbuild: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var admPath string
	var startValue string
	var endValue string
	var eventConfigPath string
	var teamConfigPath string
	var outputDir string

	flag.StringVar(&admPath, "adm", "", "Path to the ADM file")
	flag.StringVar(&startValue, "start", "", "Event start time (HH:MM:SS or RFC3339)")
	flag.StringVar(&endValue, "end", "", "Event end time (HH:MM:SS or RFC3339)")
	flag.StringVar(&eventConfigPath, "event-config", "", "Path to the event settings YAML")
	flag.StringVar(&teamConfigPath, "teams", "", "Path to the reviewed team YAML")
	flag.StringVar(&outputDir, "out", "", "Output directory for the JSON bundle")
	flag.Parse()

	if admPath == "" || startValue == "" || endValue == "" || eventConfigPath == "" || teamConfigPath == "" || outputDir == "" {
		return fmt.Errorf("--adm, --start, --end, --event-config, --teams, and --out are required")
	}

	start, err := teamguess.ParseClock(startValue)
	if err != nil {
		return fmt.Errorf("parse --start: %w", err)
	}
	end, err := teamguess.ParseClock(endValue)
	if err != nil {
		return fmt.Errorf("parse --end: %w", err)
	}

	admFile, err := os.Open(admPath)
	if err != nil {
		return fmt.Errorf("open ADM file: %w", err)
	}
	defer admFile.Close()

	eventConfigFile, err := os.Open(eventConfigPath)
	if err != nil {
		return fmt.Errorf("open event config: %w", err)
	}
	defer eventConfigFile.Close()

	teamConfigFile, err := os.Open(teamConfigPath)
	if err != nil {
		return fmt.Errorf("open team config: %w", err)
	}
	defer teamConfigFile.Close()

	settings, err := eventbuild.LoadEventSettings(eventConfigFile)
	if err != nil {
		return err
	}
	teams, err := eventbuild.LoadTeamConfig(teamConfigFile)
	if err != nil {
		return err
	}

	bundle, err := eventbuild.Build(admFile, filepath.Base(admPath), eventbuild.BuildOptions{
		Window:      teamguess.Window{Start: start, End: end},
		WindowStart: startValue,
		WindowEnd:   endValue,
		Settings:    settings,
		Teams:       teams,
	})
	if err != nil {
		return err
	}

	for _, warning := range bundle.Warnings {
		fmt.Fprintln(os.Stderr, warning)
	}

	return eventbuild.WriteBundle(bundle, outputDir)
}
