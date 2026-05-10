package parser_test

import (
	"testing"

	"github.com/jmhobbs/dayz-event-summary/internal/parser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Connecting(t *testing.T) {
	line, err := parser.ParseLine([]byte(`15:34:08 | Player "jmhobbs" (id=BFDIjY8X3a21fFxgXW339fE0IsxOK0kI7xlfig3HN1I=) is connecting`), true)
	require.NoError(t, err)
	require.NotNil(t, line)
	assert.Equal(t, line.Timestamp, "15:34:08")
	assert.Equal(t, &parser.LogLine{
		Timestamp: "15:34:08",
		Type:      "CONNECTING",
		Subject: &parser.Player{
			ID:   "BFDIjY8X3a21fFxgXW339fE0IsxOK0kI7xlfig3HN1I=",
			Name: "jmhobbs",
		},
	}, line)
}

func Test_Connected(t *testing.T) {
	line, err := parser.ParseLine([]byte(`15:34:20 | Player "jmhobbs" (id=BFDIjY8X3a21fFxgXW339fE0IsxOK0kI7xlfig3HN1I= pos=<17491.9, 6849.4, 16.1>) is connected`), true)
	require.NoError(t, err)
	require.NotNil(t, line)
	assert.Equal(t, &parser.LogLine{
		Timestamp: "15:34:20",
		Type:      "CONNECTED",
		Subject: &parser.Player{
			ID:   "BFDIjY8X3a21fFxgXW339fE0IsxOK0kI7xlfig3HN1I=",
			Name: "jmhobbs",
			Position: &parser.Position{
				X: 17491.9,
				Y: 6849.4,
				Z: 16.1,
			},
		},
	}, line)
}

func Test_Emote(t *testing.T) {
	line, err := parser.ParseLine([]byte(` 15:34:20 | Player "jmhobbs" (id=BFDIjY8X3a21fFxgXW339fE0IsxOK0kI7xlfig3HN1I= pos=<17491.9, 6849.4, 16.1>) performed EmoteLyingDown`), true)
	require.NoError(t, err)
	require.NotNil(t, line)
	assert.Equal(t, &parser.LogLine{
		Timestamp: "15:34:20",
		Type:      "EMOTE",
		Subject: &parser.Player{
			ID:   "BFDIjY8X3a21fFxgXW339fE0IsxOK0kI7xlfig3HN1I=",
			Name: "jmhobbs",
			Position: &parser.Position{
				X: 17491.9,
				Y: 6849.4,
				Z: 16.1,
			},
		},
	}, line)
}

// 15:42:16 | Player "[California Quails] Scoobies" (id=_ZVUA_p34dhRRLkvtNfloCdXF79HymmqyrwMUvE7-mk= pos=<17496.5, 6850.3, 16.8>) performed EmoteSurrender with SurrenderDummyItem

func Test_HitByPlayer(t *testing.T) {
	line, err := parser.ParseLine([]byte(`15:42:47 | Player "{ALIGATOR LIZARDS} READER_s_" (id=FDionF58K332TKe7JPTkPEGXrWjobdaWIfjaThYrqEc= pos=<17486.9, 6848.6, 17.0>)[HP: 90] hit by Player "Muller" (id=RAE6__7kNQmI4tizLgS6cLxYirk4d_D_bx58sagF-Iw= pos=<17487.5, 6849.3, 16.8>) into Head(0) for 5 damage (MeleeFist)`), true)
	require.NoError(t, err)
	require.NotNil(t, line)
	assert.Equal(t, &parser.LogLine{
		Timestamp: "15:42:47",
		Type:      "HIT",
		Subject: &parser.Player{
			ID:   "FDionF58K332TKe7JPTkPEGXrWjobdaWIfjaThYrqEc=",
			Name: "{ALIGATOR LIZARDS} READER_s_",
			Position: &parser.Position{
				X: 17486.9,
				Y: 6848.6,
				Z: 17.0,
			},
		},
		Object: &parser.Player{
			ID:   "RAE6__7kNQmI4tizLgS6cLxYirk4d_D_bx58sagF-Iw=",
			Name: "Muller",
			Position: &parser.Position{
				X: 17487.5,
				Y: 6849.3,
				Z: 16.8,
			},
		},
	}, line)
}
