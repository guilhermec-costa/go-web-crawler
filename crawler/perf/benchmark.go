package perf

import (
	"log/slog"
	"time"
)

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	slog.Info(name, "took", elapsed)
}
