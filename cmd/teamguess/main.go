package main

import (
	"flag"
	"fmt"
	"os"

	"amd-report/internal/teamguess"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "teamguess: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var admPath string
	var startValue string
	var endValue string
	var outputPath string

	flag.StringVar(&admPath, "adm", "", "Path to the ADM file")
	flag.StringVar(&startValue, "start", "", "Event start time (HH:MM:SS or RFC3339)")
	flag.StringVar(&endValue, "end", "", "Event end time (HH:MM:SS or RFC3339)")
	flag.StringVar(&outputPath, "out", "", "Optional output YAML path, defaults to stdout")
	flag.Parse()

	if admPath == "" || startValue == "" || endValue == "" {
		return fmt.Errorf("--adm, --start, and --end are required")
	}

	start, err := teamguess.ParseClock(startValue)
	if err != nil {
		return fmt.Errorf("parse --start: %w", err)
	}

	end, err := teamguess.ParseClock(endValue)
	if err != nil {
		return fmt.Errorf("parse --end: %w", err)
	}

	input, err := os.Open(admPath)
	if err != nil {
		return fmt.Errorf("open ADM file: %w", err)
	}
	defer input.Close()

	players, err := teamguess.ExtractPlayers(input, teamguess.Window{Start: start, End: end})
	if err != nil {
		return fmt.Errorf("extract players: %w", err)
	}

	rendered, err := teamguess.MarshalYAML(teamguess.Suggest(players))
	if err != nil {
		return fmt.Errorf("render YAML: %w", err)
	}

	if outputPath == "" {
		_, err = os.Stdout.Write(rendered)
		return err
	}

	return os.WriteFile(outputPath, rendered, 0o644)
}
