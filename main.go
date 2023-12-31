package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yukitsune/tamago/config"
	"github.com/yukitsune/tamago/tamago"
	"log"
	"strconv"
	"time"
)

// Todo:
//  - Completion hooks
//  - Progress bar for current phase
//  - Pretty colours

func main() {

	var rootCmd = &cobra.Command{
		Use:   "tamago",
		Short: "A lightweight pomodoro-style timer",
		RunE:  run,
	}

	if err := configureFlags(rootCmd); err != nil {
		log.Fatal(err)
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Prints the current version",
		Run: func(cmd *cobra.Command, args []string) {
			println(tamago.Version)
		},
	}

	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func configureFlags(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool(config.DryRunKey, false, "Prints the planned phases")
	if err := viper.BindPFlag(config.DryRunKey, cmd.PersistentFlags().Lookup(config.DryRunKey)); err != nil {
		return err
	}

	cmd.PersistentFlags().DurationP(config.WorkDurationKey, "w", 25*time.Minute, "The duration of a work phase. Valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\"")
	if err := viper.BindPFlag(config.WorkDurationKey, cmd.PersistentFlags().Lookup(config.WorkDurationKey)); err != nil {
		return err
	}

	cmd.PersistentFlags().DurationP(config.ShortBreakDurationKey, "s", 5*time.Minute, "The duration of a short break. Valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\"")
	if err := viper.BindPFlag(config.ShortBreakDurationKey, cmd.PersistentFlags().Lookup(config.ShortBreakDurationKey)); err != nil {
		return err
	}

	cmd.PersistentFlags().DurationP(config.LongBreakDurationKey, "l", 20*time.Minute, "The duration of a long break. Valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\"")
	if err := viper.BindPFlag(config.LongBreakDurationKey, cmd.PersistentFlags().Lookup(config.LongBreakDurationKey)); err != nil {
		return err
	}

	cmd.PersistentFlags().IntP(config.PhasesPerCycleKey, "p", 8, "The total number of phases per cycle. Note that work phases and break phases are two separate phases.")
	if err := viper.BindPFlag(config.PhasesPerCycleKey, cmd.PersistentFlags().Lookup(config.PhasesPerCycleKey)); err != nil {
		return err
	}

	cmd.PersistentFlags().IntP(config.TotalCyclesKey, "c", 1, "The total number of cycles.")
	if err := viper.BindPFlag(config.TotalCyclesKey, cmd.PersistentFlags().Lookup(config.TotalCyclesKey)); err != nil {
		return err
	}

	return nil
}

func run(_ *cobra.Command, _ []string) error {

	cfg := config.NewViperConfig(viper.GetViper())

	phasePlan := tamago.BuildPlan(cfg)
	if cfg.DryRun() {
		return printPlan(phasePlan, cfg)
	}

	m := &model{
		cfg:    cfg,
		plan:   phasePlan,
		timer:  timer.NewWithInterval(phasePlan.CurrentPhase().Timeout(cfg), time.Second),
		keymap: newKeymap(),
		help:   help.New(),
	}

	m.keymap.acknowledge.SetEnabled(false)
	m.keymap.resume.SetEnabled(false)

	_, err := tea.NewProgram(m).Run()
	return err
}

func printPlan(plan *tamago.PhasePlan, cfg config.Config) error {

	list := lipgloss.NewStyle().MarginRight(2)
	var phaseStrings []string

	// To help out with formatting, we need to know the size of the headers
	cycleWidth := len("cycle")
	phaseWidth := len("Short Break") // Longest string
	durationWidth := len("99m59s")

	fitToWidth := func(str string, width int) string {
		return lipgloss.NewStyle().PaddingRight(width - len(str)).Render(str)
	}

	// Header
	phaseStrings = append(phaseStrings, lipgloss.NewStyle().
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		MarginRight(2).
		Render(fitToWidth("Cycle", cycleWidth), fitToWidth("Phase", phaseWidth), fitToWidth("Duration", durationWidth)))

	// Phases
	for _, phase := range plan.Phases {
		cycleString := fitToWidth(strconv.Itoa(phase.CycleNumber), cycleWidth)
		phaseString := fitToWidth(phase.PhaseType.String(), phaseWidth)
		durationString := fitToWidth(phase.Timeout(cfg).String(), durationWidth)

		phaseStrings = append(phaseStrings, lipgloss.NewStyle().Render(cycleString, phaseString, durationString))
	}

	listString := list.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			phaseStrings...,
		),
	)

	_, err := fmt.Printf(listString)
	return err
}

type model struct {
	cfg                     config.Config
	plan                    *tamago.PhasePlan
	timer                   timer.Model
	keymap                  keymap
	help                    help.Model
	acknowledgementRequired bool
	acknowledgementLightOn  bool
	acknowledgementTimer    timer.Model
	showProgress            bool
	quitting                bool
}

func (m *model) AdvancePhase(requireAcknowledgement bool) (tea.Cmd, *tamago.Phase) {
	nextPhase := m.plan.AdvancePhase()
	if nextPhase.PhaseType == tamago.Completed {
		m.quitting = true
	}

	timeout := nextPhase.Timeout(m.cfg)
	m.timer.Timeout = timeout
	var cmd tea.Cmd
	if requireAcknowledgement {
		cmd = m.SeekAcknowledgement(timeout)
	}

	return cmd, nextPhase
}

func (m *model) SeekAcknowledgement(timeout time.Duration) tea.Cmd {
	m.acknowledgementTimer = timer.NewWithInterval(timeout, time.Second)
	m.acknowledgementTimer.Start()
	m.acknowledgementRequired = true
	m.acknowledgementLightOn = true
	m.keymap.acknowledge.SetEnabled(true)
	return m.acknowledgementTimer.Init()
}

func (m *model) FlashAcknowledgementLight() {
	m.acknowledgementLightOn = !m.acknowledgementLightOn
}

func (m *model) Acknowledge() {
	m.acknowledgementTimer.Stop()
	m.acknowledgementRequired = false
	m.acknowledgementLightOn = false
	m.keymap.acknowledge.SetEnabled(false)
}

func (m *model) Init() tea.Cmd {
	return m.timer.Init()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:

		var cmd tea.Cmd
		if msg.ID == m.timer.ID() {
			m.timer, cmd = m.timer.Update(msg)
		}

		if m.acknowledgementRequired && msg.ID == m.acknowledgementTimer.ID() {
			m.acknowledgementTimer, cmd = m.acknowledgementTimer.Update(msg)
			m.FlashAcknowledgementLight()
		}
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		if msg.ID == m.timer.ID() {
			m.timer, cmd = m.timer.Update(msg)
			m.keymap.pause.SetEnabled(m.timer.Running())
			m.keymap.resume.SetEnabled(!m.timer.Running())
		}

		if msg.ID == m.acknowledgementTimer.ID() {
			m.acknowledgementTimer, cmd = m.acknowledgementTimer.Update(msg)
		}
		return m, cmd

	case timer.TimeoutMsg:

		if msg.ID == m.timer.ID() {

			if m.plan.CurrentPhase().PhaseType == tamago.Completed {
				return m, tea.Quit
			}

			cmd, newPhase := m.AdvancePhase(true)
			if newPhase.PhaseType == tamago.Completed {
				return m, tea.Quit
			}

			return m, cmd
		}

		if msg.ID == m.acknowledgementTimer.ID() {
			m.Acknowledge()
			return m, nil
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.acknowledge):
			m.Acknowledge()
			return m, nil
		case key.Matches(msg, m.keymap.pause, m.keymap.resume):
			return m, m.timer.Toggle()
		case key.Matches(msg, m.keymap.progress):
			m.showProgress = !m.showProgress
			return m, nil
		case key.Matches(msg, m.keymap.reset):
			m.timer.Timeout = m.plan.CurrentPhase().Timeout(m.cfg)
		case key.Matches(msg, m.keymap.skip):

			// If the user is intentionally skipping this phase, then they don't need to acknowledge the phase change
			cmd, nextPhase := m.AdvancePhase(false)
			if nextPhase.PhaseType == tamago.Completed {
				m.quitting = true
				return m, tea.Quit
			}
			return m, cmd
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *model) View() string {

	withAcknowledgementLightOn := func(str string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(str)
	}

	// Current phase and time remaining
	currentPhase := m.plan.CurrentPhase()
	phaseType := currentPhase.PhaseType
	s := fmt.Sprintf("%s %s: %s", phaseType.Emoji(), phaseType.String(), m.timer.Timeout)
	if m.acknowledgementRequired && m.acknowledgementLightOn {
		s = withAcknowledgementLightOn(s)
	}

	// Paused indicator
	if !m.timer.Running() {
		s += " (Paused)"
	}

	if m.timer.Timedout() {
		// Todo: Summary with supportive message <3
		s = "All done!"
	}

	// Progress section
	if m.showProgress {
		s += "\n"
		s += m.progressView()
	}

	// Help section
	if !m.quitting {
		s += "\n"
		s += m.helpView()
	}

	return s
}

func (m *model) progressView() string {
	list := lipgloss.NewStyle().MarginRight(2)

	checkMark := lipgloss.NewStyle().SetString("✓").
		Foreground(lipgloss.Color("46")).
		PaddingRight(1).
		String()

	dot := lipgloss.NewStyle().SetString("·").
		PaddingRight(1).
		String()

	var phaseStrings []string

	// Header
	phaseStrings = append(phaseStrings, lipgloss.NewStyle().
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		MarginRight(2).
		Render("Phases"))

	// Phases
	for i, phase := range m.plan.Phases {
		phaseType := phase.PhaseType.String()
		if m.plan.IsCompleted(phase) {
			phaseStrings = append(phaseStrings, checkMark+lipgloss.NewStyle().
				Strikethrough(true).
				Foreground(lipgloss.Color("8")).
				Render(phaseType))
		} else if i == 0 || m.plan.IsCompleted(m.plan.Phases[i-1]) {
			phaseStrings = append(phaseStrings, dot+phaseType)
		} else {
			phaseStrings = append(phaseStrings, lipgloss.NewStyle().
				PaddingLeft(2).
				Render(phaseType))
		}
	}

	return list.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			phaseStrings...,
		),
	)
}

func (m *model) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.acknowledge,
		m.keymap.pause,
		m.keymap.resume,
		m.keymap.progress,
		m.keymap.reset,
		m.keymap.skip,
		m.keymap.quit,
	})
}

type keymap struct {
	acknowledge key.Binding
	pause       key.Binding
	resume      key.Binding
	progress    key.Binding
	reset       key.Binding
	skip        key.Binding
	quit        key.Binding
}

func newKeymap() keymap {
	return keymap{
		acknowledge: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "acknowledge"),
		),
		pause: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause"),
		),
		resume: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "resume"),
		),
		progress: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "progress"),
		),
		reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset"),
		),
		skip: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "skip"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}
