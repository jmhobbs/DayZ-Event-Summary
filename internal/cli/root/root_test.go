package root

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionUsesDevFallback(t *testing.T) {
	t.Parallel()

	stdout := &strings.Builder{}
	cmd := New(Options{
		Stdin:  strings.NewReader(""),
		Stdout: stdout,
		Stderr: &strings.Builder{},
	})

	err := cmd.ParseAndRun(context.Background(), []string{"version"})
	require.NoError(t, err)
	assert.Equal(t, "dev\n", stdout.String())
}

func TestVersionUsesConfiguredValue(t *testing.T) {
	t.Parallel()

	stdout := &strings.Builder{}
	cmd := New(Options{
		Stdin:   strings.NewReader(""),
		Stdout:  stdout,
		Stderr:  &strings.Builder{},
		Version: "v1.2.3",
	})

	err := cmd.ParseAndRun(context.Background(), []string{"version"})
	require.NoError(t, err)
	assert.Equal(t, "v1.2.3\n", stdout.String())
}

func TestRootDispatchesToSubcommands(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "init",
			args:    []string{"init"},
			wantErr: "missing required log file argument",
		},
		{
			name:    "build",
			args:    []string{"build"},
			wantErr: "--config is required",
		},
		{
			name:    "render",
			args:    []string{"render"},
			wantErr: "--config is required",
		},
		{
			name:    "generate-teams",
			args:    []string{"generate-teams"},
			wantErr: "--config is required",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := New(Options{
				Stdin:  strings.NewReader(""),
				Stdout: &strings.Builder{},
				Stderr: &strings.Builder{},
			})

			err := cmd.ParseAndRun(context.Background(), testCase.args)
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.wantErr)
		})
	}
}
