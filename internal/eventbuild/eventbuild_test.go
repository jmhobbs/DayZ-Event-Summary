package eventbuild

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmhobbs/dayz-event-summary/internal/teamguess"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEventSettingsAppliesDefaults(t *testing.T) {
	t.Parallel()

	settings, err := LoadEventSettings(strings.NewReader("event_name: Example Event\n"))
	require.NoError(t, err)

	assert.Equal(t, "Example Event", settings.EventName)
	assert.Equal(t, 30, settings.AssistWindowSeconds)
}

func TestLoadTeamConfigReadsIgnoredSection(t *testing.T) {
	t.Parallel()

	config, err := LoadTeamConfig(strings.NewReader(strings.Join([]string{
		"teams: []",
		"ungrouped: []",
		"ignored:",
		"  - player_id: ignored-1",
		"    preferred_name: Ignored Player",
	}, "\n")))
	require.NoError(t, err)
	require.Len(t, config.Ignored, 1)
	assert.Equal(t, "ignored-1", config.Ignored[0].PlayerID)
}

func TestBuildCreatesNormalizedBundle(t *testing.T) {
	t.Parallel()

	settings, err := LoadEventSettings(strings.NewReader(strings.Join([]string{
		"event_name: Example Event",
		"assist_window_seconds: 30",
	}, "\n")))
	require.NoError(t, err)

	teams, err := LoadTeamConfig(strings.NewReader(strings.Join([]string{
		"teams:",
		"  - name: Alpha",
		"    members:",
		"      - player_id: alpha-1",
		"        preferred_name: Alpha One",
		"        display_name: One",
		"      - player_id: alpha-2",
		"        preferred_name: Alpha Two",
		"        display_name: Two",
		"  - name: Bravo",
		"    members:",
		"      - player_id: bravo-1",
		"        preferred_name: Bravo Guy",
		"        display_name: Guy",
		"ungrouped:",
		"  - player_id: solo-1",
		"    preferred_name: Solo",
	}, "\n")))
	require.NoError(t, err)

	log := strings.Join([]string{
		`15:40:00 | Player "Alpha One" (id=alpha-1) is connecting`,
		`15:40:01 | Player "Alpha Two" (id=alpha-2) is connecting`,
		`15:40:02 | Player "Bravo Guy" (id=bravo-1) is connecting`,
		`15:40:03 | Player "Solo" (id=solo-1) is connecting`,
		`15:40:04 | Player "Wild Victim" (id=wild-1) is connecting`,
		`15:40:10 | Player "Bravo Guy" (id=bravo-1 pos=<1,1,1>)[HP: 90] hit by Player "Alpha One" (id=alpha-1 pos=<2,2,2>) into Torso(1) for 10 damage (Bullet_556x45) with Pioneer from 20.0 meters`,
		`15:40:15 | Player "Bravo Guy" (id=bravo-1 pos=<1,1,1>)[HP: 80] hit by Player "Alpha One" (id=alpha-1 pos=<2,2,2>) into LeftArm(2) for 10 damage (Bullet_556x45) with BK-18 from 120.0 meters`,
		`15:40:20 | Player "Bravo Guy" (id=bravo-1 pos=<1,1,1>)[HP: 50] hit by Player "Alpha Two" (id=alpha-2 pos=<2,2,2>) into Head(0) for 40 damage (Bullet_556x45) with Pioneer from 18.0 meters`,
		`15:40:25 | Player "Bravo Guy" (DEAD) (id=bravo-1 pos=<1,1,1>) killed by Player "Alpha Two" (id=alpha-2 pos=<2,2,2>) with Pioneer from 18.0 meters`,
		`15:40:30 | Player "Alpha One" (id=alpha-1 pos=<1,1,1>)[HP: 50] hit by Player "Alpha Two" (id=alpha-2 pos=<2,2,2>) into Torso(1) for 60 damage (Bullet_556x45) with Pioneer from 5.0 meters`,
		`15:40:31 | Player "Alpha One" (DEAD) (id=alpha-1 pos=<1,1,1>) killed by Player "Alpha Two" (id=alpha-2 pos=<2,2,2>) with Pioneer from 5.0 meters`,
		`15:40:32 | Player "Alpha One" (DEAD) (id=alpha-1 pos=<1,1,1>)[HP: 0] hit by Player "Bravo Guy" (id=bravo-1 pos=<2,2,2>) into Head(0) for 10 damage (Bullet_556x45) with Pioneer from 8.0 meters`,
		`15:40:40 | Player "Solo" (id=solo-1 pos=<3,3,3>) committed suicide`,
		`15:40:41 | Player "Solo" (DEAD) (id=solo-1 pos=<3,3,3>) died. Stats> Water: 0 Energy: 0 Bleed sources: 0`,
		`15:40:42 | Player "Wild Victim" (DEAD) (id=wild-1 pos=<3,3,3>) killed by Animal_CanisLupus_White`,
		`15:40:43 | Player "Wild Victim" (DEAD) (id=wild-1 pos=<3,3,3>) died. Stats> Water: 0 Energy: 0 Bleed sources: 0`,
		`15:40:45 | Player "Bravo Guy" (id=bravo-1 pos=<3,3,3>) hit by Player "Alpha One" (id=alpha-1`,
	}, "\n")

	bundle, err := Build(strings.NewReader(log), "sample.ADM", BuildOptions{
		Window: teamguess.Window{
			Start: 15*time.Hour + 40*time.Minute,
			End:   15*time.Hour + 41*time.Minute,
		},
		WindowStart: "15:40:00",
		WindowEnd:   "15:41:00",
		Settings:    settings,
		Teams:       teams,
		GeneratedAt: time.Date(2026, 5, 5, 16, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	assert.Equal(t, "Example Event", bundle.Metadata.EventName)
	assert.Equal(t, 5, bundle.Metadata.PlayerCount)
	assert.Equal(t, 2, bundle.Metadata.TeamCount)
	assert.Len(t, bundle.Events.Hits, 4)
	assert.Len(t, bundle.Events.Kills, 2)
	assert.Len(t, bundle.Events.Assists, 1)
	assert.Len(t, bundle.Events.EnvironmentalDeaths, 2)
	assert.Len(t, bundle.Warnings, 1)

	rosterByID := map[string]RosterPlayer{}
	for _, player := range bundle.Roster.Players {
		rosterByID[player.PlayerID] = player
	}
	assert.Equal(t, "One", rosterByID["alpha-1"].DisplayName)
	assert.Equal(t, "Alpha One", rosterByID["alpha-1"].PreferredName)
	assert.Equal(t, "Guy", rosterByID["bravo-1"].DisplayName)

	playerByID := map[string]PlayerReport{}
	for _, player := range bundle.Players.Players {
		playerByID[player.PlayerID] = player
	}
	assert.Equal(t, 0, playerByID["alpha-1"].Stats.Deaths)
	assert.Len(t, playerByID["alpha-1"].Deaths, 1)
	assert.True(t, playerByID["alpha-1"].Deaths[0].Teamkill)
	assert.Equal(t, 1, playerByID["alpha-1"].Stats.Assists)
	require.Len(t, playerByID["alpha-1"].Assists, 1)
	require.Len(t, playerByID["alpha-1"].Assists[0].Hits, 2)
	assert.Equal(t, "Pioneer", playerByID["alpha-1"].Assists[0].Hits[0].Weapon)
	require.NotNil(t, playerByID["alpha-1"].Assists[0].Hits[0].RangeMeters)
	assert.Equal(t, 20.0, *playerByID["alpha-1"].Assists[0].Hits[0].RangeMeters)
	assert.Equal(t, "BK-18", playerByID["alpha-1"].Assists[0].Hits[1].Weapon)
	require.NotNil(t, playerByID["alpha-1"].Assists[0].Hits[1].RangeMeters)
	assert.Equal(t, 120.0, *playerByID["alpha-1"].Assists[0].Hits[1].RangeMeters)

	assert.Equal(t, 1, playerByID["alpha-2"].Stats.Kills)
	assert.Equal(t, 1, playerByID["alpha-2"].Stats.Teamkills)
	assert.Equal(t, 2, playerByID["alpha-2"].HitsDealt.Total)

	assert.Equal(t, 1, playerByID["bravo-1"].Stats.Deaths)
	summaryByID := map[string]SummaryPlayerRow{}
	for _, row := range bundle.Summary.PlayerTable {
		summaryByID[row.PlayerID] = row
	}
	assert.Equal(t, "1.00", summaryByID["alpha-2"].RatioDisplay)
	assert.Equal(t, "0.00", summaryByID["bravo-1"].RatioDisplay)

	assert.Equal(t, 1, playerByID["solo-1"].Stats.EnvironmentalDeaths)
	assert.Equal(t, "suicide", playerByID["solo-1"].EnvironmentalDeaths[0].Cause)
	assert.Equal(t, 1, playerByID["wild-1"].Stats.EnvironmentalDeaths)
	assert.Equal(t, "Animal_CanisLupus_White", playerByID["wild-1"].EnvironmentalDeaths[0].Cause)

	require.Len(t, bundle.Teams.Teams, 2)
	assert.Equal(t, "Alpha", bundle.Teams.Teams[0].Name)
	assert.Equal(t, 1, bundle.Teams.Teams[0].Stats.Kills)
	assert.Equal(t, 1, bundle.Teams.Teams[0].Stats.Assists)
	assert.Equal(t, 1, bundle.Teams.Teams[0].Stats.Teamkills)
	assert.Len(t, bundle.Teams.Teams[0].Matchups, 1)
	assert.Equal(t, "Bravo", bundle.Teams.Teams[0].Matchups[0].OpponentTeamName)

	require.Len(t, bundle.Summary.LongestHits, 4)
	assert.Equal(t, 120.0, bundle.Summary.LongestHits[0].RangeMeters)
	assert.False(t, bundle.Summary.LongestHits[0].IsKill)
	assert.True(t, bundle.Summary.LongestHits[2].IsKill)
	require.Len(t, bundle.Summary.LongestKills, 2)
	assert.Equal(t, 18.0, bundle.Summary.LongestKills[0].RangeMeters)
	assert.True(t, bundle.Summary.LongestKills[0].IsKill)
}

func TestBuildExcludesIgnoredUsersFromBundle(t *testing.T) {
	t.Parallel()

	settings, err := LoadEventSettings(strings.NewReader("event_name: Ignore Test\nassist_window_seconds: 30\n"))
	require.NoError(t, err)

	teams, err := LoadTeamConfig(strings.NewReader(strings.Join([]string{
		"teams: []",
		"ungrouped: []",
		"ignored:",
		"  - player_id: ignored-1",
		"    preferred_name: Ignored One",
	}, "\n")))
	require.NoError(t, err)

	log := strings.Join([]string{
		`15:40:00 | Player "Ignored One" (id=ignored-1) is connecting`,
		`15:40:01 | Player "Visible One" (id=visible-1) is connecting`,
		`15:40:10 | Player "Visible One" (id=visible-1 pos=<1,1,1>)[HP: 90] hit by Player "Ignored One" (id=ignored-1 pos=<2,2,2>) into Torso(1) for 10 damage (Bullet_556x45) with Pioneer from 20.0 meters`,
		`15:40:20 | Player "Ignored One" (DEAD) (id=ignored-1 pos=<1,1,1>) killed by Player "Visible One" (id=visible-1 pos=<2,2,2>) with Pioneer from 18.0 meters`,
		`15:40:25 | Player "Ignored One" (DEAD) (id=ignored-1 pos=<1,1,1>) died. Stats> Water: 0 Energy: 0 Bleed sources: 0`,
	}, "\n")

	bundle, err := Build(strings.NewReader(log), "sample.ADM", BuildOptions{
		Window: teamguess.Window{
			Start: 15*time.Hour + 40*time.Minute,
			End:   15*time.Hour + 41*time.Minute,
		},
		WindowStart: "15:40:00",
		WindowEnd:   "15:41:00",
		Settings:    settings,
		Teams:       teams,
		GeneratedAt: time.Date(2026, 5, 6, 21, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	assert.Equal(t, 1, bundle.Metadata.PlayerCount)
	require.Len(t, bundle.Roster.Players, 1)
	assert.Equal(t, "visible-1", bundle.Roster.Players[0].PlayerID)
	assert.Empty(t, bundle.Events.Hits)
	assert.Empty(t, bundle.Events.Kills)
	assert.Empty(t, bundle.Events.EnvironmentalDeaths)
	require.Len(t, bundle.Players.Players, 1)
	assert.Equal(t, "visible-1", bundle.Players.Players[0].PlayerID)
	assert.Equal(t, "Visible One", bundle.Summary.PlayerTable[0].DisplayName)
}

func TestBuildIncludesOutOfWindowPlayersInRosterAndReports(t *testing.T) {
	t.Parallel()

	settings, err := LoadEventSettings(strings.NewReader("event_name: Roster Scope Test\nassist_window_seconds: 30\n"))
	require.NoError(t, err)

	teams, err := LoadTeamConfig(strings.NewReader(strings.Join([]string{
		"teams: []",
		"ungrouped:",
		"  - player_id: pre-1",
		"    preferred_name: Pre Window",
	}, "\n")))
	require.NoError(t, err)

	log := strings.Join([]string{
		`15:39:00 | Player "Pre Window Raw" (id=pre-1) is connecting`,
		`15:40:10 | Player "Fighter" (id=fighter-1) is connecting`,
		`15:40:20 | Player "Target" (id=target-1) is connecting`,
		`15:40:30 | Player "Target" (id=target-1 pos=<1,1,1>)[HP: 90] hit by Player "Fighter" (id=fighter-1 pos=<2,2,2>) into Torso(1) for 10 damage (Bullet_556x45) with Pioneer from 20.0 meters`,
		`15:40:40 | Player "Target" (DEAD) (id=target-1 pos=<1,1,1>) killed by Player "Fighter" (id=fighter-1 pos=<2,2,2>) with Pioneer from 18.0 meters`,
	}, "\n")

	bundle, err := Build(strings.NewReader(log), "sample.ADM", BuildOptions{
		Window: teamguess.Window{
			Start: 15*time.Hour + 40*time.Minute,
			End:   15*time.Hour + 41*time.Minute,
		},
		WindowStart: "15:40:00",
		WindowEnd:   "15:41:00",
		Settings:    settings,
		Teams:       teams,
		GeneratedAt: time.Date(2026, 5, 6, 21, 15, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	assert.Equal(t, 3, bundle.Metadata.PlayerCount)
	require.Len(t, bundle.Roster.Players, 3)

	rosterByID := map[string]RosterPlayer{}
	for _, player := range bundle.Roster.Players {
		rosterByID[player.PlayerID] = player
	}
	assert.Equal(t, "Pre Window", rosterByID["pre-1"].PreferredName)
	assert.Equal(t, "Pre Window", rosterByID["pre-1"].DisplayName)

	playerByID := map[string]PlayerReport{}
	for _, player := range bundle.Players.Players {
		playerByID[player.PlayerID] = player
	}
	require.Contains(t, playerByID, "pre-1")
	assert.Equal(t, 0, playerByID["pre-1"].Stats.Kills)
	assert.Equal(t, 0, playerByID["pre-1"].Stats.Deaths)
	assert.Equal(t, 0, playerByID["pre-1"].Stats.Assists)
	assert.Equal(t, 0, playerByID["pre-1"].Stats.EnvironmentalDeaths)

	summaryByID := map[string]SummaryPlayerRow{}
	for _, row := range bundle.Summary.PlayerTable {
		summaryByID[row.PlayerID] = row
	}
	require.Contains(t, summaryByID, "pre-1")
	assert.Equal(t, 0, summaryByID["pre-1"].Kills)
	assert.Equal(t, 0, summaryByID["pre-1"].Deaths)
	assert.Equal(t, 0, summaryByID["pre-1"].Assists)
}

func TestWriteBundleWritesExpectedFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	bundle := &Bundle{
		Metadata: Metadata{SourceADM: "sample.ADM"},
		Roster:   RosterFile{Players: []RosterPlayer{{PlayerID: "p1", PreferredName: "Player 1", DisplayName: "Player 1"}}},
		Events:   EventsFile{},
		Players:  PlayersFile{},
		Teams:    TeamsFile{},
		Summary:  SummaryFile{},
	}

	require.NoError(t, WriteBundle(bundle, tempDir))

	for _, name := range []string{"metadata.json", "roster.json", "events.json", "players.json", "teams.json", "summary.json"} {
		_, err := os.Stat(filepath.Join(tempDir, name))
		require.NoError(t, err, name)
	}
}
