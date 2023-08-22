package config

import (
	"github.com/spf13/viper"
	"time"
)

type viperConfig struct {
	viper *viper.Viper
}

func NewViperConfig(viper *viper.Viper) Config {
	return &viperConfig{viper}
}

func (c *viperConfig) DryRun() bool {
	return c.viper.GetBool(DryRunKey)
}

func (c *viperConfig) WorkDuration() time.Duration {
	return c.viper.GetDuration(WorkDurationKey)
}

func (c *viperConfig) ShortBreakDuration() time.Duration {
	return c.viper.GetDuration(ShortBreakDurationKey)
}

func (c *viperConfig) LongBreakDuration() time.Duration {
	return c.viper.GetDuration(LongBreakDurationKey)
}

func (c *viperConfig) PhasesPerCycle() int {
	return c.viper.GetInt(PhasesPerCycleKey)
}

func (c *viperConfig) TotalCycles() int {
	return c.viper.GetInt(TotalCyclesKey)
}
