package versioncmd

import (
	"context"
	"fmt"
	"io"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func New(stdout io.Writer, version string) *ffcli.Command {
	if version == "" {
		version = "dev"
	}

	return &ffcli.Command{
		Name:       "version",
		ShortUsage: "dayz-event-summary version",
		ShortHelp:  "Print the CLI version.",
		Exec: func(_ context.Context, _ []string) error {
			_, err := fmt.Fprintln(stdout, version)
			return err
		},
	}
}
