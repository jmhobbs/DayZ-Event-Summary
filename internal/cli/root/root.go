package root

import (
	"flag"
	"io"

	"github.com/jmhobbs/dayz-event-summary/internal/cli/buildcmd"
	"github.com/jmhobbs/dayz-event-summary/internal/cli/initcmd"
	"github.com/jmhobbs/dayz-event-summary/internal/cli/rendercmd"
	"github.com/jmhobbs/dayz-event-summary/internal/cli/versioncmd"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type Options struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Version string
}

func New(options Options) *ffcli.Command {
	flags := flag.NewFlagSet("dayz-event-summary", flag.ContinueOnError)
	flags.SetOutput(options.Stderr)

	return &ffcli.Command{
		ShortUsage: "dayz-event-summary <subcommand>",
		ShortHelp:  "Generate DayZ event data and static HTML reports.",
		FlagSet:    flags,
		Subcommands: []*ffcli.Command{
			initcmd.New(options.Stdin, options.Stdout, options.Stderr),
			buildcmd.New(options.Stderr),
			rendercmd.New(options.Stderr),
			versioncmd.New(options.Stdout, options.Version),
		},
	}
}
