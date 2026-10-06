package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nikhil25803/wordly/internal/db"
	"github.com/nikhil25803/wordly/internal/game"
)

func setupTestModel(t *testing.T) Model {
	t.Helper()
	t.Setenv("WORDLY_DB", filepath.Join(t.TempDir(), "wordly.db"))
	if err := db.ConnectDatabase(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.DB.Close() })
	return NewModel()
}

func startTestGame(t *testing.T) *game.Game {
	t.Helper()
	current, err := game.StartGame()
	if err != nil {
		t.Fatal(err)
	}
	return current
}

func pressText(t *testing.T, model Model, text string) Model {
	t.Helper()
	for _, letter := range text {
		updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: string(letter), Code: letter}))
		model = updated.(Model)
	}
	return model
}

func pressKey(t *testing.T, model Model, code rune) (Model, tea.Cmd) {
	t.Helper()
	updated, cmd := model.Update(tea.KeyPressMsg(tea.Key{Code: code}))
	return updated.(Model), cmd
}

func TestHomeNavigationAndAsyncLoading(t *testing.T) {
	model := setupTestModel(t)
	if model.screen != screenHome {
		t.Fatalf("initial screen = %v, want home", model.screen)
	}

	model, _ = pressKey(t, model, 'j')
	if model.menuIndex != 1 {
		t.Fatalf("j selected item %d, want 1", model.menuIndex)
	}
	model, _ = pressKey(t, model, 'k')
	if model.menuIndex != 0 {
		t.Fatalf("k selected item %d, want 0", model.menuIndex)
	}

	var cmd tea.Cmd
	model, cmd = pressKey(t, model, tea.KeyEnter)
	if cmd == nil || !model.loading {
		t.Fatal("Play Daily did not start asynchronous loading")
	}
	updated, _ := model.Update(cmd())
	model = updated.(Model)
	if model.screen != screenGame || model.game == nil || model.loading {
		t.Fatalf("loaded daily game = screen:%v game:%v loading:%v", model.screen, model.game != nil, model.loading)
	}

	model.screen = screenHome
	model.menuIndex = 1
	model, cmd = pressKey(t, model, tea.KeyEnter)
	if cmd == nil {
		t.Fatal("Statistics did not start asynchronous loading")
	}
	updated, _ = model.Update(cmd())
	model = updated.(Model)
	if model.screen != screenStats || model.previousScreen != screenHome {
		t.Fatalf("stats loaded into screen %v from %v", model.screen, model.previousScreen)
	}
}

func TestGameInputValidationAndEscape(t *testing.T) {
	model := setupTestModel(t)
	model.screen = screenGame
	model.game = startTestGame(t)

	model = pressText(t, model, "s")
	if model.input != "s" {
		t.Fatal("s did not remain a playable letter")
	}
	model, _ = pressKey(t, model, tea.KeyBackspace)
	model = pressText(t, model, "qwerty")
	if model.input != "qwert" {
		t.Fatalf("input = %q, want qwert", model.input)
	}
	if len(model.game.Guesses) != 0 {
		t.Fatal("unsubmitted input changed accepted guesses")
	}

	for range game.WordLength {
		model, _ = pressKey(t, model, tea.KeyBackspace)
	}
	model = pressText(t, model, "zzzzz")
	model, _ = pressKey(t, model, tea.KeyEnter)
	if model.message != game.ErrWordNotFound.Error() || len(model.game.Guesses) != 0 {
		t.Fatalf("unknown guess = message:%q attempts:%d", model.message, len(model.game.Guesses))
	}

	model, _ = pressKey(t, model, tea.KeyEscape)
	if model.screen != screenHome || model.input != "" || model.message != "" {
		t.Fatalf("escape = screen:%v input:%q message:%q", model.screen, model.input, model.message)
	}
	_, cmd := model.Update(tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl}))
	if cmd == nil {
		t.Fatal("ctrl+c did not quit")
	}
}

func TestKeyboardUsesStrongestFeedback(t *testing.T) {
	guesses := []game.EvaluatedGuess{
		{Tiles: [game.WordLength]game.Tile{
			{Letter: 'a', State: game.Absent},
			{Letter: 'b', State: game.Correct},
			{Letter: 'c', State: game.Present},
			{Letter: 'a', State: game.Present},
		}},
		{Tiles: [game.WordLength]game.Tile{
			{Letter: 'a', State: game.Correct},
			{Letter: 'b', State: game.Absent},
		}},
	}
	states := keyboardStates(guesses)
	if states['a'] != game.Correct || states['b'] != game.Correct || states['c'] != game.Present {
		t.Fatalf("strongest keyboard states = %v", states)
	}
	if _, used := states['z']; used {
		t.Fatal("unused letter was marked used")
	}

	model := Model{game: &game.Game{Guesses: guesses}}
	for _, width := range []int{58, 100} {
		keyboard := model.renderKeyboard(width, width >= 78)
		for _, letter := range "QWERTYUIOPASDFGHJKLZXCVBNM" {
			if !strings.Contains(keyboard, string(letter)) {
				t.Fatalf("keyboard width %d is missing %c", width, letter)
			}
		}
		if lipgloss.Width(keyboard) > width {
			t.Fatalf("keyboard width %d overflows to %d", width, lipgloss.Width(keyboard))
		}
	}
}

func TestResponsiveGameLayouts(t *testing.T) {
	model := setupTestModel(t)
	model.screen = screenGame
	model.game = startTestGame(t)
	model.stats = game.Stats{Played: 12, Wins: 11, WinPercentage: 91, CurrentStreak: 3}

	for _, width := range []int{60, 79, 80, 119, 120, 160} {
		t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
			candidate := model
			candidate.width = width
			candidate.height = 30
			content := candidate.View().Content
			if got := lipgloss.Width(content); got > width {
				t.Fatalf("content width = %d, terminal width = %d", got, width)
			}
			if got := lipgloss.Height(content); got > candidate.height {
				t.Fatalf("content height = %d, terminal height = %d", got, candidate.height)
			}
			if strings.Count(content, "·") != game.WordLength*game.MaxAttempts {
				t.Fatal("responsive layout does not contain the complete board")
			}
			if !strings.Contains(content, "Enter your guess") {
				t.Fatal("responsive layout is missing the input prompt")
			}

			switch {
			case width < 80:
				if strings.Contains(content, model.game.PuzzleDate) || strings.Contains(content, "Win 91%") {
					t.Fatal("small layout contains hidden metadata")
				}
			case width < 120:
				if !strings.Contains(content, model.game.PuzzleDate) || !strings.Contains(content, "Win 91%") {
					t.Fatal("medium layout is missing date or win rate")
				}
				if strings.Contains(content, "Guess distribution") {
					t.Fatal("medium layout contains guess distribution")
				}
			default:
				if !strings.Contains(content, model.game.PuzzleDate) || !strings.Contains(content, "Guess distribution") {
					t.Fatal("large layout is missing date or distribution")
				}
				if strings.Contains(content, strings.Repeat("─", 111)) {
					t.Fatal("large layout contains a divider wider than the bounded canvas")
				}
				if !strings.Contains(content, strings.Repeat("─", 110)) {
					t.Fatal("large layout is missing the 110-column canvas divider")
				}
				if got := lipgloss.Width(candidate.renderGameBody(110, 28, width)); got > 110 {
					t.Fatalf("large play area width = %d, canvas width = 110", got)
				}
			}
		})
	}
}

func TestShortGameKeepsBoardAndPrompt(t *testing.T) {
	model := setupTestModel(t)
	model.screen = screenGame
	model.game = startTestGame(t)
	model.width = 60
	model.height = 14
	content := model.View().Content
	if strings.Count(content, "·") != game.WordLength*game.MaxAttempts || !strings.Contains(content, "Enter your guess") {
		t.Fatal("short layout dropped the board or prompt")
	}
	if strings.Contains(content, "Q W E R T Y") {
		t.Fatal("short layout did not prune the keyboard")
	}
	if got := lipgloss.Height(content); got > model.height {
		t.Fatalf("short layout height = %d, terminal height = %d", got, model.height)
	}
}

func TestResultAndStatsNavigation(t *testing.T) {
	model := setupTestModel(t)
	current := startTestGame(t)
	puzzle, err := db.GetPuzzleByDate(current.PuzzleDate)
	if err != nil {
		t.Fatal(err)
	}
	if err := current.SubmitGuess(puzzle.Word); err != nil {
		t.Fatal(err)
	}
	updated, _ := model.Update(gameStartedMsg{game: current})
	model = updated.(Model)
	model.width = 80
	model.height = 30
	content := model.View().Content
	if !strings.Contains(content, "You got it!") {
		t.Fatal("result screen is missing the outcome")
	}
	for _, letter := range strings.ToUpper(puzzle.Word) {
		if !strings.Contains(content, string(letter)) {
			t.Fatalf("result screen is missing answer letter %c", letter)
		}
	}
	if strings.Count(content, "·") != (game.MaxAttempts-len(current.Guesses))*game.WordLength {
		t.Fatal("result screen is missing the restored six-row board")
	}

	model, cmd := pressKey(t, model, 's')
	if cmd == nil || model.message != "Copied result to clipboard" {
		t.Fatal("s did not copy the completed result")
	}
	if got, want := fmt.Sprint(cmd()), formatShareResult(current); got != want {
		t.Fatalf("clipboard content = %q, want %q", got, want)
	}

	model.resultIndex = 1
	model, _ = pressKey(t, model, tea.KeyEnter)
	if model.screen != screenStats || model.previousScreen != screenResult {
		t.Fatalf("View Stats opened screen %v from %v", model.screen, model.previousScreen)
	}
	if content := model.View().Content; !strings.Contains(content, "Win Rate") || !strings.Contains(content, "Guess Distribution") {
		t.Fatal("stats screen is missing full statistics")
	}
	model, _ = pressKey(t, model, tea.KeyEscape)
	if model.screen != screenResult {
		t.Fatalf("stats escape returned to %v, want result", model.screen)
	}
	model.resultIndex = 2
	model, _ = pressKey(t, model, tea.KeyEnter)
	if model.screen != screenHome {
		t.Fatalf("Back to Menu returned to %v", model.screen)
	}
}

func TestCompletedLossResultRestoresBoard(t *testing.T) {
	model := setupTestModel(t)
	current := startTestGame(t)
	puzzle, err := db.GetPuzzleByDate(current.PuzzleDate)
	if err != nil {
		t.Fatal(err)
	}
	wordCount, err := db.GetWordCount()
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < wordCount && !current.Done; index++ {
		word, getErr := db.GetWordByIndex(index)
		if getErr != nil {
			t.Fatal(getErr)
		}
		if word == puzzle.Word {
			continue
		}
		if submitErr := current.SubmitGuess(word); submitErr != nil {
			t.Fatal(submitErr)
		}
	}
	if !current.Done || current.Won || len(current.Guesses) != game.MaxAttempts {
		t.Fatalf("loss = done:%v won:%v guesses:%d", current.Done, current.Won, len(current.Guesses))
	}

	updated, _ := model.Update(gameStartedMsg{game: current})
	model = updated.(Model)
	model.width = 80
	model.height = 30
	content := model.View().Content
	if model.screen != screenResult || !strings.Contains(content, "Not quite") ||
		!strings.Contains(content, "Answer "+strings.ToUpper(puzzle.Word)) || !strings.Contains(content, "X/6") {
		t.Fatal("completed loss result is missing its restored outcome")
	}
	if strings.Contains(content, "·") {
		t.Fatal("six-attempt loss rendered empty board tiles")
	}
}

func TestShareResultFormatting(t *testing.T) {
	win := &game.Game{
		PuzzleDate: "2026-10-06",
		Done:       true,
		Won:        true,
		Guesses: []game.EvaluatedGuess{
			{Word: "stare", Tiles: [game.WordLength]game.Tile{
				{State: game.Absent}, {State: game.Present}, {State: game.Absent}, {State: game.Absent}, {State: game.Correct},
			}},
			{Word: "river", Tiles: [game.WordLength]game.Tile{
				{State: game.Correct}, {State: game.Correct}, {State: game.Correct}, {State: game.Correct}, {State: game.Correct},
			}},
		},
	}
	want := "WORDLY 2026-10-06 2/6\n\n⬛🟨⬛⬛🟩\n🟩🟩🟩🟩🟩\n\n" +
		"I played today's Wordly — can you solve it too?\nhttps://github.com/nikhil25803/wordly"
	if got := formatShareResult(win); got != want {
		t.Fatalf("win share = %q, want %q", got, want)
	}
	if strings.Contains(strings.ToLower(formatShareResult(win)), "river") {
		t.Fatal("share text revealed the answer")
	}

	loss := *win
	loss.Won = false
	loss.Guesses = make([]game.EvaluatedGuess, game.MaxAttempts)
	if got := formatShareResult(&loss); !strings.HasPrefix(got, "WORDLY 2026-10-06 X/6\n") {
		t.Fatalf("loss share score = %q", got)
	}
	if got := formatShareResult(&game.Game{}); got != "" {
		t.Fatalf("unfinished share = %q, want empty", got)
	}
}

func TestRenderStatsIsPlainAndComplete(t *testing.T) {
	stats := game.Stats{Played: 3, Wins: 2, WinPercentage: 66, CurrentStreak: 1, MaxStreak: 2}
	stats.Distribution = [game.MaxAttempts]int{1, 0, 1, 0, 0, 0}
	output := RenderStats(stats)
	if strings.Contains(output, "\x1b[") {
		t.Fatal("plain stats output contains ANSI escapes")
	}
	if !strings.Contains(output, "Played 3  Wins 2  Win 66%  Streak 1  Max 2") {
		t.Fatalf("stats totals missing from %q", output)
	}
	for attempt := 1; attempt <= game.MaxAttempts; attempt++ {
		if !strings.Contains(output, fmt.Sprintf("%d  ", attempt)) {
			t.Fatalf("stats output is missing attempt %d", attempt)
		}
	}
}
