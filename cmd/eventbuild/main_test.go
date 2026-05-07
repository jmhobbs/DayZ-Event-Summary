package main

import (
	"testing"

	"amd-report/internal/runconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOptionsRequiresConfig(t *testing.T) {
	t.Parallel()

	_, err := parseOptions(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--config is required")
}

func TestResolvePathsUsesConfigDefaults(t *testing.T) {
	t.Parallel()

	paths := resolvePaths("/tmp/event/config.yaml", runconfig.Config{
		LogFile:    "../DayZ.ADM",
		TeamsEvent: true,
		DataDir:    "data",
	}, "", "")

	assert.Equal(t, "/tmp/DayZ.ADM", paths.LogFilePath)
	assert.Equal(t, "/tmp/event/teams.yaml", paths.TeamConfigPath)
	assert.Equal(t, "/tmp/event/data", paths.OutputDir)
}

func TestResolvePathsSkipsTeamsForSinglesUnlessOverridden(t *testing.T) {
	t.Parallel()

	paths := resolvePaths("/tmp/event/config.yaml", runconfig.Config{
		LogFile:    "../DayZ.ADM",
		TeamsEvent: false,
		DataDir:    "data",
	}, "", "")
	assert.Empty(t, paths.TeamConfigPath)

	overridden := resolvePaths("/tmp/event/config.yaml", runconfig.Config{
		LogFile:    "../DayZ.ADM",
		TeamsEvent: false,
		DataDir:    "data",
	}, "/custom/teams.yaml", "/custom/out")
	assert.Equal(t, "/custom/teams.yaml", overridden.TeamConfigPath)
	assert.Equal(t, "/custom/out", overridden.OutputDir)
}
