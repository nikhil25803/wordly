package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nikhil25803/wordly/internal/game"
)

const legendText = "bold=correct  underline=present  dim=absent"

func (m Model) View() tea.View {
	width, height := m.viewportSize()
	innerWidth := max(1, width-2)
	innerHeight := max(1, height-2)
	canvasWidth := min(110, innerWidth)

	var content string
	switch m.screen {
	case screenGame:
		content = m.renderGame(canvasWidth, innerHeight, width)
	case screenResult:
		content = m.renderResult(canvasWidth, innerHeight)
	case screenStats:
		content = m.renderStatsScreen(canvasWidth, innerHeight)
	default:
		content = m.renderHome(canvasWidth, innerHeight)
	}

	content = lipgloss.PlaceHorizontal(innerWidth, lipgloss.Center, content)
	content = "\n" + lipgloss.NewStyle().PaddingLeft(1).Render(content)
	view := tea.NewView(content)
	view.AltScreen = true
	view.BackgroundColor = backgroundColor
	view.ForegroundColor = foregroundColor
	view.WindowTitle = "Wordly"
	return view
}

func (m Model) viewportSize() (int, int) {
	width, height := m.width, m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 30
	}
	return width, height
}

func (m Model) renderHome(width, height int) string {
	lines := []string{
		headerStyle.Render("W O R D L Y"),
		mutedStyle.Render("A minimal word game for your terminal"),
		"",
	}
	for index, item := range homeItems {
		prefix := "  "
		style := statStyle
		if index == m.menuIndex {
			prefix = "› "
			style = selectedMenuItemStyle
		}
		lines = append(lines, style.Render(prefix+item))
	}
	if m.loading {
		lines = append(lines, "", mutedStyle.Render("Loading…"))
	} else if m.message != "" {
		lines = append(lines, "", messageStyle.Render(m.message))
	}
	lines = append(lines, "", mutedStyle.Render("↑/↓ or j/k navigate  •  enter select  •  q quit"))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, strings.Join(lines, "\n"))
}

func (m Model) renderGame(width, height, viewportWidth int) string {
	header := m.renderHeader(width, viewportWidth)
	body := m.renderGameBody(width, height, viewportWidth)
	top := header + "\n\n" + body
	return anchorFooter(top, m.renderGameFooter(width), height)
}

func (m Model) renderHeader(width, viewportWidth int) string {
	date := ""
	if m.game != nil {
		date = m.game.PuzzleDate
	}
	left := headerStyle.Render("W O R D L Y")
	right := ""
	switch {
	case viewportWidth >= 120:
		left += mutedStyle.Render("  │  " + date)
		right = statStyle.Render(fmt.Sprintf("Streak %d  │  Wins %d  │  Win %d%%",
			m.stats.CurrentStreak, m.stats.Wins, m.stats.WinPercentage))
	case viewportWidth >= 80:
		left += mutedStyle.Render("  │  " + date)
	}
	return joinEdges(width, left, right) + "\n" + renderDivider(width)
}

func (m Model) renderGameBody(width, height, viewportWidth int) string {
	showKeyboard := height >= 16
	showLegend := height >= 18
	showSummary := height >= 20

	if viewportWidth >= 120 {
		return m.renderLargeGameBody(width, showKeyboard, showLegend, height >= 20)
	}

	parts := []string{m.renderBoard(width)}
	if showKeyboard {
		parts = append(parts, m.renderKeyboard(width, viewportWidth >= 80))
	}
	if showSummary {
		summary := fmt.Sprintf("Streak %d  │  Wins %d", m.stats.CurrentStreak, m.stats.Wins)
		if viewportWidth >= 80 {
			summary += fmt.Sprintf("  │  Win %d%%", m.stats.WinPercentage)
		}
		parts = append(parts, statStyle.Render(summary))
	}
	if showLegend {
		parts = append(parts, mutedStyle.Render(legendText))
	}
	return centerBlocks(width, parts)
}

func (m Model) renderLargeGameBody(width int, showKeyboard, showLegend, showDistribution bool) string {
	const leftWidth = 35
	const rightWidth = 45
	left := lipgloss.PlaceHorizontal(leftWidth, lipgloss.Center, m.renderBoard(leftWidth))

	var right []string
	if showKeyboard {
		right = append(right, statStyle.Render("Keyboard"), m.renderKeyboard(rightWidth, true))
	}
	if showDistribution {
		if len(right) > 0 {
			right = append(right, "", renderDivider(rightWidth))
		}
		right = append(right, statStyle.Render("Guess distribution"), renderDistribution(m.stats, rightWidth))
	}
	if showLegend {
		right = append(right, "", mutedStyle.Render(legendText))
	}
	rightBlock := lipgloss.PlaceHorizontal(rightWidth, lipgloss.Left, strings.Join(right, "\n"))
	separatorHeight := max(lipgloss.Height(left), lipgloss.Height(rightBlock))
	separator := dividerStyle.Render(strings.TrimSuffix(strings.Repeat("│\n", separatorHeight), "\n"))
	playArea := lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", separator, "   ", rightBlock)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, playArea)
}

func (m Model) renderBoard(width int) string {
	rows := make([]string, game.MaxAttempts)
	for row := 0; row < game.MaxAttempts; row++ {
		tiles := make([]string, game.WordLength)
		for column := 0; column < game.WordLength; column++ {
			letter := "·"
			style := emptyTileStyle
			if row < len(m.game.Guesses) {
				tile := m.game.Guesses[row].Tiles[column]
				letter = strings.ToUpper(string(tile.Letter))
				style = tileStyleForState(tile.State)
			} else if row == len(m.game.Guesses) && column < len(m.input) && !m.game.Done {
				letter = strings.ToUpper(string(m.input[column]))
				style = tileStyle
			}
			tiles[column] = style.Render(letter)
		}
		rows[row] = strings.Join(tiles, " ")
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, strings.Join(rows, "\n"))
}

func tileStyleForState(state game.LetterState) lipgloss.Style {
	switch state {
	case game.Correct:
		return correctTileStyle
	case game.Present:
		return presentTileStyle
	default:
		return absentTileStyle
	}
}

func (m Model) renderKeyboard(width int, wide bool) string {
	states := keyboardStates(m.game.Guesses)
	keyWidth := 2
	if wide {
		keyWidth = 3
	}
	rows := [][]string{
		keyboardRow("qwertyuiop", states, keyWidth),
		keyboardRow("asdfghjkl", states, keyWidth),
		append([]string{renderKeyboardKey("⌫", game.Absent, false, keyWidth+1)},
			append(keyboardRow("zxcvbnm", states, keyWidth),
				renderKeyboardKey("↵", game.Absent, false, keyWidth+1))...),
	}
	rendered := make([]string, len(rows))
	for index, row := range rows {
		rendered[index] = lipgloss.PlaceHorizontal(width, lipgloss.Center, strings.Join(row, " "))
	}
	return strings.Join(rendered, "\n")
}

func keyboardRow(letters string, states map[byte]game.LetterState, width int) []string {
	keys := make([]string, 0, len(letters))
	for index := range letters {
		letter := letters[index]
		state, used := states[letter]
		keys = append(keys, renderKeyboardKey(strings.ToUpper(string(letter)), state, used, width))
	}
	return keys
}

func renderKeyboardKey(label string, state game.LetterState, used bool, width int) string {
	style := keyboardKeyStyle.Copy().Width(width)
	if used {
		style = tileStyleForState(state).Copy().Width(width)
	}
	return style.Render(label)
}

func keyboardStates(guesses []game.EvaluatedGuess) map[byte]game.LetterState {
	states := make(map[byte]game.LetterState)
	for _, guess := range guesses {
		for _, tile := range guess.Tiles {
			if tile.Letter < 'a' || tile.Letter > 'z' {
				continue
			}
			state, used := states[tile.Letter]
			if !used || tile.State > state {
				states[tile.Letter] = tile.State
			}
		}
	}
	return states
}

func (m Model) renderGameFooter(width int) string {
	status := mutedStyle.Render("esc menu  •  ctrl+c quit")
	if m.message != "" {
		status = messageStyle.Render(m.message)
	}
	prompt := promptStyle.Render("› ")
	if m.input == "" {
		prompt += mutedStyle.Render("Enter your guess…")
	} else {
		prompt += promptStyle.Render(strings.ToUpper(m.input))
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Left, status) + "\n" +
		lipgloss.PlaceHorizontal(width, lipgloss.Left, prompt)
}

func (m Model) renderResult(width, height int) string {
	won := m.game != nil && m.game.Won
	title := "Not quite"
	if won {
		title = "You got it!"
	}

	answer := ""
	score := ""
	if m.game != nil {
		answer = strings.ToUpper(m.game.Answer())
		score = resultScore(m.game)
	}

	lines := []string{
		headerStyle.Render("W O R D L Y"),
		"",
		statStyle.Render(title),
		"",
		m.renderBoard(width),
		"",
		statStyle.Render("Answer "+answer) + mutedStyle.Render("  │  "+score),
		"",
	}
	for index, item := range resultItems {
		prefix := "  "
		style := statStyle
		if index == m.resultIndex {
			prefix = "› "
			style = selectedMenuItemStyle
		}
		lines = append(lines, style.Render(prefix+item))
	}
	if m.message != "" {
		lines = append(lines, "", selectedMenuItemStyle.Render(m.message))
	}
	lines = append(lines, "", mutedStyle.Render("↑/↓ or j/k navigate  •  s share  •  esc menu"))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, strings.Join(lines, "\n"))
}

func resultScore(current *game.Game) string {
	if current == nil || !current.Done {
		return ""
	}
	if !current.Won {
		return fmt.Sprintf("X/%d", game.MaxAttempts)
	}
	return fmt.Sprintf("%d/%d", len(current.Guesses), game.MaxAttempts)
}

func formatShareResult(current *game.Game) string {
	if current == nil || !current.Done {
		return ""
	}

	lines := []string{fmt.Sprintf("WORDLY %s %s", current.PuzzleDate, resultScore(current)), ""}
	for _, guess := range current.Guesses {
		var row strings.Builder
		for _, tile := range guess.Tiles {
			switch tile.State {
			case game.Correct:
				row.WriteString("🟩")
			case game.Present:
				row.WriteString("🟨")
			default:
				row.WriteString("⬛")
			}
		}
		lines = append(lines, row.String())
	}
	lines = append(lines, "", "I played today's Wordly — can you solve it too?", "https://github.com/nikhil25803/wordly")
	return strings.Join(lines, "\n")
}

func (m Model) renderStatsScreen(width, height int) string {
	header := joinEdges(width, headerStyle.Render("WORDLY")+mutedStyle.Render("  │  Statistics"), mutedStyle.Render("← Back"))
	metrics := fmt.Sprintf("Played %d  │  Wins %d  │  Win Rate %d%%  │  Current Streak %d",
		m.stats.Played, m.stats.Wins, m.stats.WinPercentage, m.stats.CurrentStreak)
	if width < 78 {
		metrics = fmt.Sprintf("Played %-4d  Wins %d\nWin Rate %-3d%%  Current Streak %d",
			m.stats.Played, m.stats.Wins, m.stats.WinPercentage, m.stats.CurrentStreak)
	}
	content := strings.Join([]string{
		header,
		renderDivider(width),
		"",
		statStyle.Render(metrics),
		"",
		renderDivider(width),
		"",
		statStyle.Render("Guess Distribution"),
		"",
		renderDistribution(m.stats, width),
	}, "\n")
	return anchorFooter(content, mutedStyle.Render("esc, q, or ← to go back"), height)
}

func renderDistribution(stats game.Stats, width int) string {
	maxCount := 1
	for _, count := range stats.Distribution {
		maxCount = max(maxCount, count)
	}
	maxBarWidth := min(18, max(1, width-8))
	lines := make([]string, game.MaxAttempts)
	for attempt, count := range stats.Distribution {
		barWidth := count * maxBarWidth / maxCount
		if count > 0 && barWidth == 0 {
			barWidth = 1
		}
		bar := fmt.Sprintf("%-*s", maxBarWidth, strings.Repeat("█", barWidth))
		if count == 0 {
			bar = mutedStyle.Render(bar)
		} else {
			bar = lipgloss.NewStyle().Foreground(greenColor).Render(bar)
		}
		lines[attempt] = fmt.Sprintf("%d  %s %d", attempt+1, bar, count)
	}
	return strings.Join(lines, "\n")
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

func distributionBarWidths(stats game.Stats) [game.MaxAttempts]int {
	maxCount := 1
	for _, count := range stats.Distribution {
		maxCount = max(maxCount, count)
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

func renderDivider(width int) string {
	return dividerStyle.Render(strings.Repeat("─", max(0, width)))
}

func centerBlocks(width int, blocks []string) string {
	rendered := make([]string, 0, len(blocks))
	for _, block := range blocks {
		rendered = append(rendered, lipgloss.PlaceHorizontal(width, lipgloss.Center, block))
	}
	return strings.Join(rendered, "\n\n")
}

func joinEdges(width int, left, right string) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return left
	}
	return left + strings.Repeat(" ", gap) + right
}

func anchorFooter(top, footer string, height int) string {
	newlines := height - lipgloss.Height(top) - lipgloss.Height(footer) + 1
	if newlines < 1 {
		newlines = 1
	}
	return top + strings.Repeat("\n", newlines) + footer
}
