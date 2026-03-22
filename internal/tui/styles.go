package tui

import "github.com/charmbracelet/lipgloss"

var (
	styleAdded    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // green
	styleRemoved  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))  // red
	styleModified = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))  // yellow
	styleHeader   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")) // bright blue
	styleDim      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // dim gray
	styleSelected = lipgloss.NewStyle().Background(lipgloss.Color("8")).Foreground(lipgloss.Color("15"))
	styleSecurity = lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // orange
	styleHelp     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)
