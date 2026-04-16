package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Brand color
	purple = lipgloss.Color("99")
	green  = lipgloss.Color("10")
	blue   = lipgloss.Color("12")
	cyan   = lipgloss.Color("14")
	yellow = lipgloss.Color("11")
	gray   = lipgloss.Color("8")
	white  = lipgloss.Color("15")
	red    = lipgloss.Color("9")
	dim    = lipgloss.Color("241")

	// Header bar
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white).
			Background(purple).
			Padding(0, 1)

	// Current change panel (green border)
	currentBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(green).
				Padding(0, 1)

	currentTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(green)

	// Incoming change panel (blue border)
	incomingBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(blue).
				Padding(0, 1)

	incomingTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(blue)

	// Line number style
	lineNumStyle = lipgloss.NewStyle().
			Foreground(dim).
			Width(4).
			Align(lipgloss.Right)

	// Menu
	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(yellow)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(gray)

	// Help bar
	helpStyle = lipgloss.NewStyle().
			Foreground(gray)

	// Summary styles
	successStyle = lipgloss.NewStyle().
			Foreground(green)

	warningStyle = lipgloss.NewStyle().
			Foreground(yellow)

	skipStyle = lipgloss.NewStyle().
			Foreground(gray)

	errorStyle = lipgloss.NewStyle().
			Foreground(red)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cyan).
			MarginBottom(1)
)
