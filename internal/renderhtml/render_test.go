package renderhtml

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"amd-report/internal/eventbuild"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderGeneratesStaticSite(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	outputDir := t.TempDir()

	writeJSONFile(t, filepath.Join(dataDir, "metadata.json"), eventbuild.Metadata{
		EventName:           "Test Event",
		SourceADM:           "sample.ADM",
		WindowStart:         "15:40:00",
		WindowEnd:           "16:10:00",
		AssistWindowSeconds: 30,
		GeneratedAt:         "2026-05-05T18:00:00Z",
		PlayerCount:         2,
		TeamCount:           1,
	})
	writeJSONFile(t, filepath.Join(dataDir, "roster.json"), eventbuild.RosterFile{})
	writeJSONFile(t, filepath.Join(dataDir, "events.json"), eventbuild.EventsFile{})
	writeJSONFile(t, filepath.Join(dataDir, "players.json"), eventbuild.PlayersFile{
		Players: []eventbuild.PlayerReport{
			{
				PlayerID:      "grouped-player",
				PreferredName: "Alpha One",
				DisplayName:   "One",
				TeamName:      "Alpha",
				Stats: eventbuild.PlayerStats{
					Kills:        2,
					Deaths:       1,
					Assists:      1,
					Teamkills:    0,
					RatioDisplay: "2.00",
				},
				Kills: []eventbuild.PlayerKillDetail{
					{Timestamp: "15:41:00", VictimID: "solo-player", VictimDisplayName: "Solo Survivor", Weapon: "Pioneer"},
				},
				Deaths: []eventbuild.PlayerDeathDetail{
					{Timestamp: "15:45:00", KillerID: "solo-player", KillerDisplayName: "Solo Survivor", Weapon: "M70", CountsTowardStats: true},
				},
				Assists: []eventbuild.PlayerAssistDetail{
					{
						Timestamp:         "15:43:00",
						KillerID:          "solo-player",
						KillerDisplayName: "Solo Survivor",
						VictimID:          "grouped-player",
						VictimDisplayName: "Alpha One",
						Hits: []eventbuild.AssistHitDetail{
							{Timestamp: "15:41:00", LineNumber: 12, Weapon: "Pioneer", RangeMeters: floatPtr(215.5)},
							{Timestamp: "15:42:00", LineNumber: 13, Weapon: "BK-18", RangeMeters: floatPtr(87.2)},
						},
					},
				},
				HitsDealt: eventbuild.HitBreakdown{Total: 3, ByBodyPart: map[string]int{"Head": 2, "Torso": 1}},
				HitsTaken: eventbuild.HitBreakdown{Total: 1, ByBodyPart: map[string]int{"Torso": 1}},
			},
			{
				PlayerID:      "solo-player",
				PreferredName: "Solo Survivor",
				DisplayName:   "Solo Survivor",
				Stats: eventbuild.PlayerStats{
					Kills:               1,
					Deaths:              0,
					Assists:             0,
					Teamkills:           0,
					EnvironmentalDeaths: 1,
					RatioDisplay:        "1.00",
				},
				EnvironmentalDeaths: []eventbuild.PlayerEnvironmentalDeath{
					{Timestamp: "15:50:00", Cause: "suicide"},
				},
				HitsDealt: eventbuild.HitBreakdown{Total: 1, ByBodyPart: map[string]int{"Torso": 1}},
				HitsTaken: eventbuild.HitBreakdown{Total: 0, ByBodyPart: map[string]int{}},
			},
			{
				PlayerID:      "player-3",
				PreferredName: "Player 3",
				DisplayName:   "Player 3",
				TeamName:      "Bravo",
				Stats:         eventbuild.PlayerStats{Kills: 5, Deaths: 1, RatioDisplay: "5.00"},
				HitsDealt:     eventbuild.HitBreakdown{ByBodyPart: map[string]int{}},
				HitsTaken:     eventbuild.HitBreakdown{ByBodyPart: map[string]int{}},
			},
			{
				PlayerID:      "player-4",
				PreferredName: "Player 4",
				DisplayName:   "Player 4",
				TeamName:      "Charlie",
				Stats:         eventbuild.PlayerStats{Kills: 4, Deaths: 1, RatioDisplay: "4.00"},
				HitsDealt:     eventbuild.HitBreakdown{ByBodyPart: map[string]int{}},
				HitsTaken:     eventbuild.HitBreakdown{ByBodyPart: map[string]int{}},
			},
			{
				PlayerID:      "player-5",
				PreferredName: "Player 5",
				DisplayName:   "Player 5",
				TeamName:      "Delta",
				Stats:         eventbuild.PlayerStats{Kills: 3, Deaths: 1, RatioDisplay: "3.00"},
				HitsDealt:     eventbuild.HitBreakdown{ByBodyPart: map[string]int{}},
				HitsTaken:     eventbuild.HitBreakdown{ByBodyPart: map[string]int{}},
			},
			{
				PlayerID:      "player-6",
				PreferredName: "Player 6",
				DisplayName:   "Player 6",
				TeamName:      "Echo",
				Stats:         eventbuild.PlayerStats{Kills: 0, Deaths: 1, RatioDisplay: "0.00"},
				HitsDealt:     eventbuild.HitBreakdown{ByBodyPart: map[string]int{}},
				HitsTaken:     eventbuild.HitBreakdown{ByBodyPart: map[string]int{}},
			},
		},
	})
	writeJSONFile(t, filepath.Join(dataDir, "teams.json"), eventbuild.TeamsFile{
		Teams: []eventbuild.TeamReport{
			{
				Name: "Alpha",
				Stats: eventbuild.TeamStats{
					Kills:        2,
					Deaths:       1,
					Assists:      1,
					Teamkills:    0,
					RatioDisplay: "2.00",
				},
				Members: []eventbuild.TeamMemberSummary{
					{
						PlayerID:      "grouped-player",
						PreferredName: "Alpha One",
						DisplayName:   "One",
						Kills:         2,
						Deaths:        1,
						Assists:       1,
						RatioDisplay:  "2.00",
					},
				},
				Matchups: []eventbuild.TeamMatchupSummary{
					{OpponentTeamName: "Unaffiliated", Kills: 2, Deaths: 1, RatioDisplay: "2.00"},
				},
			},
			{Name: "Bravo", Stats: eventbuild.TeamStats{Kills: 6, Deaths: 1, RatioDisplay: "6.00"}},
			{Name: "Charlie", Stats: eventbuild.TeamStats{Kills: 5, Deaths: 1, RatioDisplay: "5.00"}},
			{Name: "Delta", Stats: eventbuild.TeamStats{Kills: 4, Deaths: 1, RatioDisplay: "4.00"}},
			{Name: "Echo", Stats: eventbuild.TeamStats{Kills: 3, Deaths: 1, RatioDisplay: "3.00"}},
			{Name: "Foxtrot", Stats: eventbuild.TeamStats{Kills: 2, Deaths: 1, RatioDisplay: "2.00"}},
		},
	})
	writeJSONFile(t, filepath.Join(dataDir, "summary.json"), eventbuild.SummaryFile{
		PlayerTable: []eventbuild.SummaryPlayerRow{
			{PlayerID: "player-3", PreferredName: "Player 3", DisplayName: "Player 3", TeamName: "Bravo", Kills: 5, Deaths: 1, RatioDisplay: "5.00"},
			{PlayerID: "player-4", PreferredName: "Player 4", DisplayName: "Player 4", TeamName: "Charlie", Kills: 4, Deaths: 1, RatioDisplay: "4.00"},
			{PlayerID: "player-5", PreferredName: "Player 5", DisplayName: "Player 5", TeamName: "Delta", Kills: 3, Deaths: 1, RatioDisplay: "3.00"},
			{PlayerID: "grouped-player", PreferredName: "Alpha One", DisplayName: "One", TeamName: "Alpha", Kills: 2, Deaths: 1, Assists: 1, RatioDisplay: "2.00"},
			{PlayerID: "solo-player", PreferredName: "Solo Survivor", DisplayName: "Solo Survivor", Kills: 1, Deaths: 0, RatioDisplay: "1.00"},
			{PlayerID: "player-6", PreferredName: "Player 6", DisplayName: "Player 6", TeamName: "Echo", Kills: 0, Deaths: 1, RatioDisplay: "0.00"},
		},
		TeamTable: []eventbuild.SummaryTeamRow{
			{Name: "Bravo", Kills: 6, Deaths: 1, RatioDisplay: "6.00"},
			{Name: "Charlie", Kills: 5, Deaths: 1, RatioDisplay: "5.00"},
			{Name: "Delta", Kills: 4, Deaths: 1, RatioDisplay: "4.00"},
			{Name: "Echo", Kills: 3, Deaths: 1, RatioDisplay: "3.00"},
			{Name: "Alpha", Kills: 2, Deaths: 1, Assists: 1, RatioDisplay: "2.00"},
			{Name: "Foxtrot", Kills: 1, Deaths: 1, RatioDisplay: "1.00"},
		},
		LongestHits: []eventbuild.ShotSummary{
			{Timestamp: "15:41:00", AttackerID: "grouped-player", AttackerDisplayName: "One", VictimID: "solo-player", VictimDisplayName: "Solo Survivor", Weapon: "Pioneer", RangeMeters: 220.12, BodyPart: "Head", IsKill: true},
			{Timestamp: "15:41:10", AttackerID: "player-3", AttackerDisplayName: "Player 3", VictimID: "player-4", VictimDisplayName: "Player 4", Weapon: "Savanna", RangeMeters: 205.34, BodyPart: "Torso", IsKill: false},
			{Timestamp: "15:41:20", AttackerID: "player-4", AttackerDisplayName: "Player 4", VictimID: "player-5", VictimDisplayName: "Player 5", Weapon: "LAR", RangeMeters: 190.78, BodyPart: "Arm", IsKill: true},
			{Timestamp: "15:41:30", AttackerID: "player-5", AttackerDisplayName: "Player 5", VictimID: "player-6", VictimDisplayName: "Player 6", Weapon: "Blaze", RangeMeters: 175.67, BodyPart: "Leg", IsKill: false},
			{Timestamp: "15:41:40", AttackerID: "player-6", AttackerDisplayName: "Player 6", VictimID: "grouped-player", VictimDisplayName: "One", Weapon: "Mosin", RangeMeters: 160.89, BodyPart: "Head", IsKill: false},
			{Timestamp: "15:40:50", AttackerID: "solo-player", AttackerDisplayName: "Solo Survivor", VictimID: "grouped-player", VictimDisplayName: "One", Weapon: "BK-18", RangeMeters: 140.55, BodyPart: "Torso", IsKill: false},
		},
		LongestKills: []eventbuild.ShotSummary{
			{Timestamp: "15:45:00", AttackerID: "solo-player", AttackerDisplayName: "Solo Survivor", VictimID: "grouped-player", VictimDisplayName: "One", Weapon: "M70", RangeMeters: 300.91, IsKill: true},
			{Timestamp: "15:45:10", AttackerID: "grouped-player", AttackerDisplayName: "One", VictimID: "solo-player", VictimDisplayName: "Solo Survivor", Weapon: "Tundra", RangeMeters: 280.75, IsKill: true},
			{Timestamp: "15:45:20", AttackerID: "player-3", AttackerDisplayName: "Player 3", VictimID: "player-4", VictimDisplayName: "Player 4", Weapon: "Blaze", RangeMeters: 260.62, IsKill: true},
			{Timestamp: "15:45:30", AttackerID: "player-4", AttackerDisplayName: "Player 4", VictimID: "player-5", VictimDisplayName: "Player 5", Weapon: "VSD", RangeMeters: 240.48, IsKill: true},
			{Timestamp: "15:45:40", AttackerID: "player-5", AttackerDisplayName: "Player 5", VictimID: "player-6", VictimDisplayName: "Player 6", Weapon: "LAR", RangeMeters: 220.31, IsKill: true},
			{Timestamp: "15:45:50", AttackerID: "player-6", AttackerDisplayName: "Player 6", VictimID: "player-3", VictimDisplayName: "Player 3", Weapon: "M16", RangeMeters: 180.44, IsKill: true},
		},
	})

	bundle, err := LoadBundle(dataDir)
	require.NoError(t, err)
	require.NoError(t, Render(bundle, outputDir))

	indexContent := readFile(t, filepath.Join(outputDir, "index.html"))
	playersIndexContent := readFile(t, filepath.Join(outputDir, "players", "index.html"))
	teamsIndexContent := readFile(t, filepath.Join(outputDir, "teams", "index.html"))
	hitsIndexContent := readFile(t, filepath.Join(outputDir, "hits", "index.html"))
	killsIndexContent := readFile(t, filepath.Join(outputDir, "kills", "index.html"))
	groupedContent := readFile(t, filepath.Join(outputDir, "players", safeIDFileName("grouped-player")+".html"))
	teamContent := readFile(t, filepath.Join(outputDir, "teams", slugify("Alpha")+".html"))
	cssContent := readFile(t, filepath.Join(outputDir, "assets", "style.css"))

	assert.Contains(t, indexContent, "Test Event")
	assert.Contains(t, indexContent, "One")
	assert.Contains(t, indexContent, "Solo Survivor")
	assert.Contains(t, indexContent, "players/"+safeIDFileName("grouped-player")+".html")
	assert.Contains(t, indexContent, "teams/"+slugify("Alpha")+".html")
	assert.Contains(t, indexContent, "players/index.html")
	assert.Contains(t, indexContent, "teams/index.html")
	assert.Contains(t, indexContent, "hits/index.html")
	assert.Contains(t, indexContent, "kills/index.html")
	assert.Contains(t, indexContent, ">Players<")
	assert.Contains(t, indexContent, ">Teams<")
	assert.Contains(t, indexContent, ">Hits<")
	assert.Contains(t, indexContent, ">All Kills<")
	assert.Contains(t, indexContent, "View All")
	assert.Contains(t, indexContent, "Result")
	assert.Contains(t, indexContent, "Kill")
	assert.Contains(t, indexContent, "Hit")
	assert.Contains(t, indexContent, "220.12 m")
	assert.Contains(t, indexContent, "160.89 m")
	assert.NotContains(t, indexContent, "140.55 m")
	assert.Contains(t, indexContent, "300.91 m")
	assert.Contains(t, indexContent, "220.31 m")
	assert.NotContains(t, indexContent, "180.44 m")
	assert.NotContains(t, indexContent, "Foxtrot")
	assert.NotContains(t, indexContent, "Event details")
	assert.NotContains(t, indexContent, "entity-secondary")
	assert.True(t, strings.Index(indexContent, "href=\"teams/index.html\">View All") < strings.Index(indexContent, "href=\"players/index.html\">View All"))

	assert.Contains(t, playersIndexContent, "Full player leaderboard.")
	assert.Contains(t, playersIndexContent, "Player 6")
	assert.Contains(t, playersIndexContent, "index.html\">Players<")
	assert.Contains(t, playersIndexContent, "../hits/index.html\">Hits<")
	assert.Contains(t, playersIndexContent, "../kills/index.html\">All Kills<")
	assert.NotContains(t, playersIndexContent, "View All")

	assert.Contains(t, teamsIndexContent, "Full team leaderboard.")
	assert.Contains(t, teamsIndexContent, "Foxtrot")
	assert.Contains(t, teamsIndexContent, "index.html\">Teams<")
	assert.Contains(t, teamsIndexContent, "../hits/index.html\">Hits<")
	assert.Contains(t, teamsIndexContent, "../kills/index.html\">All Kills<")
	assert.NotContains(t, teamsIndexContent, "View All")

	assert.Contains(t, hitsIndexContent, "All ranged hits, sorted longest first.")
	assert.Contains(t, hitsIndexContent, "140.55 m")
	assert.Contains(t, hitsIndexContent, "index.html\">Hits<")
	assert.Contains(t, hitsIndexContent, "../kills/index.html\">All Kills<")
	assert.NotContains(t, hitsIndexContent, "View All")

	assert.Contains(t, killsIndexContent, "All ranged kills, sorted longest first.")
	assert.Contains(t, killsIndexContent, "180.44 m")
	assert.Contains(t, killsIndexContent, "../hits/index.html\">Hits<")
	assert.Contains(t, killsIndexContent, "index.html\">All Kills<")
	assert.NotContains(t, killsIndexContent, "View All")

	assert.Contains(t, groupedContent, "One")
	assert.Contains(t, groupedContent, "../assets/style.css")
	assert.Contains(t, groupedContent, "Solo Survivor")
	assert.Contains(t, groupedContent, "Pioneer")
	assert.Contains(t, groupedContent, "215.50 m")
	assert.Contains(t, groupedContent, "BK-18")
	assert.Contains(t, groupedContent, "87.20 m")
	assert.Contains(t, groupedContent, "Alpha")
	assert.NotContains(t, groupedContent, ">Back<")
	assert.Contains(t, groupedContent, "index.html\">Players<")
	assert.Contains(t, groupedContent, "../teams/index.html\">Teams<")
	assert.Contains(t, groupedContent, "../hits/index.html\">Hits<")
	assert.Contains(t, groupedContent, "../kills/index.html\">All Kills<")
	assert.NotContains(t, groupedContent, "Team Alpha")
	assert.NotContains(t, groupedContent, "Alpha One")
	assert.NotContains(t, groupedContent, "from hit line 12")
	assert.NotContains(t, groupedContent, "entity-secondary")

	assert.Contains(t, teamContent, "Alpha")
	assert.Contains(t, teamContent, "One")
	assert.Contains(t, teamContent, "../players/"+safeIDFileName("grouped-player")+".html")
	assert.NotContains(t, teamContent, ">Back<")
	assert.Contains(t, teamContent, "../players/index.html\">Players<")
	assert.Contains(t, teamContent, "index.html\">Teams<")
	assert.Contains(t, teamContent, "../hits/index.html\">Hits<")
	assert.Contains(t, teamContent, "../kills/index.html\">All Kills<")
	assert.NotContains(t, teamContent, "entity-secondary")
	assert.NotContains(t, teamContent, "entity-meta")

	assert.Contains(t, cssContent, "--color-bg")
	assert.Contains(t, cssContent, "--space-4")
}

func writeJSONFile(t *testing.T, path string, value any) {
	t.Helper()

	rendered, err := json.Marshal(value)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, rendered, 0o644))
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}

func floatPtr(value float64) *float64 {
	return &value
}
