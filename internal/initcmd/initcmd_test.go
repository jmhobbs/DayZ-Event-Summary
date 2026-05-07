package initcmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmhobbs/dayz-event-summary/internal/runconfig"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunWritesConfigAndTeamsForTeamEvents(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	logPath := filepath.Join(workspace, "DayZ.ADM")
	require.NoError(t, os.WriteFile(logPath, []byte(strings.Join([]string{
		`15:40:00 | Player "[Foxes] KnightZ" (id=fox-1) is connecting`,
		`15:40:10 | Player "[Foxes] Doxy" (id=fox-2) is connecting`,
	}, "\n")), 0o644))

	targetDir := filepath.Join(workspace, "battle-at-blackjack")
	err := Run(Options{
		NoColor:             true,
		WorkingDir:          workspace,
		Directory:           targetDir,
		LogFile:             logPath,
		Start:               "15:40:00",
		End:                 "16:00:00",
		EventName:           "Battle at Blackjack",
		AssistWindowSeconds: 30,
		TeamsEvent:          boolPtr(true),
		NoInput:             true,
		Stdout:              &strings.Builder{},
		Stderr:              &strings.Builder{},
	})
	require.NoError(t, err)

	config, err := runconfig.LoadFile(filepath.Join(targetDir, "config.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "../DayZ.ADM", config.LogFile)
	assert.Equal(t, "data", config.DataDir)
	assert.Equal(t, "html", config.HTMLDir)
	assert.True(t, config.TeamsEvent)

	teamConfig, err := os.ReadFile(filepath.Join(targetDir, "teams.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(teamConfig), "teams:")
	assert.Contains(t, string(teamConfig), "Foxes")
}

func TestRunPrintsUpdatedStatusOutput(t *testing.T) {
	workspace := t.TempDir()
	logPath := filepath.Join(workspace, "DayZ.ADM")
	require.NoError(t, os.WriteFile(logPath, []byte(strings.Join([]string{
		`15:40:00 | Player "[Foxes] KnightZ" (id=fox-1) is connecting`,
		`15:40:10 | Player "[Foxes] Doxy" (id=fox-2) is connecting`,
	}, "\n")), 0o644))

	targetDir := filepath.Join(workspace, "battle-at-blackjack")

	output, err := captureStdout(t, func() error {
		return Run(Options{
			NoColor:             true,
			WorkingDir:          workspace,
			Directory:           targetDir,
			LogFile:             logPath,
			Start:               "15:40:00",
			End:                 "16:00:00",
			EventName:           "Battle at Blackjack",
			AssistWindowSeconds: 30,
			TeamsEvent:          boolPtr(true),
			NoInput:             true,
			Stdout:              &strings.Builder{},
			Stderr:              &strings.Builder{},
		})
	})
	require.NoError(t, err)

	assert.Contains(t, output, "DayZ Event Summary")
	assert.Contains(t, output, "┌────────────────────┐")
	assert.Contains(t, output, "✓ Event initialized in "+targetDir)
	assert.Contains(t, output, "! Teams file generated; please review before continuing.")
}

func TestRunSkipsTeamsForSingles(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	logPath := filepath.Join(workspace, "DayZ.ADM")
	require.NoError(t, os.WriteFile(logPath, []byte(`15:40:00 | Player "Solo" (id=solo-1) is connecting`), 0o644))

	targetDir := filepath.Join(workspace, "singles")
	err := Run(Options{
		NoColor:             true,
		WorkingDir:          workspace,
		Directory:           targetDir,
		LogFile:             logPath,
		Start:               "15:40:00",
		End:                 "16:00:00",
		EventName:           "Singles",
		AssistWindowSeconds: 30,
		TeamsEvent:          boolPtr(false),
		NoInput:             true,
		Stdout:              &strings.Builder{},
		Stderr:              &strings.Builder{},
	})
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(targetDir, "teams.yaml"))
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestRunRefusesOverwriteWithoutForce(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	logPath := filepath.Join(workspace, "DayZ.ADM")
	require.NoError(t, os.WriteFile(logPath, []byte(`15:40:00 | Player "Solo" (id=solo-1) is connecting`), 0o644))

	targetDir := filepath.Join(workspace, "singles")
	require.NoError(t, os.MkdirAll(targetDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "config.yaml"), []byte("existing"), 0o644))

	err := Run(Options{
		NoColor:             true,
		WorkingDir:          workspace,
		Directory:           targetDir,
		LogFile:             logPath,
		Start:               "15:40:00",
		End:                 "16:00:00",
		EventName:           "Singles",
		AssistWindowSeconds: 30,
		TeamsEvent:          boolPtr(false),
		NoInput:             true,
		Stdout:              &strings.Builder{},
		Stderr:              &strings.Builder{},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRunSuggestsWindowFromLogFile(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	logPath := filepath.Join(workspace, "DayZ.ADM")
	require.NoError(t, os.WriteFile(logPath, []byte(strings.Join([]string{
		"AdminLog started on 2026-05-02 at 15:33:25",
		`15:40:00 | Player "Solo" (id=solo-1) is connecting`,
		`15:45:10 | Player "Solo" (id=solo-1) is connected`,
		`16:12:59 | Player "Solo" (id=solo-1 pos=<1,1,1>)[HP: 90] hit by Player "Other" (id=other-1 pos=<2,2,2>) into Torso(1) for 10 damage (Bullet_556x45) with Pioneer from 20.0 meters`,
	}, "\n")), 0o644))

	targetDir := filepath.Join(workspace, "singles")
	stdout := &strings.Builder{}
	err := Run(Options{
		NoColor:             true,
		WorkingDir:          workspace,
		Directory:           targetDir,
		LogFile:             logPath,
		EventName:           "Singles",
		AssistWindowSeconds: 30,
		TeamsEvent:          boolPtr(false),
		Stdin:               strings.NewReader(promptInput("", "", "y")),
		Stdout:              stdout,
		Stderr:              &strings.Builder{},
	})
	require.NoError(t, err)

	config, err := runconfig.LoadFile(filepath.Join(targetDir, "config.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "15:40:00", config.Start)
	assert.Equal(t, "16:12:59", config.End)
	assert.Equal(t, "DayZ.ADM", config.LogFile)
	assert.Contains(t, stdout.String(), "Start time")
	assert.Contains(t, stdout.String(), "End time")
	assert.Contains(t, stdout.String(), "Move log file into event folder")
}

func TestDefaultDirectoryNameUsesDateFromLogFile(t *testing.T) {
	t.Parallel()

	name := defaultDirectoryName("", "/tmp/DayZServer_x64_2026-05-02_15-33-25.ADM", "15:33:25")
	assert.Equal(t, "dayzserver-x64-2026-05-02", name)
}

func TestRunOffersToMoveLogIntoEventFolder(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	logPath := filepath.Join(workspace, "DayZ.ADM")
	require.NoError(t, os.WriteFile(logPath, []byte(`15:40:00 | Player "Solo" (id=solo-1) is connecting`), 0o644))

	targetDir := filepath.Join(workspace, "singles")
	stdout := &strings.Builder{}
	err := Run(Options{
		NoColor:             true,
		WorkingDir:          workspace,
		Directory:           targetDir,
		LogFile:             logPath,
		Start:               "15:40:00",
		End:                 "16:00:00",
		EventName:           "Singles",
		AssistWindowSeconds: 30,
		TeamsEvent:          boolPtr(false),
		Stdin:               strings.NewReader(promptInput("y")),
		Stdout:              stdout,
		Stderr:              &strings.Builder{},
	})
	require.NoError(t, err)

	config, err := runconfig.LoadFile(filepath.Join(targetDir, "config.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "DayZ.ADM", config.LogFile)
	_, err = os.Stat(filepath.Join(targetDir, "DayZ.ADM"))
	require.NoError(t, err)
	_, err = os.Stat(logPath)
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
	assert.Contains(t, stdout.String(), "Move log file into event folder")
}

func TestRunSkipsMovePromptWhenLogAlreadyInEventFolder(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	targetDir := filepath.Join(workspace, "singles")
	require.NoError(t, os.MkdirAll(targetDir, 0o755))
	logPath := filepath.Join(targetDir, "DayZ.ADM")
	require.NoError(t, os.WriteFile(logPath, []byte(`15:40:00 | Player "Solo" (id=solo-1) is connecting`), 0o644))

	stdout := &strings.Builder{}
	err := Run(Options{
		NoColor:             true,
		WorkingDir:          workspace,
		Directory:           targetDir,
		LogFile:             logPath,
		Start:               "15:40:00",
		End:                 "16:00:00",
		EventName:           "Singles",
		AssistWindowSeconds: 30,
		TeamsEvent:          boolPtr(false),
		Stdin:               strings.NewReader(""),
		Stdout:              stdout,
		Stderr:              &strings.Builder{},
	})
	require.NoError(t, err)

	config, err := runconfig.LoadFile(filepath.Join(targetDir, "config.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "DayZ.ADM", config.LogFile)
	assert.NotContains(t, stdout.String(), "Move log file into event folder")
}

func boolPtr(value bool) *bool {
	return &value
}

func promptInput(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

func captureStdout(t *testing.T, run func() error) (string, error) {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = writer

	runErr := run()

	require.NoError(t, writer.Close())
	os.Stdout = originalStdout

	output, readErr := io.ReadAll(reader)
	require.NoError(t, reader.Close())
	require.NoError(t, readErr)

	return string(output), runErr
}
