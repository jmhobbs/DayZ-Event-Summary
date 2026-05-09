package eventbuild

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/jmhobbs/dayz-event-summary/internal/teamguess"
)

var (
	hitPattern           = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}) \| Player "([^"]+)"( \(DEAD\))? \(id=([^ )]+)[^)]*\)\[HP: [^\]]+\] hit by Player "([^"]+)" \(id=([^ )]+)[^)]*\) into ([^(]+)\(\d+\) for ([0-9.]+) damage \(([^)]+)\)(?: with (.+?))?(?: from ([0-9.]+) meters)?\s*$`)
	killPattern          = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}) \| Player "([^"]+)"(?: \(DEAD\))? \(id=([^ )]+)[^)]*\) killed by Player "([^"]+)" \(id=([^ )]+)[^)]*\) with (.+?)(?: from ([0-9.]+) meters)?\s*$`)
	nonPlayerHitPattern  = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}) \| Player "([^"]+)"(?: \(DEAD\))? \(id=([^ )]+)[^)]*\)\[HP: [^\]]+\] hit by (.+?) into ([^(]+)\(\d+\) for ([0-9.]+) damage \(([^)]+)\)\s*$`)
	nonPlayerKillPattern = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}) \| Player "([^"]+)"(?: \(DEAD\))? \(id=([^ )]+)[^)]*\) killed by (.+?)\s*$`)
	suicidePattern       = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}) \| Player "([^"]+)"(?: \(DEAD\))? \(id=([^ )]+)[^)]*\) committed suicide\s*$`)
	diedPattern          = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}) \| Player "([^"]+)"(?: \(DEAD\))? \(id=([^ )]+)[^)]*\) died\.`)
)

type BuildOptions struct {
	Window      teamguess.Window
	WindowStart string
	WindowEnd   string
	Settings    EventSettings
	Teams       TeamConfig
	GeneratedAt time.Time
}

type teamAssignment struct {
	TeamName      string
	PreferredName string
	DisplayName   string
}

type deathMarker struct {
	Clock time.Duration
	Line  int
}

type hitRecord struct {
	Event HitEvent
	Clock time.Duration
}

type rosterEntry struct {
	PlayerID      string
	PreferredName string
	DisplayName   string
	TeamName      string
	SeenNames     []string
}

type playerAccumulator struct {
	Report PlayerReport
}

type teamAccumulator struct {
	Report   TeamReport
	Matchups map[string]*TeamMatchupSummary
}

type bundleFile struct {
	Name  string
	Value any
}

func Build(reader io.Reader, sourceADM string, options BuildOptions) (*Bundle, error) {
	options.Settings.ApplyDefaults()
	if options.GeneratedAt.IsZero() {
		options.GeneratedAt = time.Now().UTC()
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read ADM content: %w", err)
	}

	players, err := teamguess.ExtractPlayersAll(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("extract roster: %w", err)
	}

	ignoredIDs := buildIgnoredIDs(options.Teams)
	players = filterPlayers(players, ignoredIDs)
	assignments := buildAssignments(options.Teams)
	roster := buildRoster(players, assignments)

	events, warnings, err := parseEvents(bytes.NewReader(content), options.Window, roster, options.Settings.AssistWindowSeconds, ignoredIDs)
	if err != nil {
		return nil, err
	}

	bundle := &Bundle{
		Metadata: Metadata{
			EventName:           options.Settings.EventName,
			SourceADM:           sourceADM,
			WindowStart:         options.WindowStart,
			WindowEnd:           options.WindowEnd,
			AssistWindowSeconds: options.Settings.AssistWindowSeconds,
			GeneratedAt:         options.GeneratedAt.Format(time.RFC3339),
			PlayerCount:         len(roster),
		},
		Roster:   RosterFile{Players: rosterPlayers(roster)},
		Events:   events,
		Warnings: warnings,
	}

	playersFile, teamsFile, summaryFile := buildAggregates(bundle.Roster.Players, events)
	bundle.Players = playersFile
	bundle.Teams = teamsFile
	bundle.Summary = summaryFile
	bundle.Metadata.TeamCount = len(bundle.Teams.Teams)

	return bundle, nil
}

func WriteBundle(bundle *Bundle, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	for _, file := range bundleFiles(bundle) {
		rendered, err := json.MarshalIndent(file.Value, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal %s: %w", file.Name, err)
		}
		rendered = append(rendered, '\n')

		if err := os.WriteFile(filepath.Join(outputDir, file.Name), rendered, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", file.Name, err)
		}

		schema, err := marshalSchema(file.Value)
		if err != nil {
			return fmt.Errorf("marshal %s schema: %w", file.Name, err)
		}

		schemaPath := filepath.Join(outputDir, schemaFileName(file.Name))
		if err := os.WriteFile(schemaPath, schema, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", schemaFileName(file.Name), err)
		}
	}

	return nil
}

func bundleFiles(bundle *Bundle) []bundleFile {
	return []bundleFile{
		{Name: "metadata.json", Value: bundle.Metadata},
		{Name: "roster.json", Value: bundle.Roster},
		{Name: "events.json", Value: bundle.Events},
		{Name: "players.json", Value: bundle.Players},
		{Name: "teams.json", Value: bundle.Teams},
		{Name: "summary.json", Value: bundle.Summary},
	}
}

func marshalSchema(value any) ([]byte, error) {
	reflector := jsonschema.Reflector{
		Anonymous:      true,
		ExpandedStruct: true,
	}

	rendered, err := json.MarshalIndent(reflector.Reflect(value), "", "  ")
	if err != nil {
		return nil, err
	}

	return append(rendered, '\n'), nil
}

func schemaFileName(fileName string) string {
	return strings.TrimSuffix(fileName, ".json") + ".schema.json"
}

func buildAssignments(config TeamConfig) map[string]teamAssignment {
	assignments := map[string]teamAssignment{}

	for _, team := range config.Teams {
		for _, member := range team.Members {
			if isIgnoredMember(config, member.PlayerID) {
				continue
			}
			assignments[member.PlayerID] = teamAssignment{
				TeamName:      team.Name,
				PreferredName: member.PreferredName,
				DisplayName:   member.DisplayName,
			}
		}
	}

	for _, member := range config.Ungrouped {
		if isIgnoredMember(config, member.PlayerID) {
			continue
		}
		assignments[member.PlayerID] = teamAssignment{
			PreferredName: member.PreferredName,
		}
	}

	return assignments
}

func buildIgnoredIDs(config TeamConfig) map[string]struct{} {
	ignored := make(map[string]struct{}, len(config.Ignored))
	for _, member := range config.Ignored {
		if strings.TrimSpace(member.PlayerID) == "" {
			continue
		}
		ignored[member.PlayerID] = struct{}{}
	}

	return ignored
}

func isIgnoredMember(config TeamConfig, playerID string) bool {
	for _, member := range config.Ignored {
		if member.PlayerID == playerID {
			return true
		}
	}

	return false
}

func filterPlayers(players []teamguess.Player, ignoredIDs map[string]struct{}) []teamguess.Player {
	filtered := make([]teamguess.Player, 0, len(players))
	for _, player := range players {
		if isIgnoredID(ignoredIDs, player.PlayerID) {
			continue
		}
		filtered = append(filtered, player)
	}

	return filtered
}

func buildRoster(players []teamguess.Player, assignments map[string]teamAssignment) map[string]rosterEntry {
	roster := map[string]rosterEntry{}
	for _, player := range players {
		entry := rosterEntry{
			PlayerID:      player.PlayerID,
			PreferredName: player.LatestName,
			SeenNames:     append([]string(nil), player.SeenNames...),
		}

		if assignment, ok := assignments[player.PlayerID]; ok {
			if assignment.PreferredName != "" {
				entry.PreferredName = assignment.PreferredName
			}
			if assignment.DisplayName != "" {
				entry.DisplayName = assignment.DisplayName
			}
			entry.TeamName = assignment.TeamName
		}

		entry.DisplayName = displayName(entry.DisplayName, entry.PreferredName)
		roster[player.PlayerID] = entry
	}

	return roster
}

func rosterPlayers(roster map[string]rosterEntry) []RosterPlayer {
	players := make([]RosterPlayer, 0, len(roster))
	for _, entry := range roster {
		players = append(players, RosterPlayer(entry))
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].DisplayName < players[j].DisplayName
	})

	return players
}

func parseEvents(reader io.Reader, window teamguess.Window, roster map[string]rosterEntry, assistWindowSeconds int, ignoredIDs map[string]struct{}) (EventsFile, []string, error) {
	scanner := bufio.NewScanner(reader)
	lineNumber := 0
	warnings := []string{}
	events := EventsFile{}
	hitsByVictim := map[string][]hitRecord{}
	lastDeaths := map[string]deathMarker{}
	assistWindow := time.Duration(assistWindowSeconds) * time.Second

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		clock, ok, err := parseClockFromLine(line)
		if err != nil {
			return EventsFile{}, nil, fmt.Errorf("parse line %d time: %w", lineNumber, err)
		}
		if !ok || !window.Contains(clock) {
			continue
		}

		if matches := hitPattern.FindStringSubmatch(line); len(matches) > 0 {
			if matches[3] != "" {
				continue
			}
			if strings.TrimSpace(matches[10]) == "TransportHit" {
				continue
			}
			if isIgnoredID(ignoredIDs, matches[4]) || isIgnoredID(ignoredIDs, matches[6]) {
				continue
			}

			damage, err := strconv.ParseFloat(matches[8], 64)
			if err != nil {
				warnings = append(warnings, formatWarning(lineNumber, "invalid hit damage", line))
				continue
			}

			hit := HitEvent{
				Timestamp:           matches[1],
				LineNumber:          lineNumber,
				AttackerID:          matches[6],
				AttackerDisplayName: rosterDisplayName(roster, matches[6], matches[5]),
				VictimID:            matches[4],
				VictimDisplayName:   rosterDisplayName(roster, matches[4], matches[2]),
				BodyPart:            strings.TrimSpace(matches[7]),
				Damage:              damage,
				AmmoType:            strings.TrimSpace(matches[9]),
				Weapon:              strings.TrimSpace(matches[10]),
				RangeMeters:         parseOptionalFloat(matches[11]),
				Teamkill:            isTeamkill(roster, matches[6], matches[4]),
			}

			events.Hits = append(events.Hits, hit)
			hitsByVictim[hit.VictimID] = append(hitsByVictim[hit.VictimID], hitRecord{Event: hit, Clock: clock})
			continue
		}

		if matches := killPattern.FindStringSubmatch(line); len(matches) > 0 {
			if isIgnoredID(ignoredIDs, matches[3]) || isIgnoredID(ignoredIDs, matches[5]) {
				continue
			}
			kill := KillEvent{
				Timestamp:         matches[1],
				LineNumber:        lineNumber,
				KillerID:          matches[5],
				KillerDisplayName: rosterDisplayName(roster, matches[5], matches[4]),
				VictimID:          matches[3],
				VictimDisplayName: rosterDisplayName(roster, matches[3], matches[2]),
				Weapon:            strings.TrimSpace(matches[6]),
				RangeMeters:       parseOptionalFloat(matches[7]),
				Teamkill:          isTeamkill(roster, matches[5], matches[3]),
			}

			events.Kills = append(events.Kills, kill)
			lastDeaths[kill.VictimID] = deathMarker{Clock: clock, Line: lineNumber}

			if !kill.Teamkill {
				assists := collectAssists(hitsByVictim[kill.VictimID], kill, roster, assistWindow, clock)
				events.Assists = append(events.Assists, assists...)
			}
			continue
		}

		if matches := nonPlayerHitPattern.FindStringSubmatch(line); len(matches) > 0 {
			if isIgnoredID(ignoredIDs, matches[3]) {
				continue
			}
			continue
		}

		if matches := nonPlayerKillPattern.FindStringSubmatch(line); len(matches) > 0 {
			if isIgnoredID(ignoredIDs, matches[3]) {
				continue
			}
			env := EnvironmentalDeathEvent{
				Timestamp:         matches[1],
				LineNumber:        lineNumber,
				VictimID:          matches[3],
				VictimDisplayName: rosterDisplayName(roster, matches[3], matches[2]),
				Cause:             strings.TrimSpace(matches[4]),
			}

			events.EnvironmentalDeaths = append(events.EnvironmentalDeaths, env)
			lastDeaths[env.VictimID] = deathMarker{Clock: clock, Line: lineNumber}
			continue
		}

		if matches := suicidePattern.FindStringSubmatch(line); len(matches) > 0 {
			if isIgnoredID(ignoredIDs, matches[3]) {
				continue
			}
			env := EnvironmentalDeathEvent{
				Timestamp:         matches[1],
				LineNumber:        lineNumber,
				VictimID:          matches[3],
				VictimDisplayName: rosterDisplayName(roster, matches[3], matches[2]),
				Cause:             "suicide",
			}

			events.EnvironmentalDeaths = append(events.EnvironmentalDeaths, env)
			lastDeaths[env.VictimID] = deathMarker{Clock: clock, Line: lineNumber}
			continue
		}

		if matches := diedPattern.FindStringSubmatch(line); len(matches) > 0 {
			if isIgnoredID(ignoredIDs, matches[3]) {
				continue
			}
			if marker, ok := lastDeaths[matches[3]]; ok && clock-marker.Clock <= 10*time.Second {
				continue
			}

			env := EnvironmentalDeathEvent{
				Timestamp:         matches[1],
				LineNumber:        lineNumber,
				VictimID:          matches[3],
				VictimDisplayName: rosterDisplayName(roster, matches[3], matches[2]),
				Cause:             "died",
			}

			events.EnvironmentalDeaths = append(events.EnvironmentalDeaths, env)
			lastDeaths[env.VictimID] = deathMarker{Clock: clock, Line: lineNumber}
			continue
		}

		if looksCombatLike(line) {
			warnings = append(warnings, formatWarning(lineNumber, "unparsed combat-like line", line))
		}
	}

	if err := scanner.Err(); err != nil {
		return EventsFile{}, nil, err
	}

	return events, warnings, nil
}

func collectAssists(hitRecords []hitRecord, kill KillEvent, roster map[string]rosterEntry, assistWindow time.Duration, killClock time.Duration) []AssistEvent {
	assisters := map[string]AssistEvent{}
	for _, record := range hitRecords {
		if record.Clock > killClock {
			continue
		}
		if killClock-record.Clock > assistWindow {
			continue
		}
		if record.Event.AttackerID == kill.KillerID {
			continue
		}

		assist, exists := assisters[record.Event.AttackerID]
		if !exists {
			assist = AssistEvent{
				Timestamp:           kill.Timestamp,
				KillLineNumber:      kill.LineNumber,
				KillerID:            kill.KillerID,
				KillerDisplayName:   kill.KillerDisplayName,
				VictimID:            kill.VictimID,
				VictimDisplayName:   kill.VictimDisplayName,
				AssisterID:          record.Event.AttackerID,
				AssisterDisplayName: rosterDisplayName(roster, record.Event.AttackerID, record.Event.AttackerDisplayName),
				Hits:                []AssistHitDetail{},
			}
		}
		assist.Hits = append(assist.Hits, AssistHitDetail{
			Timestamp:   record.Event.Timestamp,
			LineNumber:  record.Event.LineNumber,
			Weapon:      record.Event.Weapon,
			RangeMeters: record.Event.RangeMeters,
		})
		assisters[record.Event.AttackerID] = assist
	}

	events := make([]AssistEvent, 0, len(assisters))
	for _, assist := range assisters {
		sort.Slice(assist.Hits, func(i, j int) bool {
			if assist.Hits[i].LineNumber != assist.Hits[j].LineNumber {
				return assist.Hits[i].LineNumber < assist.Hits[j].LineNumber
			}
			return assist.Hits[i].Timestamp < assist.Hits[j].Timestamp
		})
		events = append(events, assist)
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].Timestamp != events[j].Timestamp {
			return events[i].Timestamp < events[j].Timestamp
		}

		return events[i].AssisterDisplayName < events[j].AssisterDisplayName
	})

	return events
}

func buildAggregates(roster []RosterPlayer, events EventsFile) (PlayersFile, TeamsFile, SummaryFile) {
	rosterByID := map[string]RosterPlayer{}
	playerAccumulators := map[string]*playerAccumulator{}
	teamAccumulators := map[string]*teamAccumulator{}

	for _, player := range roster {
		rosterByID[player.PlayerID] = player
		playerAccumulators[player.PlayerID] = &playerAccumulator{
			Report: PlayerReport{
				PlayerID:            player.PlayerID,
				PreferredName:       player.PreferredName,
				DisplayName:         player.DisplayName,
				TeamName:            player.TeamName,
				HitsDealt:           HitBreakdown{ByBodyPart: map[string]int{}},
				HitsTaken:           HitBreakdown{ByBodyPart: map[string]int{}},
				Kills:               []PlayerKillDetail{},
				Deaths:              []PlayerDeathDetail{},
				Assists:             []PlayerAssistDetail{},
				Teamkills:           []PlayerKillDetail{},
				EnvironmentalDeaths: []PlayerEnvironmentalDeath{},
			},
		}
		if player.TeamName != "" {
			teamAccumulators[player.TeamName] = ensureTeamAccumulator(teamAccumulators, player.TeamName)
		}
	}

	for _, hit := range events.Hits {
		attacker := ensurePlayerAccumulator(playerAccumulators, rosterByID, hit.AttackerID)
		victim := ensurePlayerAccumulator(playerAccumulators, rosterByID, hit.VictimID)
		attacker.Report.HitsDealt.Total++
		attacker.Report.HitsDealt.ByBodyPart[hit.BodyPart]++
		victim.Report.HitsTaken.Total++
		victim.Report.HitsTaken.ByBodyPart[hit.BodyPart]++
	}

	for _, kill := range events.Kills {
		killer := ensurePlayerAccumulator(playerAccumulators, rosterByID, kill.KillerID)
		victim := ensurePlayerAccumulator(playerAccumulators, rosterByID, kill.VictimID)
		victimRoster := rosterByID[kill.VictimID]
		killerRoster := rosterByID[kill.KillerID]

		detail := PlayerKillDetail{
			Timestamp:         kill.Timestamp,
			LineNumber:        kill.LineNumber,
			VictimID:          kill.VictimID,
			VictimDisplayName: rosterDisplay(kill.VictimID, rosterByID, kill.VictimDisplayName),
			VictimTeamName:    victimRoster.TeamName,
			Weapon:            kill.Weapon,
			RangeMeters:       kill.RangeMeters,
			Teamkill:          kill.Teamkill,
		}

		deathDetail := PlayerDeathDetail{
			Timestamp:         kill.Timestamp,
			LineNumber:        kill.LineNumber,
			KillerID:          kill.KillerID,
			KillerDisplayName: rosterDisplay(kill.KillerID, rosterByID, kill.KillerDisplayName),
			KillerTeamName:    killerRoster.TeamName,
			Weapon:            kill.Weapon,
			RangeMeters:       kill.RangeMeters,
			Teamkill:          kill.Teamkill,
			CountsTowardStats: !kill.Teamkill,
		}

		if kill.Teamkill {
			killer.Report.Stats.Teamkills++
			killer.Report.Teamkills = append(killer.Report.Teamkills, detail)
		} else {
			killer.Report.Stats.Kills++
			killer.Report.Kills = append(killer.Report.Kills, detail)
			victim.Report.Stats.Deaths++
		}
		victim.Report.Deaths = append(victim.Report.Deaths, deathDetail)

		if killerRoster.TeamName != "" {
			team := ensureTeamAccumulator(teamAccumulators, killerRoster.TeamName)
			if kill.Teamkill {
				team.Report.Stats.Teamkills++
				team.Report.Teamkills = append(team.Report.Teamkills, detail)
			} else {
				team.Report.Stats.Kills++
			}
		}
		if !kill.Teamkill && victimRoster.TeamName != "" {
			ensureTeamAccumulator(teamAccumulators, victimRoster.TeamName).Report.Stats.Deaths++
		}
		if !kill.Teamkill && killerRoster.TeamName != "" && victimRoster.TeamName != "" && killerRoster.TeamName != victimRoster.TeamName {
			killerMatchup := ensureMatchup(teamAccumulators, killerRoster.TeamName, victimRoster.TeamName)
			killerMatchup.Kills++
			victimMatchup := ensureMatchup(teamAccumulators, victimRoster.TeamName, killerRoster.TeamName)
			victimMatchup.Deaths++
		}
	}

	for _, assist := range events.Assists {
		assister := ensurePlayerAccumulator(playerAccumulators, rosterByID, assist.AssisterID)
		assister.Report.Stats.Assists++
		assister.Report.Assists = append(assister.Report.Assists, PlayerAssistDetail{
			Timestamp:         assist.Timestamp,
			KillLineNumber:    assist.KillLineNumber,
			KillerID:          assist.KillerID,
			KillerDisplayName: rosterDisplay(assist.KillerID, rosterByID, assist.KillerDisplayName),
			VictimID:          assist.VictimID,
			VictimDisplayName: rosterDisplay(assist.VictimID, rosterByID, assist.VictimDisplayName),
			Hits:              append([]AssistHitDetail(nil), assist.Hits...),
		})

		if rosterByID[assist.AssisterID].TeamName != "" {
			ensureTeamAccumulator(teamAccumulators, rosterByID[assist.AssisterID].TeamName).Report.Stats.Assists++
		}
	}

	for _, death := range events.EnvironmentalDeaths {
		victim := ensurePlayerAccumulator(playerAccumulators, rosterByID, death.VictimID)
		victim.Report.Stats.EnvironmentalDeaths++
		victim.Report.EnvironmentalDeaths = append(victim.Report.EnvironmentalDeaths, PlayerEnvironmentalDeath{
			Timestamp:  death.Timestamp,
			LineNumber: death.LineNumber,
			Cause:      death.Cause,
		})
	}

	players := make([]PlayerReport, 0, len(playerAccumulators))
	for _, accumulator := range playerAccumulators {
		accumulator.Report.Stats.RatioValue, accumulator.Report.Stats.RatioDisplay = ratio(accumulator.Report.Stats.Kills, accumulator.Report.Stats.Deaths)
		players = append(players, accumulator.Report)
	}
	sort.Slice(players, func(i, j int) bool {
		return compareRows(players[i].Stats.Kills, players[j].Stats.Kills, players[i].Stats.Deaths, players[j].Stats.Deaths, players[i].Stats.RatioValue, players[j].Stats.RatioValue, players[i].DisplayName, players[j].DisplayName)
	})

	teams := make([]TeamReport, 0, len(teamAccumulators))
	for _, accumulator := range teamAccumulators {
		accumulator.Report.Stats.RatioValue, accumulator.Report.Stats.RatioDisplay = ratio(accumulator.Report.Stats.Kills, accumulator.Report.Stats.Deaths)
		for _, player := range players {
			if player.TeamName != accumulator.Report.Name {
				continue
			}
			accumulator.Report.Members = append(accumulator.Report.Members, TeamMemberSummary{
				PlayerID:            player.PlayerID,
				PreferredName:       player.PreferredName,
				DisplayName:         player.DisplayName,
				Kills:               player.Stats.Kills,
				Deaths:              player.Stats.Deaths,
				Assists:             player.Stats.Assists,
				Teamkills:           player.Stats.Teamkills,
				EnvironmentalDeaths: player.Stats.EnvironmentalDeaths,
				RatioValue:          player.Stats.RatioValue,
				RatioDisplay:        player.Stats.RatioDisplay,
			})
		}
		sort.Slice(accumulator.Report.Members, func(i, j int) bool {
			return compareRows(accumulator.Report.Members[i].Kills, accumulator.Report.Members[j].Kills, accumulator.Report.Members[i].Deaths, accumulator.Report.Members[j].Deaths, accumulator.Report.Members[i].RatioValue, accumulator.Report.Members[j].RatioValue, accumulator.Report.Members[i].DisplayName, accumulator.Report.Members[j].DisplayName)
		})

		for _, matchup := range accumulator.Matchups {
			matchup.RatioValue, matchup.RatioDisplay = ratio(matchup.Kills, matchup.Deaths)
			accumulator.Report.Matchups = append(accumulator.Report.Matchups, *matchup)
		}
		sort.Slice(accumulator.Report.Matchups, func(i, j int) bool {
			if accumulator.Report.Matchups[i].Kills != accumulator.Report.Matchups[j].Kills {
				return accumulator.Report.Matchups[i].Kills > accumulator.Report.Matchups[j].Kills
			}
			return accumulator.Report.Matchups[i].OpponentTeamName < accumulator.Report.Matchups[j].OpponentTeamName
		})

		teams = append(teams, accumulator.Report)
	}
	sort.Slice(teams, func(i, j int) bool {
		return compareRows(teams[i].Stats.Kills, teams[j].Stats.Kills, teams[i].Stats.Deaths, teams[j].Stats.Deaths, teams[i].Stats.RatioValue, teams[j].Stats.RatioValue, teams[i].Name, teams[j].Name)
	})

	summary := SummaryFile{
		PlayerTable:  make([]SummaryPlayerRow, 0, len(players)),
		TeamTable:    make([]SummaryTeamRow, 0, len(teams)),
		LongestHits:  buildLongestHits(events.Hits, events.Kills),
		LongestKills: buildLongestKills(events.Kills),
	}
	for _, player := range players {
		summary.PlayerTable = append(summary.PlayerTable, SummaryPlayerRow{
			PlayerID:            player.PlayerID,
			PreferredName:       player.PreferredName,
			DisplayName:         player.DisplayName,
			TeamName:            player.TeamName,
			Kills:               player.Stats.Kills,
			Deaths:              player.Stats.Deaths,
			Assists:             player.Stats.Assists,
			Teamkills:           player.Stats.Teamkills,
			EnvironmentalDeaths: player.Stats.EnvironmentalDeaths,
			RatioValue:          player.Stats.RatioValue,
			RatioDisplay:        player.Stats.RatioDisplay,
		})
	}
	for _, team := range teams {
		summary.TeamTable = append(summary.TeamTable, SummaryTeamRow{
			Name:         team.Name,
			Kills:        team.Stats.Kills,
			Deaths:       team.Stats.Deaths,
			Assists:      team.Stats.Assists,
			Teamkills:    team.Stats.Teamkills,
			RatioValue:   team.Stats.RatioValue,
			RatioDisplay: team.Stats.RatioDisplay,
		})
	}

	return PlayersFile{Players: players}, TeamsFile{Teams: teams}, summary
}

func ensurePlayerAccumulator(players map[string]*playerAccumulator, rosterByID map[string]RosterPlayer, playerID string) *playerAccumulator {
	if accumulator, ok := players[playerID]; ok {
		return accumulator
	}

	player := rosterByID[playerID]
	accumulator := &playerAccumulator{
		Report: PlayerReport{
			PlayerID:            playerID,
			PreferredName:       player.PreferredName,
			DisplayName:         player.DisplayName,
			TeamName:            player.TeamName,
			HitsDealt:           HitBreakdown{ByBodyPart: map[string]int{}},
			HitsTaken:           HitBreakdown{ByBodyPart: map[string]int{}},
			Kills:               []PlayerKillDetail{},
			Deaths:              []PlayerDeathDetail{},
			Assists:             []PlayerAssistDetail{},
			Teamkills:           []PlayerKillDetail{},
			EnvironmentalDeaths: []PlayerEnvironmentalDeath{},
		},
	}
	players[playerID] = accumulator
	return accumulator
}

func ensureTeamAccumulator(teams map[string]*teamAccumulator, teamName string) *teamAccumulator {
	if team, ok := teams[teamName]; ok {
		return team
	}

	team := &teamAccumulator{
		Report: TeamReport{
			Name:      teamName,
			Members:   []TeamMemberSummary{},
			Matchups:  []TeamMatchupSummary{},
			Teamkills: []PlayerKillDetail{},
		},
		Matchups: map[string]*TeamMatchupSummary{},
	}
	teams[teamName] = team
	return team
}

func ensureMatchup(teams map[string]*teamAccumulator, teamName string, opponentTeam string) *TeamMatchupSummary {
	team := ensureTeamAccumulator(teams, teamName)
	if matchup, ok := team.Matchups[opponentTeam]; ok {
		return matchup
	}

	matchup := &TeamMatchupSummary{OpponentTeamName: opponentTeam}
	team.Matchups[opponentTeam] = matchup
	return matchup
}

func buildLongestHits(hits []HitEvent, kills []KillEvent) []ShotSummary {
	fatalHitLines := fatalHitLineNumbers(hits, kills)
	entries := make([]ShotSummary, 0, len(hits))
	for _, hit := range hits {
		if hit.RangeMeters == nil {
			continue
		}
		entries = append(entries, ShotSummary{
			Timestamp:           hit.Timestamp,
			LineNumber:          hit.LineNumber,
			AttackerID:          hit.AttackerID,
			AttackerDisplayName: hit.AttackerDisplayName,
			VictimID:            hit.VictimID,
			VictimDisplayName:   hit.VictimDisplayName,
			Weapon:              hit.Weapon,
			RangeMeters:         *hit.RangeMeters,
			BodyPart:            hit.BodyPart,
			IsKill:              isFatalHit(fatalHitLines, hit.LineNumber),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].RangeMeters != entries[j].RangeMeters {
			return entries[i].RangeMeters > entries[j].RangeMeters
		}
		return entries[i].LineNumber < entries[j].LineNumber
	})
	return entries
}

func buildLongestKills(kills []KillEvent) []ShotSummary {
	entries := make([]ShotSummary, 0, len(kills))
	for _, kill := range kills {
		if kill.RangeMeters == nil {
			continue
		}
		entries = append(entries, ShotSummary{
			Timestamp:           kill.Timestamp,
			LineNumber:          kill.LineNumber,
			AttackerID:          kill.KillerID,
			AttackerDisplayName: kill.KillerDisplayName,
			VictimID:            kill.VictimID,
			VictimDisplayName:   kill.VictimDisplayName,
			Weapon:              kill.Weapon,
			RangeMeters:         *kill.RangeMeters,
			IsKill:              true,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].RangeMeters != entries[j].RangeMeters {
			return entries[i].RangeMeters > entries[j].RangeMeters
		}
		return entries[i].LineNumber < entries[j].LineNumber
	})
	return entries
}

func fatalHitLineNumbers(hits []HitEvent, kills []KillEvent) map[int]struct{} {
	lines := make(map[int]struct{}, len(kills))
	used := make(map[int]struct{}, len(kills))

	for _, kill := range kills {
		match := -1
		for i := len(hits) - 1; i >= 0; i-- {
			hit := hits[i]
			if hit.LineNumber >= kill.LineNumber {
				continue
			}
			if _, ok := used[hit.LineNumber]; ok {
				continue
			}
			if !sameShot(hit, kill) {
				continue
			}
			match = hit.LineNumber
			break
		}
		if match == -1 {
			continue
		}
		lines[match] = struct{}{}
		used[match] = struct{}{}
	}

	return lines
}

func sameShot(hit HitEvent, kill KillEvent) bool {
	if hit.AttackerID != kill.KillerID || hit.VictimID != kill.VictimID {
		return false
	}
	if strings.TrimSpace(hit.Weapon) != strings.TrimSpace(kill.Weapon) {
		return false
	}
	if kill.RangeMeters == nil {
		return true
	}
	if hit.RangeMeters == nil {
		return false
	}
	return *hit.RangeMeters == *kill.RangeMeters
}

func isFatalHit(fatalHitLines map[int]struct{}, lineNumber int) bool {
	_, ok := fatalHitLines[lineNumber]
	return ok
}

func compareRows(leftKills int, rightKills int, leftDeaths int, rightDeaths int, leftRatio *float64, rightRatio *float64, leftName string, rightName string) bool {
	if leftKills != rightKills {
		return leftKills > rightKills
	}
	if leftDeaths != rightDeaths {
		return leftDeaths < rightDeaths
	}

	leftRatioValue := ratioSortValue(leftRatio)
	rightRatioValue := ratioSortValue(rightRatio)
	if leftRatioValue != rightRatioValue {
		return leftRatioValue > rightRatioValue
	}

	return leftName < rightName
}

func ratioSortValue(value *float64) float64 {
	if value == nil {
		return 1_000_000
	}
	return *value
}

func ratio(kills int, deaths int) (*float64, string) {
	if deaths == 0 {
		if kills > 0 {
			return nil, fmt.Sprintf("%.2f", float64(kills))
		}
		return nil, "Perfect"
	}

	value := float64(kills) / float64(deaths)
	return &value, fmt.Sprintf("%.2f", value)
}

func parseClockFromLine(line string) (time.Duration, bool, error) {
	if len(line) < len("15:04:05 | ") {
		return 0, false, nil
	}
	if line[2] != ':' || line[5] != ':' {
		return 0, false, nil
	}

	clock, err := teamguess.ParseClock(line[:8])
	if err != nil {
		return 0, false, err
	}
	return clock, true, nil
}

func parseOptionalFloat(value string) *float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func rosterDisplayName(roster map[string]rosterEntry, playerID string, fallback string) string {
	if entry, ok := roster[playerID]; ok && entry.DisplayName != "" {
		return entry.DisplayName
	}
	return strings.TrimSpace(fallback)
}

func rosterDisplay(playerID string, rosterByID map[string]RosterPlayer, fallback string) string {
	if player, ok := rosterByID[playerID]; ok && player.DisplayName != "" {
		return player.DisplayName
	}
	return strings.TrimSpace(fallback)
}

func displayName(cleanName string, preferredName string) string {
	if cleanName != "" {
		return cleanName
	}
	return preferredName
}

func isTeamkill(roster map[string]rosterEntry, attackerID string, victimID string) bool {
	attacker, attackerOK := roster[attackerID]
	victim, victimOK := roster[victimID]
	if !attackerOK || !victimOK {
		return false
	}
	if attacker.TeamName == "" || victim.TeamName == "" {
		return false
	}
	return attacker.TeamName == victim.TeamName
}

func isIgnoredID(ignoredIDs map[string]struct{}, playerID string) bool {
	_, ok := ignoredIDs[playerID]
	return ok
}

func looksCombatLike(line string) bool {
	return strings.Contains(line, " hit by ") ||
		strings.Contains(line, " killed by ") ||
		strings.Contains(line, " committed suicide") ||
		strings.Contains(line, " died.")
}

func formatWarning(lineNumber int, reason string, line string) string {
	return fmt.Sprintf("line %d: %s: %s", lineNumber, reason, line)
}
