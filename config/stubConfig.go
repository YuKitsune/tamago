package config

import "time"

type stubConfig struct {
	dryRun             bool
	workDuration       time.Duration
	shortBreakDuration time.Duration
	longBreakDuration  time.Duration
	phasesPerCycle     int
	totalCycles        int
}

func (c *stubConfig) DryRun() bool {
	return c.dryRun
}

func (c *stubConfig) WorkDuration() time.Duration {
	return c.workDuration
}

func (c *stubConfig) ShortBreakDuration() time.Duration {
	return c.shortBreakDuration
}

func (c *stubConfig) LongBreakDuration() time.Duration {
	return c.longBreakDuration
}

func (c *stubConfig) PhasesPerCycle() int {
	return c.phasesPerCycle
}

func (c *stubConfig) TotalCycles() int {
	return c.totalCycles
}
