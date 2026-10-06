package ui

import "charm.land/lipgloss/v2"

var (
	backgroundColor = lipgloss.Color("#0B1118")
	foregroundColor = lipgloss.Color("#F3F4F6")
	greenColor      = lipgloss.Color("#5BCB76")
	yellowColor     = lipgloss.Color("#D7B93E")
	absentColor     = lipgloss.Color("#4B5563")
	emptyColor      = lipgloss.Color("#202833")
	mutedColor      = lipgloss.Color("#94A3B8")
	dividerColor    = lipgloss.Color("#334155")
	errorColor      = lipgloss.Color("#F87171")

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(greenColor)
	tileStyle = lipgloss.NewStyle().
			Width(5).
			Align(lipgloss.Center).
			Bold(true).
			Foreground(foregroundColor).
			Background(emptyColor)
	correctTileStyle = tileStyle.Copy().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(greenColor)
	presentTileStyle = tileStyle.Copy().
				Underline(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(yellowColor)
	absentTileStyle = tileStyle.Copy().
			Bold(false).
			Faint(true).
			Foreground(lipgloss.Color("#E5E7EB")).
			Background(absentColor)
	emptyTileStyle = tileStyle.Copy().
			Bold(false).
			Faint(true).
			Foreground(mutedColor).
			Background(emptyColor)
	keyboardKeyStyle = lipgloss.NewStyle().
				Width(3).
				Align(lipgloss.Center).
				Foreground(foregroundColor).
				Background(emptyColor)
	selectedMenuItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(greenColor)
	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
	dividerStyle = lipgloss.NewStyle().
			Foreground(dividerColor)
	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(foregroundColor)
	statStyle = lipgloss.NewStyle().
			Foreground(foregroundColor)
	messageStyle = lipgloss.NewStyle().
			Foreground(errorColor)
)
