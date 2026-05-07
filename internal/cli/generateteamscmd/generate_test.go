package generateteamscmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOptionsRequiresConfig(t *testing.T) {
	t.Parallel()

	_, err := parseOptions(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--config is required")
}

func TestParseOptionsAcceptsOutputOverride(t *testing.T) {
	t.Parallel()

	options, err := parseOptions([]string{"--config", "config.yaml", "--out", "teams.generated.yaml"})
	require.NoError(t, err)
	assert.Equal(t, "config.yaml", options.ConfigPath)
	assert.Equal(t, "teams.generated.yaml", options.OutputPath)
}
