package exporter

import (
	"time"
)

// durationToMillis converts the given duration to the number of milliseconds it represents. This can return
// sub-millisecond (i.e. < 1ms) values as well.
func durationToMillis(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / float64(time.Millisecond.Nanoseconds())
}
