package ui

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nikhil25803/wordly/internal/db"
	"github.com/nikhil25803/wordly/internal/game"
)

func setupModel(t *testing.T) Model {
	t.Helper()
	t.Setenv("WORDLY_DB", filepath.Join(t.TempDir(), "wordly.db"))
	if err := db.ConnectDatabase(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.DB.Close() })
	currentGame, err := game.StartGame()
	if err != nil {
		t.Fatal(err)
	}
	return NewModel(currentGame)
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

func TestModelInputAndValidation(t *testing.T) {
	model := setupModel(t)
	model = pressText(t, model, "qwerty")
	if model.input != "qwert" {
		t.Fatalf("input = %q, want qwert", model.input)
	}
	if len(usedLetters(model.game.Guesses)) != 0 {
		t.Fatal("unsubmitted input changed keyboard state")
	}

	model, _ = pressKey(t, model, tea.KeyBackspace)
	if model.input != "qwer" {
		t.Fatalf("input after backspace = %q", model.input)
	}
	model, _ = pressKey(t, model, tea.KeyEnter)
	if model.message != game.ErrGuessLength.Error() {
		t.Fatalf("short guess message = %q", model.message)
	}

	model, _ = pressKey(t, model, tea.KeyBackspace)
	model, _ = pressKey(t, model, tea.KeyBackspace)
	model, _ = pressKey(t, model, tea.KeyBackspace)
	model, _ = pressKey(t, model, tea.KeyBackspace)
	model = pressText(t, model, "zzzzz")
	model, _ = pressKey(t, model, tea.KeyEnter)
	if model.message != game.ErrWordNotFound.Error() || len(model.game.Guesses) != 0 {
		t.Fatalf("unknown guess = message:%q attempts:%d", model.message, len(model.game.Guesses))
	}
	if len(usedLetters(model.game.Guesses)) != 0 {
		t.Fatal("invalid guess changed keyboard state")
	}
	if !strings.Contains(model.View().Content, "bold=correct") {
		t.Fatal("view is missing the non-color tile legend")
	}
	if !strings.Contains(model.View().Content, "keyboard: dim=used") {
		t.Fatal("view is missing the non-color keyboard legend")
	}
}

func TestKeyboardMarksAcceptedLettersUsed(t *testing.T) {
	guesses := []game.EvaluatedGuess{
		{Tiles: [game.WordLength]game.Tile{
			{Letter: 'a', State: game.Absent},
			{Letter: 'b', State: game.Correct},
			{Letter: 'c', State: game.Present},
			{Letter: 'd', State: game.Absent},
			{Letter: 'a', State: game.Present},
		}},
		{Tiles: [game.WordLength]game.Tile{
			{Letter: 'a', State: game.Correct},
			{Letter: 'b', State: game.Absent},
		}},
	}
	used := usedLetters(guesses)
	for _, letter := range []byte{'a', 'b', 'c', 'd'} {
		if !used[letter] {
			t.Fatalf("accepted letter %c is not marked used", letter)
		}
	}
	if used['z'] {
		t.Fatal("unused letter is marked as used")
	}
}

func TestKeyboardLayoutAndStyles(t *testing.T) {
	model := setupModel(t)
	compact := model.renderKeyboard()
	if height := lipgloss.Height(compact); height != 6 {
		t.Fatalf("compact keyboard height = %d, want 6", height)
	}
	if width := lipgloss.Width(compact); width != 33 {
		t.Fatalf("compact keyboard width = %d, want 33", width)
	}
	plain := regexp.MustCompile(`\x1b\[[0-9:;]*m`).ReplaceAllString(compact, "")
	lines := strings.Split(plain, "\n")
	for row, want := range []string{"QWERT", "YUIOP", "ASDFG", "HJKLZ", "XCVBN", "M"} {
		if got := strings.ReplaceAll(strings.TrimSpace(lines[row]), " ", ""); got != want {
			t.Fatalf("keyboard row %d = %q, want %q", row, got, want)
		}
	}

	model.width = 40
	model.height = 48
	large := model.renderKeyboard()
	if height := lipgloss.Height(large); height != 18 {
		t.Fatalf("large keyboard height = %d, want 18", height)
	}
	for _, letter := range "QWERTYUIOPASDFGHJKLZXCVBNM" {
		if !strings.Contains(large, string(letter)) {
			t.Fatalf("keyboard is missing %c", letter)
		}
	}

	if style := keyboardKeyStyle(false, false); style.GetBold() || style.GetUnderline() || style.GetFaint() {
		t.Fatal("unused key has feedback styling")
	}
	if style := keyboardKeyStyle(true, false); !style.GetFaint() || style.GetBold() || style.GetUnderline() {
		t.Fatal("used key is not dim-only")
	}
}

func TestPlayAreaResponsiveLayout(t *testing.T) {
	model := setupModel(t)
	model.width = 80
	model.height = 40
	compact := model.renderPlayArea()
	if height := lipgloss.Height(compact); height != 13 {
		t.Fatalf("compact play area height = %d, want 13", height)
	}
	if width := lipgloss.Width(compact); width != 33 {
		t.Fatalf("compact play area width = %d, want 33", width)
	}

	model.height = 48
	large := model.renderPlayArea()
	if height := lipgloss.Height(large); height != 37 {
		t.Fatalf("large play area height = %d, want 37", height)
	}
	if width := lipgloss.Width(large); width != 33 {
		t.Fatalf("large play area width = %d, want 33", width)
	}
}

func TestKeyboardRestoresAcceptedGuesses(t *testing.T) {
	model := setupModel(t)
	model.game.Guesses = []game.EvaluatedGuess{{Tiles: [game.WordLength]game.Tile{
		{Letter: 'r', State: game.Correct},
		{Letter: 'e', State: game.Present},
		{Letter: 's', State: game.Absent},
	}}}
	used := usedLetters(NewModel(model.game).game.Guesses)
	if !used['r'] || !used['e'] || !used['s'] {
		t.Fatalf("restored keyboard letters = %v", used)
	}
}

func TestModelQuitControls(t *testing.T) {
	model := setupModel(t)
	model = pressText(t, model, "q")
	if model.input != "q" {
		t.Fatal("q did not remain a playable letter")
	}

	model.game.Done = true
	_, cmd := pressKey(t, model, 'q')
	if cmd == nil {
		t.Fatal("q did not quit a completed game")
	}
	_, cmd = pressKey(t, model, tea.KeyEscape)
	if cmd == nil {
		t.Fatal("escape did not quit")
	}
}

func TestModelLayoutIsLeftAlignedAndGrouped(t *testing.T) {
	model := setupModel(t)
	compact := model.View().Content
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 200, Height: 48})
	largeModel := updated.(Model)
	large := largeModel.View().Content
	if lipgloss.Height(largeModel.renderBoard()) != game.MaxAttempts*3 {
		t.Fatal("large terminal did not render three-line tiles")
	}
	if width := lipgloss.Width(largeModel.renderBoard()); width != game.WordLength*5+(game.WordLength-1)*2 {
		t.Fatalf("large board width = %d, want five-column tiles with two-column gaps", width)
	}
	if lipgloss.Height(model.renderBoard()) != game.MaxAttempts {
		t.Fatal("unknown terminal size did not use compact tiles")
	}
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 39, Height: 40})
	if lipgloss.Height(updated.(Model).renderBoard()) != game.MaxAttempts {
		t.Fatal("narrow terminal did not use compact tiles")
	}
	if strings.Count(large, "·") != game.WordLength*game.MaxAttempts {
		t.Fatal("layout does not contain the complete six-row board")
	}
	if !strings.HasPrefix(compact, " ") || !strings.HasPrefix(large, " ") {
		t.Fatal("layout is not left aligned with its one-column margin")
	}
	if len(regexp.MustCompile(`\n +\n`).FindAllString(large, -1)) < 2 {
		t.Fatal("layout is missing blank lines between content groups")
	}
	if !strings.Contains(large, "W O R D L Y") {
		t.Fatal("layout is missing the spaced title")
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

func TestStyledStatsAndCompletedLayout(t *testing.T) {
	stats := game.Stats{Played: 3, Wins: 2, WinPercentage: 66, CurrentStreak: 1, MaxStreak: 2}
	stats.Distribution = [game.MaxAttempts]int{1, 0, 1, 0, 0, 0}
	styled := renderStyledStats(stats)
	if !strings.Contains(styled, "\x1b[") {
		t.Fatal("styled stats contain no terminal styling")
	}
	for _, label := range []string{"Played 3", "Wins 2", "Win 66%", "Streak 1", "Max 2"} {
		if !strings.Contains(styled, label) {
			t.Fatalf("styled stats are missing card %q", label)
		}
	}
	for attempt := 1; attempt <= game.MaxAttempts; attempt++ {
		if !strings.Contains(styled, fmt.Sprintf("%d  ", attempt)) {
			t.Fatalf("styled stats are missing attempt %d", attempt)
		}
	}

	model := setupModel(t)
	model.game.Done = true
	model.game.Won = true
	model.game.Guesses = make([]game.EvaluatedGuess, 5)
	model.game.Stats = stats
	if !strings.Contains(model.renderStatus(), "\n\n") {
		t.Fatal("completed result has no separator before statistics")
	}
	if !strings.Contains(model.View().Content, "Q") {
		t.Fatal("completed layout is missing the keyboard")
	}
	if height := lipgloss.Height(model.View().Content); height > 33 {
		t.Fatalf("completed layout height = %d, want at most 33", height)
	}
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	if height := lipgloss.Height(updated.(Model).View().Content); height > 40 {
		t.Fatalf("large completed layout height = %d, want at most 40", height)
	}
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 58})
	if height := lipgloss.Height(updated.(Model).View().Content); height > 58 {
		t.Fatalf("large-grid completed layout height = %d, want at most 58", height)
	}
}
