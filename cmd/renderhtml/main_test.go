package main

import (
	"testing"

	"github.com/jmhobbs/dayz-event-summary/internal/runconfig"
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

	dataDir, outputDir := resolvePaths("/tmp/event/config.yaml", runconfig.Config{
		DataDir: "data",
		HTMLDir: "html",
	}, "", "")

	assert.Equal(t, "/tmp/event/data", dataDir)
	assert.Equal(t, "/tmp/event/html", outputDir)
}

func TestResolvePathsRespectsOverrides(t *testing.T) {
	t.Parallel()

	dataDir, outputDir := resolvePaths("/tmp/event/config.yaml", runconfig.Config{
		DataDir: "data",
		HTMLDir: "html",
	}, "/custom/data", "/custom/html")

	assert.Equal(t, "/custom/data", dataDir)
	assert.Equal(t, "/custom/html", outputDir)
}
