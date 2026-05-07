package initcmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jmhobbs/dayz-event-summary/internal/runconfig"
	"github.com/jmhobbs/dayz-event-summary/internal/teamguess"

	"github.com/fatih/color"
)

var (
	logTimePattern      = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}) \| `)
	logFilenameDateExpr = regexp.MustCompile(`(\d{4}-\d{2}-\d{2})`)
	logFilenameTimeExpr = regexp.MustCompile(`^[_-]\d{2}-\d{2}-\d{2}$`)
)

type Options struct {
	NoColor             bool
	WorkingDir          string
	Directory           string
	LogFile             string
	Start               string
	End                 string
	EventName           string
	AssistWindowSeconds int
	TeamsEvent          *bool
	Force               bool
	NoInput             bool
	Stdin               io.Reader
	Stdout              io.Writer
	Stderr              io.Writer
}

func Run(options Options) error {
	color.NoColor = options.NoColor

	boxBorderColorizer := color.RGB(135, 0, 0).SprintFunc()
	fmt.Println("")
	fmt.Println(boxBorderColorizer("  ┌────────────────────┐"))
	fmt.Println(boxBorderColorizer("  │"), "DayZ Event Summary", boxBorderColorizer("│"))
	fmt.Println(boxBorderColorizer("  └────────────────────┘"))
	fmt.Println("")

	workingDir, err := normalizeWorkingDir(options.WorkingDir)
	if err != nil {
		return err
	}

	stdin := options.Stdin
	if stdin == nil {
		stdin = os.Stdin
	}
	stdout := options.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	prompter := promptState{
		reader:  bufio.NewReader(stdin),
		writer:  stdout,
		noInput: options.NoInput,
		workDir: workingDir,
	}

	logFilePath, err := absolutePath(workingDir, options.LogFile)
	if err != nil {
		return err
	}
	fmt.Printf("Creating event config for %q\n\n", logFilePath)

	if _, err := os.Stat(logFilePath); err != nil {
		return fmt.Errorf("log file not found, or can not be read: %w", err)
	}

	defaultStart, defaultEnd, err := suggestWindow(logFilePath)
	if err != nil {
		return err
	}

	start, err := prompter.promptString("Start time", options.Start, defaultStart)
	if err != nil {
		return err
	}
	end, err := prompter.promptString("End time", options.End, defaultEnd)
	if err != nil {
		return err
	}
	eventName, err := prompter.promptString("Event name", options.EventName, "")
	if err != nil {
		return err
	}
	teamsEvent, err := prompter.promptBool("Teams event", options.TeamsEvent, true)
	if err != nil {
		return err
	}
	assistWindowSeconds, err := prompter.promptInt("Assist window seconds", options.AssistWindowSeconds, 30)
	if err != nil {
		return err
	}

	directoryDefault := defaultDirectoryName(eventName, options.LogFile, start)
	directory, err := prompter.promptString("Event folder", options.Directory, directoryDefault)
	if err != nil {
		return err
	}

	targetDir, err := absolutePath(workingDir, directory)
	if err != nil {
		return err
	}

	configPath := filepath.Join(targetDir, "config.yaml")
	teamsPath := filepath.Join(targetDir, "teams.yaml")
	if err := checkOverwrite(options, prompter, configPath, teamsPath, teamsEvent); err != nil {
		return err
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create event folder: %w", err)
	}

	var rendered []byte
	if teamsEvent {
		startClock, err := teamguess.ParseClock(start)
		if err != nil {
			return fmt.Errorf("parse start: %w", err)
		}
		endClock, err := teamguess.ParseClock(end)
		if err != nil {
			return fmt.Errorf("parse end: %w", err)
		}

		input, err := os.Open(logFilePath)
		if err != nil {
			return fmt.Errorf("open log file: %w", err)
		}

		players, err := teamguess.ExtractPlayers(input, teamguess.Window{Start: startClock, End: endClock})
		closeErr := input.Close()
		if err != nil {
			return fmt.Errorf("extract players: %w", err)
		}
		if closeErr != nil {
			return fmt.Errorf("close log file: %w", closeErr)
		}

		rendered, err = teamguess.MarshalYAML(teamguess.Suggest(players))
		if err != nil {
			return fmt.Errorf("render teams.yaml: %w", err)
		}
	}

	finalLogFilePath, err := maybeMoveLogFile(options, prompter, logFilePath, targetDir)
	if err != nil {
		return err
	}

	config := runconfig.Config{
		LogFile:             relativeIfPossible(targetDir, finalLogFilePath),
		Start:               start,
		End:                 end,
		EventName:           eventName,
		AssistWindowSeconds: assistWindowSeconds,
		TeamsEvent:          teamsEvent,
		DataDir:             "data",
		HTMLDir:             "html",
	}

	configFile, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("create config.yaml: %w", err)
	}
	if err := runconfig.Write(configFile, config); err != nil {
		configFile.Close()
		return err
	}
	if err := configFile.Close(); err != nil {
		return fmt.Errorf("close config.yaml: %w", err)
	}

	if teamsEvent {
		if err := os.WriteFile(teamsPath, rendered, 0o644); err != nil {
			return fmt.Errorf("write teams.yaml: %w", err)
		}
	}

	fmt.Println("")
	fmt.Printf(color.GreenString("✓")+" Event initialized in %s\n", targetDir)
	if teamsEvent {
		fmt.Println(color.YellowString("!") + " Teams file generated; please review before continuing.")
	}

	return nil
}

type promptState struct {
	reader  *bufio.Reader
	writer  io.Writer
	noInput bool
	workDir string
}

func (state promptState) promptString(label string, provided string, defaultValue string) (string, error) {
	if strings.TrimSpace(provided) != "" {
		return strings.TrimSpace(provided), nil
	}
	if state.noInput {
		return "", fmt.Errorf("%s is required in non-interactive mode", strings.ToLower(label))
	}

	if err := state.writePrompt(label, defaultValue); err != nil {
		return "", err
	}

	return state.readString(label, defaultValue)
}

func (state promptState) promptBool(label string, provided *bool, defaultValue bool) (bool, error) {
	if provided != nil {
		return *provided, nil
	}
	if state.noInput {
		return false, fmt.Errorf("%s is required in non-interactive mode", strings.ToLower(label))
	}

	defaultText := "Y/n"
	defaultValueStr := "y"
	if !defaultValue {
		defaultText = "y/N"
		defaultValueStr = "n"
	}

	if err := state.writePrompt(label, defaultText); err != nil {
		return false, err
	}

	value, err := state.readString(label, defaultValueStr)
	if err != nil {
		return false, err
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "Y/n", "y", "yes", "true":
		return true, nil
	case "y/N", "n", "no", "false":
		return false, nil
	default:
		return false, fmt.Errorf("unsupported boolean response %q", value)
	}
}

func (state promptState) promptInt(label string, provided int, defaultValue int) (int, error) {
	if provided > 0 {
		return provided, nil
	}
	if state.noInput {
		return 0, fmt.Errorf("%s is required in non-interactive mode", strings.ToLower(label))
	}

	if err := state.writePrompt(label, strconv.Itoa(defaultValue)); err != nil {
		return 0, err
	}

	value, err := state.readString(label, strconv.Itoa(defaultValue))
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", strings.ToLower(label), err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", strings.ToLower(label))
	}

	return parsed, nil
}

var promptColorizer = color.New(color.FgWhite, color.Bold)
var promptDefaultColorizer = color.RGB(148, 148, 148)

func (state promptState) writePrompt(label string, defaultValue string) error {
	prompt := promptColorizer.Sprint(label)
	if strings.TrimSpace(defaultValue) != "" {
		prompt = fmt.Sprintf("%s %s", promptColorizer.Sprint(label), promptDefaultColorizer.Sprintf("[%s]", defaultValue))
	}
	_, err := fmt.Fprintf(state.writer, "%s: ", prompt)
	return err
}

func (state promptState) readString(label string, defaultValue string) (string, error) {
	line, err := state.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	value := strings.TrimSpace(line)
	if value == "" {
		value = defaultValue
	}
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required", strings.ToLower(label))
	}

	return value, nil
}

func checkOverwrite(options Options, state promptState, configPath string, teamsPath string, teamsEvent bool) error {
	existing := []string{}
	if fileExists(configPath) {
		existing = append(existing, "config.yaml")
	}
	if teamsEvent && fileExists(teamsPath) {
		existing = append(existing, "teams.yaml")
	}
	if len(existing) == 0 || options.Force {
		return nil
	}
	if options.NoInput {
		return fmt.Errorf("%s already exists; rerun with --force", strings.Join(existing, ", "))
	}

	overwrite, err := state.promptBool(
		fmt.Sprintf("Overwrite %s", strings.Join(existing, " and ")),
		nil,
		false,
	)
	if err != nil {
		return err
	}
	if !overwrite {
		return fmt.Errorf("refusing to overwrite existing scaffold files")
	}

	return nil
}

func maybeMoveLogFile(options Options, state promptState, logFilePath string, targetDir string) (string, error) {
	if sameDirectory(logFilePath, targetDir) || options.NoInput {
		return logFilePath, nil
	}

	move, err := state.promptBool("Move log file into event folder", nil, true)
	if err != nil {
		return "", err
	}
	if !move {
		return logFilePath, nil
	}

	destination := filepath.Join(targetDir, filepath.Base(logFilePath))
	if filepath.Clean(destination) == filepath.Clean(logFilePath) {
		return logFilePath, nil
	}
	if fileExists(destination) {
		return "", fmt.Errorf("destination log file already exists: %s", destination)
	}
	if err := os.Rename(logFilePath, destination); err != nil {
		return "", fmt.Errorf("move log file into event folder: %w", err)
	}

	return destination, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func sameDirectory(logFilePath string, targetDir string) bool {
	return filepath.Clean(filepath.Dir(logFilePath)) == filepath.Clean(targetDir)
}

func normalizeWorkingDir(value string) (string, error) {
	if strings.TrimSpace(value) != "" {
		return absolutePath("", value)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	return workingDir, nil
}

func absolutePath(workingDir string, value string) (string, error) {
	if filepath.IsAbs(value) {
		return filepath.Clean(value), nil
	}
	base := workingDir
	if base == "" {
		var err error
		base, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}
	}

	return filepath.Clean(filepath.Join(base, value)), nil
}

func relativeIfPossible(baseDir string, target string) string {
	relative, err := filepath.Rel(baseDir, target)
	if err != nil {
		return filepath.Clean(target)
	}

	return filepath.Clean(relative)
}

func defaultDirectoryName(eventName string, logFile string, start string) string {
	name := strings.TrimSpace(eventName)
	if name == "" {
		name = logFileDirectoryStem(logFile)
		name = slugify(name)
		if name != "" {
			return name
		}
	} else {
		name = slugify(name)
		if date := logFileDate(logFile); date != "" {
			return fmt.Sprintf("%s-%s", name, date)
		}
	}
	if strings.TrimSpace(start) == "" {
		return name
	}

	return fmt.Sprintf("%s-%s", name, strings.ReplaceAll(start, ":", "-"))
}

func logFileDate(path string) string {
	match := logFilenameDateExpr.FindStringSubmatch(filepath.Base(path))
	if len(match) < 2 {
		return ""
	}

	return match[1]
}

func logFileDirectoryStem(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	date := logFileDate(path)
	if date == "" {
		return base
	}
	index := strings.Index(base, date)
	if index == -1 {
		return base
	}
	suffix := base[index+len(date):]
	if logFilenameTimeExpr.MatchString(suffix) {
		return base[:index+len(date)]
	}

	return base
}

func suggestWindow(logFilePath string) (string, string, error) {
	file, err := os.Open(logFilePath)
	if err != nil {
		return "", "", fmt.Errorf("open log file for time suggestions: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	first := ""
	last := ""
	for scanner.Scan() {
		match := logTimePattern.FindStringSubmatch(scanner.Text())
		if len(match) == 0 {
			continue
		}
		if first == "" {
			first = match[1]
		}
		last = match[1]
	}
	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("scan log file for time suggestions: %w", err)
	}

	return first, last, nil
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastHyphen := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastHyphen = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen && builder.Len() > 0 {
				builder.WriteByte('-')
				lastHyphen = true
			}
		}
	}

	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "event"
	}

	return result
}
