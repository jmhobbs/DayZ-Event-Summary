package runconfig

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAppliesDefaults(t *testing.T) {
	t.Parallel()

	config, err := Load(strings.NewReader(strings.Join([]string{
		"log_file: ../DayZ.ADM",
		"start: 15:33:25",
		"end: 20:52:21",
		"event_name: Saturday Teams Event",
		"teams_event: true",
	}, "\n")))
	require.NoError(t, err)

	assert.Equal(t, "../DayZ.ADM", config.LogFile)
	assert.Equal(t, "15:33:25", config.Start)
	assert.Equal(t, "20:52:21", config.End)
	assert.Equal(t, "Saturday Teams Event", config.EventName)
	assert.Equal(t, 30, config.AssistWindowSeconds)
	assert.Equal(t, "data", config.DataDir)
	assert.Equal(t, "html", config.HTMLDir)
	assert.True(t, config.TeamsEvent)
}

func TestLoadRejectsInvalidWindowFormat(t *testing.T) {
	t.Parallel()

	_, err := Load(strings.NewReader(strings.Join([]string{
		"log_file: ../DayZ.ADM",
		"start: 2026-05-02T15:33:25Z",
		"end: 20:52:21",
		"event_name: Saturday Teams Event",
		"assist_window_seconds: 30",
		"teams_event: false",
		"data_dir: data",
		"html_dir: html",
	}, "\n")))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start")
}

func TestResolvePathUsesConfigDirectory(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		"/tmp/event/logs/DayZ.ADM",
		ResolvePath("/tmp/event/config.yaml", "logs/DayZ.ADM"),
	)
	assert.Equal(t,
		"/var/tmp/DayZ.ADM",
		ResolvePath("/tmp/event/config.yaml", "/var/tmp/DayZ.ADM"),
	)
}

func TestWriteSerializesExpectedFields(t *testing.T) {
	t.Parallel()

	var output strings.Builder
	err := Write(&output, Config{
		LogFile:             "../DayZ.ADM",
		Start:               "15:33:25",
		End:                 "20:52:21",
		EventName:           "Saturday Teams Event",
		AssistWindowSeconds: 45,
		TeamsEvent:          true,
		DataDir:             "data",
		HTMLDir:             "html",
	})
	require.NoError(t, err)

	text := output.String()
	assert.Contains(t, text, "log_file: ../DayZ.ADM")
	assert.Contains(t, text, "start: \"15:33:25\"")
	assert.Contains(t, text, "end: \"20:52:21\"")
	assert.Contains(t, text, "event_name: Saturday Teams Event")
	assert.Contains(t, text, "assist_window_seconds: 45")
	assert.Contains(t, text, "teams_event: true")
	assert.Contains(t, text, "data_dir: data")
	assert.Contains(t, text, "html_dir: html")
}
