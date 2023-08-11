package tamago

import "github.com/yukitsune/tamago/config"

type PhasePlan struct {
	Phases       []Phase
	currentIndex int
}

func BuildPlan(cfg config.Config) *PhasePlan {

	phases := []Phase{
		*InitialPhase(),
	}

	i := 0
	for {
		// Todo: Migrate NextPhase into this method
		previousPhase := phases[len(phases)-1]
		nextPhase := NextPhase(&previousPhase, cfg)
		phases = append(phases, *nextPhase)

		if nextPhase.PhaseType == Completed {
			break
		}

		// Hard limit and panic
		if i > 1000 {
			panic("maximum iterations exceeded when building phase plan")
		}

		i++
	}

	plan := &PhasePlan{phases, 0}
	return plan
}

func (p *PhasePlan) CurrentPhase() *Phase {
	return &p.Phases[p.currentIndex]
}

func (p *PhasePlan) AdvancePhase() *Phase {
	p.currentIndex++
	return p.CurrentPhase()
}

func (p *PhasePlan) IsCompleted(phase Phase) bool {
	for i, p1 := range p.Phases {
		if p1.PhaseNumber == phase.PhaseNumber && p1.CycleNumber == phase.CycleNumber {
			return i < p.currentIndex
		}
	}

	return false
}
