package renderhtml

import (
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"amd-report/internal/eventbuild"
	"github.com/yosssi/ace"
)

//go:embed templates/*.ace templates/pages/*.ace templates/partials/*.ace assets/style.css
var assetsFS embed.FS

type Bundle struct {
	Metadata eventbuild.Metadata
	Roster   eventbuild.RosterFile
	Events   eventbuild.EventsFile
	Players  eventbuild.PlayersFile
	Teams    eventbuild.TeamsFile
	Summary  eventbuild.SummaryFile
}

type basePageData struct {
	Title       string
	EventName   string
	Heading     string
	Subheading  string
	AssetHref   string
	HomeHref    string
	PlayersHref string
	TeamsHref   string
	HitsHref    string
	KillsHref   string
	FooterText  string
}

type nameData struct {
	Display string
	Meta    string
	Href    string
}

type statData struct {
	Label string
	Value string
	Note  string
}

type indexPlayerRow struct {
	Name                nameData
	Team                string
	Kills               int
	Deaths              int
	Assists             int
	Teamkills           int
	EnvironmentalDeaths int
	Ratio               string
}

type indexTeamRow struct {
	Name      string
	Href      string
	Kills     int
	Deaths    int
	Assists   int
	Teamkills int
	Ratio     string
}

type shotRow struct {
	Timestamp string
	Attacker  nameData
	Victim    nameData
	Result    string
	Weapon    string
	BodyPart  string
	Range     string
}

type indexPageData struct {
	basePageData
	PlayerSection playerTableData
	TeamSection   teamTableData
	HitSection    shotTableData
	KillSection   shotTableData
}

type playerTableData struct {
	Title       string
	Note        string
	ViewAllHref string
	Rows        []indexPlayerRow
}

type teamTableData struct {
	Title       string
	Note        string
	ViewAllHref string
	Rows        []indexTeamRow
}

type playersPageData struct {
	basePageData
	PlayerSection playerTableData
}

type teamsPageData struct {
	basePageData
	TeamSection teamTableData
}

type shotTableData struct {
	Title        string
	Note         string
	ViewAllHref  string
	Rows         []shotRow
	ShowResult   bool
	ShowBodyPart bool
}

type shotsPageData struct {
	basePageData
	ShotSection shotTableData
}

type detailRow struct {
	Timestamp string
	Name      nameData
	Weapon    string
	Range     string
	Note      string
}

type assistRow struct {
	Timestamp string
	Killer    nameData
	Victim    nameData
	Hits      []assistHitRow
}

type assistHitRow struct {
	Weapon string
	Range  string
}

type environmentalRow struct {
	Timestamp string
	Cause     string
}

type breakdownRow struct {
	Label string
	Count int
}

type playerPageData struct {
	basePageData
	PlayerName          nameData
	Stats               []statData
	Kills               []detailRow
	Deaths              []detailRow
	Assists             []assistRow
	Teamkills           []detailRow
	EnvironmentalDeaths []environmentalRow
	HitsDealt           []breakdownRow
	HitsTaken           []breakdownRow
}

type memberRow struct {
	Name                nameData
	Kills               int
	Deaths              int
	Assists             int
	Teamkills           int
	EnvironmentalDeaths int
	Ratio               string
}

type matchupRow struct {
	Opponent string
	Kills    int
	Deaths   int
	Ratio    string
}

type teamPageData struct {
	basePageData
	TeamName  string
	Stats     []statData
	Members   []memberRow
	Matchups  []matchupRow
	Teamkills []detailRow
}

func LoadBundle(dataDir string) (Bundle, error) {
	var bundle Bundle

	if err := loadJSON(filepath.Join(dataDir, "metadata.json"), &bundle.Metadata); err != nil {
		return Bundle{}, err
	}
	if err := loadJSON(filepath.Join(dataDir, "roster.json"), &bundle.Roster); err != nil {
		return Bundle{}, err
	}
	if err := loadJSON(filepath.Join(dataDir, "events.json"), &bundle.Events); err != nil {
		return Bundle{}, err
	}
	if err := loadJSON(filepath.Join(dataDir, "players.json"), &bundle.Players); err != nil {
		return Bundle{}, err
	}
	if err := loadJSON(filepath.Join(dataDir, "teams.json"), &bundle.Teams); err != nil {
		return Bundle{}, err
	}
	if err := loadJSON(filepath.Join(dataDir, "summary.json"), &bundle.Summary); err != nil {
		return Bundle{}, err
	}

	return bundle, nil
}

func Render(bundle Bundle, outputDir string) error {
	renderer, err := newRenderer(bundle)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(outputDir, "assets"), 0o755); err != nil {
		return fmt.Errorf("create assets directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(outputDir, "players"), 0o755); err != nil {
		return fmt.Errorf("create players directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(outputDir, "hits"), 0o755); err != nil {
		return fmt.Errorf("create hits directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(outputDir, "kills"), 0o755); err != nil {
		return fmt.Errorf("create kills directory: %w", err)
	}
	if len(bundle.Teams.Teams) > 0 {
		if err := os.MkdirAll(filepath.Join(outputDir, "teams"), 0o755); err != nil {
			return fmt.Errorf("create teams directory: %w", err)
		}
	}

	css, err := assetsFS.ReadFile("assets/style.css")
	if err != nil {
		return fmt.Errorf("read embedded CSS: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "assets", "style.css"), css, 0o644); err != nil {
		return fmt.Errorf("write CSS: %w", err)
	}

	if err := renderer.renderIndex(filepath.Join(outputDir, "index.html")); err != nil {
		return err
	}
	if err := renderer.renderPlayersIndex(filepath.Join(outputDir, "players", "index.html")); err != nil {
		return err
	}
	if err := renderer.renderHitsIndex(filepath.Join(outputDir, "hits", "index.html")); err != nil {
		return err
	}
	if err := renderer.renderKillsIndex(filepath.Join(outputDir, "kills", "index.html")); err != nil {
		return err
	}
	if len(bundle.Teams.Teams) > 0 {
		if err := renderer.renderTeamsIndex(filepath.Join(outputDir, "teams", "index.html")); err != nil {
			return err
		}
	}
	for _, player := range bundle.Players.Players {
		if err := renderer.renderPlayer(filepath.Join(outputDir, "players", renderer.playerFile(player.PlayerID)), player); err != nil {
			return err
		}
	}
	for _, team := range bundle.Teams.Teams {
		if err := renderer.renderTeam(filepath.Join(outputDir, "teams", renderer.teamFile(team.Name)), team); err != nil {
			return err
		}
	}

	return nil
}

type renderer struct {
	bundle        Bundle
	playerReports map[string]eventbuild.PlayerReport
	teamReports   map[string]eventbuild.TeamReport
	playerFiles   map[string]string
	teamFiles     map[string]string
	templates     map[string]*template.Template
}

func newRenderer(bundle Bundle) (*renderer, error) {
	renderer := &renderer{
		bundle:        bundle,
		playerReports: map[string]eventbuild.PlayerReport{},
		teamReports:   map[string]eventbuild.TeamReport{},
		playerFiles:   map[string]string{},
		teamFiles:     map[string]string{},
		templates:     map[string]*template.Template{},
	}

	for _, player := range bundle.Players.Players {
		renderer.playerReports[player.PlayerID] = player
		renderer.playerFiles[player.PlayerID] = safeIDFileName(player.PlayerID) + ".html"
	}
	for _, team := range bundle.Teams.Teams {
		renderer.teamReports[team.Name] = team
		renderer.teamFiles[team.Name] = slugify(team.Name) + ".html"
	}

	for key, inner := range map[string]string{
		"index":       "pages/index",
		"hits-index":  "pages/hits",
		"kills-index": "pages/kills",
		"player":      "pages/player",
		"players":     "pages/players",
		"team":        "pages/team",
		"teams-index": "pages/teams",
	} {
		tpl, err := ace.Load("base", inner, &ace.Options{
			BaseDir: "templates",
			Indent:  "  ",
			Asset: func(name string) ([]byte, error) {
				return assetsFS.ReadFile(name)
			},
		})
		if err != nil {
			return nil, fmt.Errorf("load %s template: %w", key, err)
		}
		renderer.templates[key] = tpl
	}

	return renderer, nil
}

func (r *renderer) renderIndex(path string) error {
	playerRows := r.indexPlayers("players/")
	teamRows := r.indexTeams("teams/")
	hitRows := r.indexShots(r.bundle.Summary.LongestHits, "players/", true)
	killRows := r.indexShots(r.bundle.Summary.LongestKills, "players/", false)
	data := indexPageData{
		basePageData: r.basePage(
			r.pageTitle("Summary"),
			r.heading(),
			"Summary of players, teams, and longest-range actions.",
			"assets/style.css",
			"index.html",
			"players/index.html",
			r.teamsIndexHref("teams/index.html"),
			"hits/index.html",
			"kills/index.html",
		),
		PlayerSection: playerTableData{
			Title:       "Players",
			Note:        "Top 5 by kills, deaths, and ratio.",
			ViewAllHref: "players/index.html",
			Rows:        limitIndexPlayerRows(playerRows, 5),
		},
		TeamSection: teamTableData{
			Title:       "Teams",
			Note:        "Top 5 by kills, deaths, and ratio.",
			ViewAllHref: "teams/index.html",
			Rows:        limitIndexTeamRows(teamRows, 5),
		},
		HitSection: shotTableData{
			Title:        "Longest hits",
			ViewAllHref:  "hits/index.html",
			Rows:         limitShotRows(hitRows, 5),
			ShowResult:   true,
			ShowBodyPart: true,
		},
		KillSection: shotTableData{
			Title:       "Longest kills",
			ViewAllHref: "kills/index.html",
			Rows:        limitShotRows(killRows, 5),
		},
	}

	return r.execute("index", path, data)
}

func (r *renderer) renderPlayersIndex(path string) error {
	data := playersPageData{
		basePageData: r.basePage(
			r.pageTitle("Players"),
			"Players",
			"Full player leaderboard.",
			"../assets/style.css",
			"../index.html",
			"index.html",
			r.teamsIndexHref("../teams/index.html"),
			"../hits/index.html",
			"../kills/index.html",
		),
		PlayerSection: playerTableData{
			Title: "Players",
			Note:  "Full leaderboard sorted by kills, deaths, and ratio.",
			Rows:  r.indexPlayers(""),
		},
	}

	return r.execute("players", path, data)
}

func (r *renderer) renderPlayer(path string, player eventbuild.PlayerReport) error {
	data := playerPageData{
		basePageData: r.basePage(
			r.pageTitle(player.DisplayName),
			player.DisplayName,
			playerSubheading(player),
			"../assets/style.css",
			"../index.html",
			"index.html",
			r.teamsIndexHref("../teams/index.html"),
			"../hits/index.html",
			"../kills/index.html",
		),
		PlayerName: r.nameForPlayer(player.PlayerID, "", "", false),
		Stats: []statData{
			{Label: "Kills", Value: fmt.Sprintf("%d", player.Stats.Kills)},
			{Label: "Deaths", Value: fmt.Sprintf("%d", player.Stats.Deaths)},
			{Label: "Ratio", Value: player.Stats.RatioDisplay},
			{Label: "Assists", Value: fmt.Sprintf("%d", player.Stats.Assists)},
			{Label: "Teamkills", Value: fmt.Sprintf("%d", player.Stats.Teamkills)},
			{Label: "Environmental deaths", Value: fmt.Sprintf("%d", player.Stats.EnvironmentalDeaths)},
			{Label: "Hits dealt", Value: fmt.Sprintf("%d", player.HitsDealt.Total)},
			{Label: "Hits taken", Value: fmt.Sprintf("%d", player.HitsTaken.Total)},
		},
		Kills:               r.playerKills(player.Kills, false),
		Deaths:              r.playerDeaths(player.Deaths),
		Assists:             r.playerAssists(player.Assists),
		Teamkills:           r.playerKills(player.Teamkills, true),
		EnvironmentalDeaths: r.environmentalRows(player.EnvironmentalDeaths),
		HitsDealt:           breakdownRows(player.HitsDealt.ByBodyPart),
		HitsTaken:           breakdownRows(player.HitsTaken.ByBodyPart),
	}

	return r.execute("player", path, data)
}

func (r *renderer) renderTeam(path string, team eventbuild.TeamReport) error {
	members := make([]memberRow, 0, len(team.Members))
	for _, member := range team.Members {
		members = append(members, memberRow{
			Name:                r.nameForPlayer(member.PlayerID, "../players/"+r.playerFile(member.PlayerID), "", false),
			Kills:               member.Kills,
			Deaths:              member.Deaths,
			Assists:             member.Assists,
			Teamkills:           member.Teamkills,
			EnvironmentalDeaths: member.EnvironmentalDeaths,
			Ratio:               member.RatioDisplay,
		})
	}

	matchups := make([]matchupRow, 0, len(team.Matchups))
	for _, matchup := range team.Matchups {
		matchups = append(matchups, matchupRow{
			Opponent: matchup.OpponentTeamName,
			Kills:    matchup.Kills,
			Deaths:   matchup.Deaths,
			Ratio:    matchup.RatioDisplay,
		})
	}

	data := teamPageData{
		basePageData: r.basePage(
			r.pageTitle(team.Name),
			team.Name,
			"Team performance and member breakdown.",
			"../assets/style.css",
			"../index.html",
			"../players/index.html",
			"index.html",
			"../hits/index.html",
			"../kills/index.html",
		),
		TeamName: team.Name,
		Stats: []statData{
			{Label: "Kills", Value: fmt.Sprintf("%d", team.Stats.Kills)},
			{Label: "Deaths", Value: fmt.Sprintf("%d", team.Stats.Deaths)},
			{Label: "Ratio", Value: team.Stats.RatioDisplay},
			{Label: "Assists", Value: fmt.Sprintf("%d", team.Stats.Assists)},
			{Label: "Teamkills", Value: fmt.Sprintf("%d", team.Stats.Teamkills)},
			{Label: "Members", Value: fmt.Sprintf("%d", len(team.Members))},
		},
		Members:   members,
		Matchups:  matchups,
		Teamkills: r.playerKills(team.Teamkills, true),
	}

	return r.execute("team", path, data)
}

func (r *renderer) renderTeamsIndex(path string) error {
	data := teamsPageData{
		basePageData: r.basePage(
			r.pageTitle("Teams"),
			"Teams",
			"Full team leaderboard.",
			"../assets/style.css",
			"../index.html",
			"../players/index.html",
			"index.html",
			"../hits/index.html",
			"../kills/index.html",
		),
		TeamSection: teamTableData{
			Title: "Teams",
			Note:  "Full leaderboard sorted by kills, deaths, and ratio.",
			Rows:  r.indexTeams(""),
		},
	}

	return r.execute("teams-index", path, data)
}

func (r *renderer) renderHitsIndex(path string) error {
	data := shotsPageData{
		basePageData: r.basePage(
			r.pageTitle("Hits"),
			"Hits",
			"All ranged hits, sorted longest first.",
			"../assets/style.css",
			"../index.html",
			"../players/index.html",
			r.teamsIndexHref("../teams/index.html"),
			"index.html",
			"../kills/index.html",
		),
		ShotSection: shotTableData{
			Title:        "Longest hits",
			Note:         "Full list sorted by range.",
			Rows:         r.indexShots(r.bundle.Summary.LongestHits, "../players/", true),
			ShowResult:   true,
			ShowBodyPart: true,
		},
	}

	return r.execute("hits-index", path, data)
}

func (r *renderer) renderKillsIndex(path string) error {
	data := shotsPageData{
		basePageData: r.basePage(
			r.pageTitle("All Kills"),
			"All Kills",
			"All ranged kills, sorted longest first.",
			"../assets/style.css",
			"../index.html",
			"../players/index.html",
			r.teamsIndexHref("../teams/index.html"),
			"../hits/index.html",
			"index.html",
		),
		ShotSection: shotTableData{
			Title: "Longest kills",
			Note:  "Full list sorted by range.",
			Rows:  r.indexShots(r.bundle.Summary.LongestKills, "../players/", false),
		},
	}

	return r.execute("kills-index", path, data)
}

func (r *renderer) execute(key string, path string, data any) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer file.Close()

	if err := r.templates[key].Execute(file, data); err != nil {
		return fmt.Errorf("render %s: %w", key, err)
	}

	return nil
}

func (r *renderer) basePage(title string, heading string, subheading string, assetHref string, homeHref string, playersHref string, teamsHref string, hitsHref string, killsHref string) basePageData {
	return basePageData{
		Title:       title,
		EventName:   r.heading(),
		Heading:     heading,
		Subheading:  subheading,
		AssetHref:   assetHref,
		HomeHref:    homeHref,
		PlayersHref: playersHref,
		TeamsHref:   teamsHref,
		HitsHref:    hitsHref,
		KillsHref:   killsHref,
		FooterText:  fmt.Sprintf("Source: %s • Window: %s to %s • Generated: %s", r.bundle.Metadata.SourceADM, r.bundle.Metadata.WindowStart, r.bundle.Metadata.WindowEnd, r.bundle.Metadata.GeneratedAt),
	}
}

func (r *renderer) heading() string {
	if strings.TrimSpace(r.bundle.Metadata.EventName) != "" {
		return r.bundle.Metadata.EventName
	}
	return "DayZ Event Report"
}

func (r *renderer) pageTitle(section string) string {
	return fmt.Sprintf("%s • %s", section, r.heading())
}

func limitIndexPlayerRows(rows []indexPlayerRow, limit int) []indexPlayerRow {
	if limit <= 0 || len(rows) <= limit {
		return rows
	}

	return rows[:limit]
}

func limitIndexTeamRows(rows []indexTeamRow, limit int) []indexTeamRow {
	if limit <= 0 || len(rows) <= limit {
		return rows
	}

	return rows[:limit]
}

func limitShotRows(rows []shotRow, limit int) []shotRow {
	if limit <= 0 || len(rows) <= limit {
		return rows
	}

	return rows[:limit]
}

func (r *renderer) teamsIndexHref(href string) string {
	if len(r.bundle.Teams.Teams) == 0 {
		return ""
	}

	return href
}

func (r *renderer) indexPlayers(hrefPrefix string) []indexPlayerRow {
	rows := make([]indexPlayerRow, 0, len(r.bundle.Summary.PlayerTable))
	for _, row := range r.bundle.Summary.PlayerTable {
		rows = append(rows, indexPlayerRow{
			Name: nameData{
				Display: row.DisplayName,
				Href:    hrefPrefix + r.playerFile(row.PlayerID),
			},
			Team:                row.TeamName,
			Kills:               row.Kills,
			Deaths:              row.Deaths,
			Assists:             row.Assists,
			Teamkills:           row.Teamkills,
			EnvironmentalDeaths: row.EnvironmentalDeaths,
			Ratio:               row.RatioDisplay,
		})
	}

	return rows
}

func (r *renderer) indexTeams(hrefPrefix string) []indexTeamRow {
	rows := make([]indexTeamRow, 0, len(r.bundle.Summary.TeamTable))
	for _, row := range r.bundle.Summary.TeamTable {
		rows = append(rows, indexTeamRow{
			Name:      row.Name,
			Href:      hrefPrefix + r.teamFile(row.Name),
			Kills:     row.Kills,
			Deaths:    row.Deaths,
			Assists:   row.Assists,
			Teamkills: row.Teamkills,
			Ratio:     row.RatioDisplay,
		})
	}

	return rows
}

func (r *renderer) indexShots(shots []eventbuild.ShotSummary, hrefPrefix string, includeBodyPart bool) []shotRow {
	rows := make([]shotRow, 0, len(shots))
	for _, shot := range shots {
		row := shotRow{
			Timestamp: shot.Timestamp,
			Attacker:  r.nameForPlayer(shot.AttackerID, hrefPrefix+r.playerFile(shot.AttackerID), "", true),
			Victim:    r.nameForPlayer(shot.VictimID, hrefPrefix+r.playerFile(shot.VictimID), "", true),
			Result:    shotResult(shot.IsKill),
			Weapon:    dashIfEmpty(shot.Weapon),
			Range:     fmt.Sprintf("%.2f m", shot.RangeMeters),
		}
		if includeBodyPart {
			row.BodyPart = dashIfEmpty(shot.BodyPart)
		} else {
			row.BodyPart = dashIfEmpty(shot.BodyPart)
		}
		rows = append(rows, row)
	}

	return rows
}

func shotResult(isKill bool) string {
	if isKill {
		return "Kill"
	}
	return "Hit"
}

func (r *renderer) playerKills(kills []eventbuild.PlayerKillDetail, teamkill bool) []detailRow {
	rows := make([]detailRow, 0, len(kills))
	for _, kill := range kills {
		name := r.nameForPlayer(kill.VictimID, r.playerRelativeHref(kill.VictimID), kill.VictimTeamName, true)
		rows = append(rows, detailRow{
			Timestamp: kill.Timestamp,
			Name:      name,
			Weapon:    dashIfEmpty(kill.Weapon),
			Range:     rangeDisplay(kill.RangeMeters),
			Note:      teamkillNote(teamkill, kill.Teamkill),
		})
	}

	return rows
}

func (r *renderer) playerDeaths(deaths []eventbuild.PlayerDeathDetail) []detailRow {
	rows := make([]detailRow, 0, len(deaths))
	for _, death := range deaths {
		name := r.nameForPlayer(death.KillerID, r.playerRelativeHref(death.KillerID), death.KillerTeamName, true)
		note := ""
		if death.Teamkill {
			note = "Teamkill"
		}
		if !death.CountsTowardStats && note == "" {
			note = "Does not affect K/D/R"
		}
		rows = append(rows, detailRow{
			Timestamp: death.Timestamp,
			Name:      name,
			Weapon:    dashIfEmpty(death.Weapon),
			Range:     rangeDisplay(death.RangeMeters),
			Note:      note,
		})
	}

	return rows
}

func (r *renderer) playerAssists(assists []eventbuild.PlayerAssistDetail) []assistRow {
	rows := make([]assistRow, 0, len(assists))
	for _, assist := range assists {
		hits := make([]assistHitRow, 0, len(assist.Hits))
		for _, hit := range assist.Hits {
			hits = append(hits, assistHitRow{
				Weapon: dashIfEmpty(hit.Weapon),
				Range:  rangeDisplay(hit.RangeMeters),
			})
		}
		rows = append(rows, assistRow{
			Timestamp: assist.Timestamp,
			Killer:    r.nameForPlayer(assist.KillerID, r.playerRelativeHref(assist.KillerID), "", true),
			Victim:    r.nameForPlayer(assist.VictimID, r.playerRelativeHref(assist.VictimID), "", true),
			Hits:      hits,
		})
	}

	return rows
}

func (r *renderer) environmentalRows(events []eventbuild.PlayerEnvironmentalDeath) []environmentalRow {
	rows := make([]environmentalRow, 0, len(events))
	for _, event := range events {
		rows = append(rows, environmentalRow{
			Timestamp: event.Timestamp,
			Cause:     event.Cause,
		})
	}

	return rows
}

func (r *renderer) nameForPlayer(playerID string, href string, team string, includeTeam bool) nameData {
	if player, ok := r.playerReports[playerID]; ok {
		meta := ""
		if includeTeam {
			meta = player.TeamName
		}
		if team != "" {
			meta = team
		}
		return nameData{
			Display: player.DisplayName,
			Meta:    meta,
			Href:    href,
		}
	}

	return nameData{
		Display: playerID,
		Meta:    team,
		Href:    href,
	}
}

func (r *renderer) playerRelativeHref(playerID string) string {
	return r.playerFile(playerID)
}

func (r *renderer) playerFile(playerID string) string {
	if file, ok := r.playerFiles[playerID]; ok {
		return file
	}
	return safeIDFileName(playerID) + ".html"
}

func (r *renderer) teamFile(teamName string) string {
	if file, ok := r.teamFiles[teamName]; ok {
		return file
	}
	return slugify(teamName) + ".html"
}

func playerSubheading(player eventbuild.PlayerReport) string {
	if player.TeamName == "" {
		return "Player detail page"
	}

	return player.TeamName
}

func loadJSON(path string, target any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(content, target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func dashIfEmpty(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}

func rangeDisplay(value *float64) string {
	if value == nil {
		return "—"
	}
	return fmt.Sprintf("%.2f m", *value)
}

func teamkillNote(explicitTeamkill bool, rowTeamkill bool) string {
	if explicitTeamkill || rowTeamkill {
		return "Teamkill"
	}
	return ""
}

func breakdownRows(values map[string]int) []breakdownRow {
	rows := make([]breakdownRow, 0, len(values))
	for label, count := range values {
		rows = append(rows, breakdownRow{Label: label, Count: count})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count != rows[j].Count {
			return rows[i].Count > rows[j].Count
		}
		return rows[i].Label < rows[j].Label
	})
	return rows
}

func safeIDFileName(value string) string {
	return hex.EncodeToString([]byte(value))
}

func slugify(value string) string {
	var builder strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}
	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		return "team"
	}
	return slug
}

func readAsset(path string) ([]byte, error) {
	return fs.ReadFile(assetsFS, path)
}
