package internal

import (
	"log/slog"
	"time"
)

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	slog.Debug(name, "took", elapsed)
}
