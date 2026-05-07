package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"amd-report/internal/initcmd"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var options initcmd.Options
	var teamsEvent optionalBool

	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.BoolVar(&options.NoColor, "no-color", false, "Disable color output")
	flags.StringVar(&options.Directory, "dir", "", "Event folder to create")
	flags.StringVar(&options.Start, "start", "", "Event start time (HH:MM:SS)")
	flags.StringVar(&options.End, "end", "", "Event end time (HH:MM:SS)")
	flags.StringVar(&options.EventName, "event-name", "", "Event name")
	flags.IntVar(&options.AssistWindowSeconds, "assist-window-seconds", 0, "Assist window in seconds")
	flags.Var(&teamsEvent, "teams-event", "Whether this is a team event")
	flags.BoolVar(&options.Force, "force", false, "Overwrite existing scaffold files")
	flags.BoolVar(&options.NoInput, "no-input", false, "Disable prompts and require values from flags")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if flags.Arg(0) == "" {
		return fmt.Errorf("missing required log file argument")
	}
	options.LogFile = flags.Arg(0)

	if teamsEvent.set {
		options.TeamsEvent = &teamsEvent.value
	}
	options.Stdin = os.Stdin
	options.Stdout = os.Stdout
	options.Stderr = os.Stderr

	return initcmd.Run(options)
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
