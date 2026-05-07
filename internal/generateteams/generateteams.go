package generateteams

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmhobbs/dayz-event-summary/internal/runconfig"
	"github.com/jmhobbs/dayz-event-summary/internal/teamguess"
)

func Run(configPath string, outputPath string) error {
	config, err := runconfig.LoadFile(configPath)
	if err != nil {
		return err
	}
	if !config.TeamsEvent {
		return fmt.Errorf("teams_event must be true to generate teams.yaml")
	}

	logPath := runconfig.ResolvePath(configPath, config.LogFile)
	start, err := teamguess.ParseClock(config.Start)
	if err != nil {
		return fmt.Errorf("parse start: %w", err)
	}
	end, err := teamguess.ParseClock(config.End)
	if err != nil {
		return fmt.Errorf("parse end: %w", err)
	}

	logFile, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer func() {
		_ = logFile.Close()
	}()

	players, err := teamguess.ExtractPlayers(logFile, teamguess.Window{Start: start, End: end})
	if err != nil {
		return fmt.Errorf("extract players: %w", err)
	}

	rendered, err := teamguess.MarshalYAML(teamguess.Suggest(players))
	if err != nil {
		return fmt.Errorf("render teams.yaml: %w", err)
	}

	targetPath := resolveOutputPath(configPath, outputPath)
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("teams output already exists: %s", targetPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check teams output: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create teams output directory: %w", err)
	}
	if err := os.WriteFile(targetPath, rendered, 0o644); err != nil {
		return fmt.Errorf("write teams.yaml: %w", err)
	}

	return nil
}

func resolveOutputPath(configPath string, outputPath string) string {
	if outputPath != "" {
		if filepath.IsAbs(outputPath) {
			return filepath.Clean(outputPath)
		}
		return filepath.Clean(filepath.Join(filepath.Dir(configPath), outputPath))
	}

	return filepath.Join(filepath.Dir(configPath), "teams.yaml")
}
