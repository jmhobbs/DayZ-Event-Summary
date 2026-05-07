package teamguess

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractPlayersAppliesWindowAndCanonicalIdentity(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(strings.Join([]string{
		"******************************************************************************",
		"AdminLog started on 2026-05-02 at 15:33:25",
		`15:34:08 | Player "outside" (id=outside-id) is connecting`,
		`15:40:00 | Player "[Catalina Island Foxes] KnightZ" (id=fox-1) is connecting`,
		`15:40:01 | Player "[Catalina Island Foxes] KnightZ" (id=fox-1 pos=<1, 2, 3>) is connected`,
		`15:40:10 | Chat("Fox Leader"(id=fox-1)): ready`,
		`15:40:20 | Player "SPR-Snikende" (id=spr-1 pos=<1, 1, 1>)[HP: 50] hit by Player "SPR - Doxy" (id=spr-2 pos=<2, 2, 2>) into Torso(6) for 33 damage (Bullet_556x45) with Pioneer from 74.1937 meters`,
		`15:46:00 | Player "outside after" (id=outside-after) is connecting`,
	}, "\n"))

	window := Window{
		Start: 15*time.Hour + 40*time.Minute,
		End:   15*time.Hour + 45*time.Minute,
	}

	players, err := ExtractPlayers(input, window)
	require.NoError(t, err)
	require.Len(t, players, 3)

	assert.Equal(t, "fox-1", players[0].PlayerID)
	assert.Equal(t, "Fox Leader", players[0].LatestName)
	assert.ElementsMatch(t, []string{"[Catalina Island Foxes] KnightZ", "Fox Leader"}, players[0].SeenNames)

	assert.Equal(t, "spr-1", players[1].PlayerID)
	assert.Equal(t, "SPR-Snikende", players[1].LatestName)

	assert.Equal(t, "spr-2", players[2].PlayerID)
	assert.Equal(t, "SPR - Doxy", players[2].LatestName)
}

func TestExtractPlayersAllIncludesPlayersOutsideWindow(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(strings.Join([]string{
		"******************************************************************************",
		"AdminLog started on 2026-05-02 at 15:33:25",
		`15:34:08 | Player "outside" (id=outside-id) is connecting`,
		`15:40:00 | Player "inside" (id=inside-id) is connecting`,
		`15:46:00 | Player "outside after" (id=outside-after) is connecting`,
	}, "\n"))

	players, err := ExtractPlayersAll(input)
	require.NoError(t, err)
	require.Len(t, players, 3)

	assert.Equal(t, "outside-id", players[0].PlayerID)
	assert.Equal(t, "inside-id", players[1].PlayerID)
	assert.Equal(t, "outside-after", players[2].PlayerID)
}

func TestSuggestBuildsConservativeGroups(t *testing.T) {
	t.Parallel()

	players := []Player{
		{PlayerID: "fox-1", LatestName: "[Catalina Island Foxes] KnightZ"},
		{PlayerID: "fox-2", LatestName: "[Catalina Island Foxes] [DONG RICKY]"},
		{PlayerID: "hawks-1", LatestName: "Cooper's Hawks-ColdarionMf"},
		{PlayerID: "hawks-2", LatestName: "[Cooper's Hawks] krypt0ne0n"},
		{PlayerID: "spr-1", LatestName: "SPR-Snikende"},
		{PlayerID: "spr-2", LatestName: "SPR - Doxy"},
		{PlayerID: "solo-1", LatestName: "Police CassInRuby"},
	}

	output := Suggest(players)
	require.Len(t, output.Teams, 3)

	teamsByName := map[string]TeamSuggestion{}
	for _, team := range output.Teams {
		teamsByName[team.Name] = team
	}

	assert.Contains(t, teamsByName, "Catalina Island Foxes")
	assert.Len(t, teamsByName["Catalina Island Foxes"].Members, 2)
	assert.Equal(t, "high", teamsByName["Catalina Island Foxes"].Confidence)
	foxMembers := membersByID(teamsByName["Catalina Island Foxes"].Members)
	assert.Equal(t, "KnightZ", foxMembers["fox-1"].DisplayName)
	assert.Equal(t, "[DONG RICKY]", foxMembers["fox-2"].DisplayName)

	assert.Contains(t, teamsByName, "Cooper's Hawks")
	assert.Len(t, teamsByName["Cooper's Hawks"].Members, 2)
	hawkMembers := membersByID(teamsByName["Cooper's Hawks"].Members)
	assert.Equal(t, "ColdarionMf", hawkMembers["hawks-1"].DisplayName)
	assert.Equal(t, "krypt0ne0n", hawkMembers["hawks-2"].DisplayName)

	assert.Contains(t, teamsByName, "SPR")
	assert.Len(t, teamsByName["SPR"].Members, 2)
	sprMembers := membersByID(teamsByName["SPR"].Members)
	assert.Equal(t, "Snikende", sprMembers["spr-1"].DisplayName)
	assert.Equal(t, "Doxy", sprMembers["spr-2"].DisplayName)

	require.Len(t, output.Ungrouped, 1)
	assert.Equal(t, "solo-1", output.Ungrouped[0].PlayerID)
	assert.Empty(t, output.Ungrouped[0].DisplayName)
}

func TestMarshalYAMLOutput(t *testing.T) {
	t.Parallel()

	output := Output{
		Teams: []TeamSuggestion{
			{
				Name:       "Catalina Island Foxes",
				Confidence: "high",
				Reason:     "shared leading tag across 2 players",
				Members: []Member{
					{PlayerID: "fox-1", PreferredName: "KnightZ", DisplayName: "KnightZ"},
				},
			},
		},
		Ungrouped: []Member{
			{PlayerID: "solo-1", PreferredName: "Police CassInRuby"},
		},
	}

	rendered, err := MarshalYAML(output)
	require.NoError(t, err)

	text := string(rendered)
	assert.Contains(t, text, "teams:")
	assert.Contains(t, text, "confidence: high")
	assert.Contains(t, text, "player_id: fox-1")
	assert.Contains(t, text, "display_name: KnightZ")
	assert.Contains(t, text, "ungrouped:")
	assert.Contains(t, text, "Police CassInRuby")
	assert.NotContains(t, text, "display_name: \"\"")
}

func membersByID(members []Member) map[string]Member {
	indexed := make(map[string]Member, len(members))
	for _, member := range members {
		indexed[member.PlayerID] = member
	}

	return indexed
}
