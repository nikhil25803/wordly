package ui

import (
	"fmt"
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nikhil25803/wordly/internal/game"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#6AAA64"))
	tileStyle  = lipgloss.NewStyle().Width(5).Align(lipgloss.Center).
			Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#374151"))
	correctStyle = tileStyle.Copy().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#6AAA64"))
	presentStyle = tileStyle.Copy().Underline(true).Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#C9B458"))
	absentStyle = tileStyle.Copy().Faint(true).Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#787C7E"))
	messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	mutedStyle   = lipgloss.NewStyle().Faint(true)
	cardStyle    = lipgloss.NewStyle().Padding(0, 1).Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#374151"))
	statsHeadingStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#C9B458"))
	barStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#6AAA64"))
)

type Model struct {
	game    *game.Game
	input   string
	message string
	width   int
	height  int
}

func NewModel(currentGame *game.Game) Model {
	return Model{game: currentGame}
}

func Run(currentGame *game.Game, output io.Writer) error {
	_, err := tea.NewProgram(NewModel(currentGame), tea.WithOutput(output)).Run()
	return err
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		key := msg.String()
		if key == "ctrl+c" || key == "esc" {
			return m, tea.Quit
		}
		if m.game.Done {
			if key == "enter" || key == "q" {
				return m, tea.Quit
			}
			return m, nil
		}

		switch key {
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
				m.message = ""
			}
		case "enter":
			if len(m.input) != game.WordLength {
				m.message = game.ErrGuessLength.Error()
				return m, nil
			}
			if err := m.game.SubmitGuess(m.input); err != nil {
				m.message = err.Error()
				return m, nil
			}
			m.input = ""
			m.message = ""
		default:
			for _, letter := range strings.ToLower(tea.Key(msg).Text) {
				if letter >= 'a' && letter <= 'z' && len(m.input) < game.WordLength {
					m.input += string(letter)
					m.message = ""
				}
			}
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	header := titleStyle.Render("W O R D L Y") + "\n" +
		mutedStyle.Render(m.game.PuzzleDate+"  •  "+m.game.User.Username)
	sections := []string{
		header,
		m.renderBoard(),
		m.renderStatus(),
	}
	content := lipgloss.NewStyle().Padding(1).Render(strings.Join(sections, "\n\n"))
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func (m Model) renderBoard() string {
	large := m.width >= 40 && m.height >= 36
	rows := make([]string, game.MaxAttempts)
	for row := 0; row < game.MaxAttempts; row++ {
		tiles := make([]string, game.WordLength)
		for column := 0; column < game.WordLength; column++ {
			letter := "·"
			style := tileStyleFor(tileStyle, large)
			if row < len(m.game.Guesses) {
				tile := m.game.Guesses[row].Tiles[column]
				letter = strings.ToUpper(string(tile.Letter))
				switch tile.State {
				case game.Correct:
					style = tileStyleFor(correctStyle, large)
				case game.Present:
					style = tileStyleFor(presentStyle, large)
				default:
					style = tileStyleFor(absentStyle, large)
				}
			} else if row == len(m.game.Guesses) && column < len(m.input) && !m.game.Done {
				letter = strings.ToUpper(string(m.input[column]))
				style = tileStyleFor(tileStyle.Copy().Bold(true), large)
			}
			if column < game.WordLength-1 {
				style = style.MarginRight(2)
			}
			tiles[column] = style.Render(letter)
		}
		rows[row] = lipgloss.JoinHorizontal(lipgloss.Top, tiles...)
	}
	return lipgloss.JoinVertical(lipgloss.Center, rows...)
}

func tileStyleFor(style lipgloss.Style, large bool) lipgloss.Style {
	if large {
		return style.Padding(1, 0)
	}
	return style
}

func (m Model) renderStatus() string {
	legend := mutedStyle.Render("bold=correct  underline=present  dim=absent")
	if m.game.Done {
		result := "Not quite."
		if m.game.Won {
			result = fmt.Sprintf("Solved in %d/%d!", len(m.game.Guesses), game.MaxAttempts)
		}
		return strings.Join([]string{
			titleStyle.Render(result) + "  Answer: " + strings.ToUpper(m.game.Answer()),
			legend,
			"",
			renderStyledStats(m.game.Stats),
			mutedStyle.Render("Enter/q/esc/ctrl+c: quit"),
		}, "\n")
	}
	help := legend + "\n" + mutedStyle.Render("letters: type  •  enter: submit  •  esc/ctrl+c: quit")
	if m.message != "" {
		return messageStyle.Render(m.message) + "\n" + help
	}
	return help
}

func RenderStats(stats game.Stats) string {
	lines := []string{
		fmt.Sprintf("Played %d  Wins %d  Win %d%%  Streak %d  Max %d",
			stats.Played, stats.Wins, stats.WinPercentage, stats.CurrentStreak, stats.MaxStreak),
		"Guess distribution:",
	}
	barWidths := distributionBarWidths(stats)
	for attempt, count := range stats.Distribution {
		lines = append(lines, fmt.Sprintf("%d  %-12s %d", attempt+1, strings.Repeat("█", barWidths[attempt]), count))
	}
	return strings.Join(lines, "\n")
}

func renderStyledStats(stats game.Stats) string {
	cards := []string{
		cardStyle.Render(fmt.Sprintf("Played %d", stats.Played)),
		cardStyle.Render(fmt.Sprintf("Wins %d", stats.Wins)),
		cardStyle.Render(fmt.Sprintf("Win %d%%", stats.WinPercentage)),
		cardStyle.Render(fmt.Sprintf("Streak %d", stats.CurrentStreak)),
		cardStyle.Render(fmt.Sprintf("Max %d", stats.MaxStreak)),
	}
	lines := []string{strings.Join(cards, " "), statsHeadingStyle.Render("Guess distribution")}
	barWidths := distributionBarWidths(stats)
	for attempt, count := range stats.Distribution {
		bar := fmt.Sprintf("%-12s", strings.Repeat("█", barWidths[attempt]))
		if count == 0 {
			lines = append(lines, mutedStyle.Render(fmt.Sprintf("%d  %s %d", attempt+1, bar, count)))
			continue
		}
		lines = append(lines, fmt.Sprintf("%d  %s %d", attempt+1, barStyle.Render(bar), count))
	}
	return strings.Join(lines, "\n")
}

func distributionBarWidths(stats game.Stats) [game.MaxAttempts]int {
	maxCount := 1
	for _, count := range stats.Distribution {
		if count > maxCount {
			maxCount = count
		}
	}

	var widths [game.MaxAttempts]int
	for attempt, count := range stats.Distribution {
		widths[attempt] = count * 12 / maxCount
		if count > 0 && widths[attempt] == 0 {
			widths[attempt] = 1
		}
	}
	return widths
}
