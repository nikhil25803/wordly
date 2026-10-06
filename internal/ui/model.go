package ui

import (
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/nikhil25803/wordly/internal/game"
)

type screen uint8

const (
	screenHome screen = iota
	screenGame
	screenResult
	screenStats
)

var (
	homeItems   = []string{"Play Daily", "Statistics", "Quit"}
	resultItems = []string{"Share Result", "View Stats", "Back to Menu", "Quit"}
)

type gameStartedMsg struct {
	game *game.Game
	err  error
}

type statsLoadedMsg struct {
	stats game.Stats
	err   error
}

type Model struct {
	screen         screen
	previousScreen screen
	game           *game.Game
	stats          game.Stats
	input          string
	message        string
	width          int
	height         int
	menuIndex      int
	resultIndex    int
	loading        bool
}

func NewModel() Model {
	return Model{screen: screenHome}
}

func Run(output io.Writer) error {
	_, err := tea.NewProgram(NewModel(), tea.WithOutput(output)).Run()
	return err
}

func (m Model) Init() tea.Cmd { return nil }

func startGameCmd() tea.Msg {
	currentGame, err := game.StartGame()
	return gameStartedMsg{game: currentGame, err: err}
}

func loadStatsCmd() tea.Msg {
	stats, err := game.GetCurrentUserStats()
	return statsLoadedMsg{stats: stats, err: err}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case gameStartedMsg:
		m.loading = false
		if msg.err != nil {
			m.message = msg.err.Error()
			return m, nil
		}
		m.game = msg.game
		m.stats = msg.game.Stats
		m.input = ""
		m.message = ""
		if msg.game.Done {
			m.screen = screenResult
		} else {
			m.screen = screenGame
		}
		return m, nil
	case statsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.message = msg.err.Error()
			return m, nil
		}
		m.stats = msg.stats
		m.message = ""
		m.screen = screenStats
		return m, nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.loading {
			return m, nil
		}
		switch m.screen {
		case screenHome:
			return m.updateHome(msg)
		case screenGame:
			return m.updateGame(msg)
		case screenResult:
			return m.updateResult(msg)
		case screenStats:
			return m.updateStats(msg)
		}
	}
	return m, nil
}

func (m Model) updateHome(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.menuIndex = previousIndex(m.menuIndex, len(homeItems))
	case "down", "j":
		m.menuIndex = nextIndex(m.menuIndex, len(homeItems))
	case "q", "esc":
		return m, tea.Quit
	case "enter":
		m.message = ""
		switch m.menuIndex {
		case 0:
			m.loading = true
			return m, startGameCmd
		case 1:
			m.loading = true
			m.previousScreen = screenHome
			return m, loadStatsCmd
		default:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) updateGame(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "esc" {
		m.screen = screenHome
		m.input = ""
		m.message = ""
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
		m.stats = m.game.Stats
		if m.game.Done {
			m.resultIndex = 0
			m.screen = screenResult
		}
	default:
		for _, letter := range strings.ToLower(tea.Key(msg).Text) {
			if letter >= 'a' && letter <= 'z' && len(m.input) < game.WordLength {
				m.input += string(letter)
				m.message = ""
			}
		}
	}
	return m, nil
}

func (m Model) updateResult(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.resultIndex = previousIndex(m.resultIndex, len(resultItems))
	case "down", "j":
		m.resultIndex = nextIndex(m.resultIndex, len(resultItems))
	case "esc":
		m.screen = screenHome
		m.message = ""
	case "s":
		return m.shareResult()
	case "enter":
		switch m.resultIndex {
		case 0:
			return m.shareResult()
		case 1:
			m.stats = m.game.Stats
			m.previousScreen = screenResult
			m.screen = screenStats
			m.message = ""
		case 2:
			m.screen = screenHome
			m.message = ""
		default:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) shareResult() (tea.Model, tea.Cmd) {
	m.message = "Copied result to clipboard"
	return m, tea.SetClipboard(formatShareResult(m.game))
}

func (m Model) updateStats(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "left", "backspace", "q":
		if m.previousScreen == screenResult {
			m.screen = screenResult
		} else {
			m.screen = screenHome
		}
	}
	return m, nil
}

func nextIndex(index, length int) int {
	return (index + 1) % length
}

func previousIndex(index, length int) int {
	return (index - 1 + length) % length
}
