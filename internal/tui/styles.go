package tui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	DimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	AccentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
)
