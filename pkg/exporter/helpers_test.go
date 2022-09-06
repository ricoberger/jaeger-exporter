package exporter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDurationToMillis(t *testing.T) {
	require.Equal(t, float64(1000), durationToMillis(time.Duration(1*time.Second)))
	require.Equal(t, float64(1), durationToMillis(time.Duration(1*time.Millisecond)))
	require.Equal(t, float64(0.1), durationToMillis(time.Duration(100*time.Microsecond)))
}
