package parser_test

import (
	"testing"

	"github.com/jmhobbs/dayz-event-summary/internal/parser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Connecting(t *testing.T) {
	log, err := parser.Parse([]byte(`15:34:08 | Player "jmhobbs" (id=BFDIjY8X3a21fFxgXW339fE0IsxOK0kI7xlfig3HN1I=) is connecting`), true)
	require.NoError(t, err)
	assert.Equal(t, 1, len(log))
	assert.Equal(t, log[0], parser.LogLine{
		Timestamp: "15:34:08",
	})
}
