package config

import (
	"time"
)

type Config interface {
	WorkDuration() time.Duration
	ShortBreakDuration() time.Duration
	LongBreakDuration() time.Duration
	PhasesPerCycle() int
	TotalCycles() int
}

const (
	WorkDurationKey       = "work-duration"
	ShortBreakDurationKey = "short-break-duration"
	LongBreakDurationKey  = "long-break-duration"
	PhasesPerCycleKey     = "phases-per-cycle"
	TotalCyclesKey        = "total-cycles"
)
