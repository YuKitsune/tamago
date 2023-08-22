package config

import (
	"time"
)

type Config interface {
	DryRun() bool
	WorkDuration() time.Duration
	ShortBreakDuration() time.Duration
	LongBreakDuration() time.Duration
	PhasesPerCycle() int
	TotalCycles() int
}

const (
	DryRunKey             = "dry-run"
	WorkDurationKey       = "work-duration"
	ShortBreakDurationKey = "short-break-duration"
	LongBreakDurationKey  = "long-break-duration"
	PhasesPerCycleKey     = "phases-per-cycle"
	TotalCyclesKey        = "total-cycles"
)
