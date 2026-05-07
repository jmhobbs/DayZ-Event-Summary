package generateteams

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmhobbs/dayz-event-summary/internal/runconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunWritesTeamsFileNextToConfig(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	logPath := filepath.Join(workspace, "DayZ.ADM")
	require.NoError(t, os.WriteFile(logPath, []byte(strings.Join([]string{
		`15:40:00 | Player "[Foxes] KnightZ" (id=fox-1) is connecting`,
		`15:40:10 | Player "[Foxes] Doxy" (id=fox-2) is connecting`,
	}, "\n")), 0o644))

	configPath := filepath.Join(workspace, "config.yaml")
	configFile, err := os.Create(configPath)
	require.NoError(t, err)
	require.NoError(t, runconfig.Write(configFile, runconfig.Config{
		LogFile:             "DayZ.ADM",
		Start:               "15:40:00",
		End:                 "16:00:00",
		EventName:           "Battle at Blackjack",
		AssistWindowSeconds: 30,
		TeamsEvent:          true,
		DataDir:             "data",
		HTMLDir:             "html",
	}))
	require.NoError(t, configFile.Close())

	require.NoError(t, Run(configPath, ""))

	teamConfig, err := os.ReadFile(filepath.Join(workspace, "teams.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(teamConfig), "teams:")
	assert.Contains(t, string(teamConfig), "Foxes")
}

func TestRunRespectsOutputOverride(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	logPath := filepath.Join(workspace, "DayZ.ADM")
	require.NoError(t, os.WriteFile(logPath, []byte(`15:40:00 | Player "[Foxes] KnightZ" (id=fox-1) is connecting`), 0o644))

	configPath := filepath.Join(workspace, "config.yaml")
	configFile, err := os.Create(configPath)
	require.NoError(t, err)
	require.NoError(t, runconfig.Write(configFile, runconfig.Config{
		LogFile:             "DayZ.ADM",
		Start:               "15:40:00",
		End:                 "16:00:00",
		EventName:           "Battle at Blackjack",
		AssistWindowSeconds: 30,
		TeamsEvent:          true,
		DataDir:             "data",
		HTMLDir:             "html",
	}))
	require.NoError(t, configFile.Close())

	outputPath := filepath.Join(workspace, "custom-teams.yaml")
	require.NoError(t, Run(configPath, outputPath))

	_, err = os.Stat(outputPath)
	require.NoError(t, err)
}

func TestRunRejectsSinglesConfig(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	logPath := filepath.Join(workspace, "DayZ.ADM")
	require.NoError(t, os.WriteFile(logPath, []byte(`15:40:00 | Player "Solo" (id=solo-1) is connecting`), 0o644))

	configPath := filepath.Join(workspace, "config.yaml")
	configFile, err := os.Create(configPath)
	require.NoError(t, err)
	require.NoError(t, runconfig.Write(configFile, runconfig.Config{
		LogFile:             "DayZ.ADM",
		Start:               "15:40:00",
		End:                 "16:00:00",
		EventName:           "Singles",
		AssistWindowSeconds: 30,
		TeamsEvent:          false,
		DataDir:             "data",
		HTMLDir:             "html",
	}))
	require.NoError(t, configFile.Close())

	err = Run(configPath, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "teams_event must be true")
}
