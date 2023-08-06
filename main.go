package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yukitsune/tamago/config"
	"github.com/yukitsune/tamago/tamago"
	"log"
	"time"
)

// Todo:
//  - List of completed and upcoming phases (toggleable)
//  - Dry-run: List phases, cycles, and local start/end times
//  - Acknowledgement: Seek user acknowledgment when phase changes
//  - Flash when phase changes
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
	phase := tamago.InitialPhase()

	m := &model{
		cfg:          cfg,
		currentPhase: phase,
		timer:        timer.NewWithInterval(phase.Timeout(cfg), time.Second),
		keymap:       newKeymap(),
		help:         help.New(),
	}

	m.keymap.resume.SetEnabled(false)

	_, err := tea.NewProgram(m).Run()
	return err
}

type model struct {
	cfg          config.Config
	currentPhase *tamago.Phase
	timer        timer.Model
	keymap       keymap
	help         help.Model
	quitting     bool
}

func (m *model) AdvancePhase() *tamago.Phase {
	m.currentPhase = tamago.NextPhase(m.currentPhase, m.cfg)
	if m.currentPhase.PhaseType == tamago.Completed {
		m.quitting = true
	}
	m.timer.Timeout = m.currentPhase.Timeout(m.cfg)
	return m.currentPhase
}

func (m *model) Init() tea.Cmd {
	return m.timer.Init()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		m.keymap.pause.SetEnabled(m.timer.Running())
		m.keymap.resume.SetEnabled(!m.timer.Running())
		return m, cmd

	case timer.TimeoutMsg:
		if m.currentPhase.PhaseType == tamago.Completed {
			return m, tea.Quit
		}

		newPhase := m.AdvancePhase()
		if newPhase.PhaseType == tamago.Completed {
			return m, tea.Quit
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.pause, m.keymap.resume):
			return m, m.timer.Toggle()
		case key.Matches(msg, m.keymap.reset):
			m.timer.Timeout = m.currentPhase.Timeout(m.cfg)
		case key.Matches(msg, m.keymap.skip):
			nextPhase := m.AdvancePhase()
			if nextPhase.PhaseType == tamago.Completed {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *model) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.pause,
		m.keymap.resume,
		m.keymap.reset,
		m.keymap.skip,
		m.keymap.quit,
	})
}

func (m *model) View() string {

	s := fmt.Sprintf("%s %s: %s", emojiFor(m.currentPhase.PhaseType), m.currentPhase.PhaseType.String(), m.timer.Timeout)

	if !m.timer.Running() {
		s += " (Paused)"
	}

	if m.timer.Timedout() {
		// Todo: Summary with supportive message <3
		s = "All done!"
	} else {
		s += "\n"
	}

	if !m.quitting {
		s += m.helpView()
	}

	return s
}

func emojiFor(phaseType tamago.PhaseType) string {
	switch phaseType {
	case tamago.WorkPhase:
		return "💻"
	case tamago.ShortBreakPhase:
		return "☕️"
	case tamago.LongBreakPhase:
		return "🍔"
	case tamago.Completed:
		return "🎉"
	}

	return ""
}

type keymap struct {
	pause  key.Binding
	resume key.Binding
	reset  key.Binding
	skip   key.Binding
	quit   key.Binding
}

func newKeymap() keymap {
	return keymap{
		pause: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause"),
		),
		resume: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "resume"),
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
