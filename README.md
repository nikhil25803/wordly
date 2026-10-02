# wordly
Wordly is a terminal-based Wordle game where everyone gets the same word each UTC day. Progress and statistics stay in a local SQLite database tied to your operating-system username—no account or server required.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/nikhil25803/wordly/main/install.sh | sh
```

The installer supports macOS, Linux, and Windows from a POSIX shell such as Git Bash. It installs to `~/.local/bin` by default; set `WORDLY_INSTALL_DIR` to choose another directory.

## Play

Run `wordly`, type a five-letter word, and press Enter. Correct letters are bold and green, present letters are underlined and yellow, and absent letters are dim and gray. You have six attempts; submitted guesses resume if you leave and return later.

- Backspace: remove a letter
- Enter: submit a guess
- Esc or Ctrl+C: quit and keep submitted progress
- Enter or q: quit after the game finishes

Use `wordly --stats` to print the current user's statistics without launching the game. Use `wordly --reset` to clear the current user's game history and statistics; the dictionary and daily puzzles are preserved. Use `wordly --words` to print the dictionary size.
