package log

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	logger1, err := New("debug", "json")
	require.NoError(t, err)
	require.NotNil(t, logger1)
}
