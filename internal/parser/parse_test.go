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
	assert.Equal(t, line, &parser.LogLine{
		Timestamp: "15:34:08",
		Type:      "CONNECTING",
		Player: &parser.Player{
			ID:   "BFDIjY8X3a21fFxgXW339fE0IsxOK0kI7xlfig3HN1I=",
			Name: "jmhobbs",
		},
	})
}

func Test_Connected(t *testing.T) {
	line, err := parser.ParseLine([]byte(`15:34:20 | Player "jmhobbs" (id=BFDIjY8X3a21fFxgXW339fE0IsxOK0kI7xlfig3HN1I= pos=<17491.9, 6849.4, 16.1>) is connected`), true)
	require.NoError(t, err)
	require.NotNil(t, line)
	assert.Equal(t, line, &parser.LogLine{
		Timestamp: "15:34:20",
		Type:      "CONNECTED",
		Player: &parser.Player{
			ID:   "BFDIjY8X3a21fFxgXW339fE0IsxOK0kI7xlfig3HN1I=",
			Name: "jmhobbs",
			Position: &parser.Position{
				X: 17491.9,
				Y: 6849.4,
				Z: 16.1,
			},
		},
	})
}
