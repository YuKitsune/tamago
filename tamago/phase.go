package tamago

import (
	"fmt"
	"github.com/yukitsune/tamago/config"
	"time"
)

type PhaseType int

const (
	WorkPhase       PhaseType = 0
	ShortBreakPhase PhaseType = 1
	LongBreakPhase  PhaseType = 2
	Completed       PhaseType = 3
)

func (t PhaseType) String() string {
	switch t {
	case WorkPhase:
		return "Work"
	case ShortBreakPhase:
		return "Short break"
	case LongBreakPhase:
		return "Long break"
	case Completed:
		return "Complete"
	}

	panic(fmt.Sprintf("unexpected phase type: %d", t))
}

func (t PhaseType) Emoji() string {
	switch t {
	case WorkPhase:
		return "💻"
	case ShortBreakPhase:
		return "☕️"
	case LongBreakPhase:
		return "🍔"
	case Completed:
		return "🎉"
	}

	return ""
}

type Phase struct {
	PhaseType   PhaseType
	PhaseNumber int
	CycleNumber int
}

func (p *Phase) Timeout(cfg config.Config) time.Duration {
	switch p.PhaseType {
	case WorkPhase:
		return cfg.WorkDuration()
	case ShortBreakPhase:
		return cfg.ShortBreakDuration()
	case LongBreakPhase:
		return cfg.LongBreakDuration()
	case Completed:
		return 0
	default:
		panic(fmt.Sprintf("unexpected phase type: %d", p.PhaseType))
	}
}

func InitialPhase() *Phase {
	return &Phase{
		PhaseType:   WorkPhase,
		PhaseNumber: 0,
		CycleNumber: 0,
	}
}

func NextPhase(currentPhase *Phase, cfg config.Config) *Phase {

	nextPhaseNumber := currentPhase.PhaseNumber + 1

	if nextPhaseNumber >= cfg.PhasesPerCycle() {

		nextCycleNumber := currentPhase.CycleNumber + 1
		if nextCycleNumber >= cfg.TotalCycles() {
			return &Phase{PhaseType: Completed, PhaseNumber: nextPhaseNumber, CycleNumber: currentPhase.CycleNumber}
		}

		return &Phase{
			PhaseType:   WorkPhase,
			PhaseNumber: 0,
			CycleNumber: currentPhase.CycleNumber + 1,
		}
	}

	switch currentPhase.PhaseType {

	// If we're currently in a work phase, then the next phase should be a break phase
	case WorkPhase:
		if nextPhaseNumber >= cfg.PhasesPerCycle()-1 {
			return &Phase{PhaseType: LongBreakPhase, PhaseNumber: nextPhaseNumber, CycleNumber: currentPhase.CycleNumber}
		}

		return &Phase{PhaseType: ShortBreakPhase, PhaseNumber: nextPhaseNumber, CycleNumber: currentPhase.CycleNumber}

	// If we're currently in a break phase, then the next phase should be a work phase
	case ShortBreakPhase:
		fallthrough
	case LongBreakPhase:
		return &Phase{PhaseType: WorkPhase, PhaseNumber: nextPhaseNumber, CycleNumber: currentPhase.CycleNumber}

	default:
		panic(fmt.Sprintf("unexpected phase type: %d", currentPhase.PhaseType))
	}

	return nil
}
