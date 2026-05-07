package initcmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strconv"

	appinitcmd "github.com/jmhobbs/dayz-event-summary/internal/initcmd"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func New(stdin io.Reader, stdout io.Writer, stderr io.Writer) *ffcli.Command {
	var options appinitcmd.Options
	var teamsEvent optionalBool
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.BoolVar(&options.NoColor, "no-color", false, "Disable color output")
	flags.StringVar(&options.Directory, "dir", "", "Event folder to create")
	flags.StringVar(&options.Start, "start", "", "Event start time (HH:MM:SS)")
	flags.StringVar(&options.End, "end", "", "Event end time (HH:MM:SS)")
	flags.StringVar(&options.EventName, "event-name", "", "Event name")
	flags.IntVar(&options.AssistWindowSeconds, "assist-window-seconds", 0, "Assist window in seconds")
	flags.Var(&teamsEvent, "teams-event", "Whether this is a team event")
	flags.BoolVar(&options.Force, "force", false, "Overwrite existing scaffold files")
	flags.BoolVar(&options.NoInput, "no-input", false, "Disable prompts and require values from flags")

	return &ffcli.Command{
		Name:       "init",
		ShortUsage: "dayz-event-summary init [flags] <log-file>",
		ShortHelp:  "Scaffold an event folder and write config files.",
		FlagSet:    flags,
		Exec: func(_ context.Context, args []string) error {
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("missing required log file argument")
			}

			runOptions := options
			runOptions.LogFile = args[0]
			if teamsEvent.set {
				value := teamsEvent.value
				runOptions.TeamsEvent = &value
			} else {
				runOptions.TeamsEvent = nil
			}
			runOptions.Stdin = stdin
			runOptions.Stdout = stdout
			runOptions.Stderr = stderr

			return appinitcmd.Run(runOptions)
		},
	}
}

type optionalBool struct {
	set   bool
	value bool
}

func (flagValue *optionalBool) String() string {
	if !flagValue.set {
		return ""
	}

	return strconv.FormatBool(flagValue.value)
}

func (flagValue *optionalBool) Set(value string) error {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}

	flagValue.set = true
	flagValue.value = parsed
	return nil
}

func (flagValue *optionalBool) IsBoolFlag() bool {
	return true
}
