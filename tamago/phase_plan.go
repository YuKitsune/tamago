package tamago

import "github.com/yukitsune/tamago/config"

type PhasePlanEntry struct {
	Phase     *Phase
	Completed bool
}

type PhasePlan struct {
	Phases       []PhasePlanEntry
	currentIndex int
}

func BuildPlan(cfg config.Config) *PhasePlan {

	phases := []PhasePlanEntry{
		{
			Phase:     InitialPhase(),
			Completed: false,
		},
	}

	i := 0
	for {
		previousPhase := phases[len(phases)-1]
		nextPhase := NextPhase(previousPhase.Phase, cfg)
		phases = append(phases, PhasePlanEntry{
			Phase:     nextPhase,
			Completed: false,
		})

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
	return p.Phases[p.currentIndex].Phase
}

func (p *PhasePlan) AdvancePhase() *Phase {
	p.Phases[p.currentIndex].Completed = true
	p.currentIndex++
	return p.CurrentPhase()
}
