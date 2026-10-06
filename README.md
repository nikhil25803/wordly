<div align="center">

# Wordly

**One shared Wordle every UTC day, played entirely in your terminal.**

[![Go version](https://img.shields.io/github/go-mod/go-version/nikhil25803/wordly?logo=go)](go.mod)
[![CI](https://github.com/nikhil25803/wordly/actions/workflows/ci.yml/badge.svg)](https://github.com/nikhil25803/wordly/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/nikhil25803/wordly)](https://github.com/nikhil25803/wordly/releases/latest)
[![License](https://img.shields.io/github/license/nikhil25803/wordly)](LICENSE)
[![Bubble Tea](https://img.shields.io/badge/Bubble_Tea-v2-ff69b4)](https://github.com/charmbracelet/bubbletea)
[![Lip Gloss](https://img.shields.io/badge/Lip_Gloss-v2-7d56f4)](https://github.com/charmbracelet/lipgloss)

</div>

![Wordly terminal game banner](docs/assets/Banner.png)

## Contents

- [Overview](#overview)
- [Demo](#demo)
- [Features](#features)
- [Install](#install)
- [Commands](#commands)
- [Controls](#controls)
- [Sharing](#sharing)
- [Supported platforms](#supported-platforms)
- [Local data](#local-data)
- [Development](#development)
- [License](#license)

## Overview

Wordly is a daily, six-attempt word game for the command line. Everyone receives the same puzzle for the UTC calendar day. Accepted guesses, unfinished progress, and statistics are saved locally for the current operating-system user—there is no account, network service, or telemetry.

## Demo

<p align="center">
  <strong>Home</strong><br>
  <img src="docs/assets/home.svg" alt="Wordly home menu" width="620">
</p>

<p align="center">
  <strong>Daily game</strong><br>
  <img src="docs/assets/gameplay.svg" alt="Responsive Wordly daily game" width="900">
</p>

<p align="center">
  <strong>Completed result</strong><br>
  <img src="docs/assets/result.svg" alt="Completed Wordly result with sharing option" width="620">
</p>

The interface adapts its board, keyboard, and statistics to the terminal width. Submitted guesses resume automatically, and reopening a finished daily game restores its complete Result screen.

## Features

- Responsive Home, Game, Result, and Statistics screens built with Bubble Tea and Lip Gloss
- Bounded large-screen layout with compact medium and small terminal variants
- On-screen QWERTY keyboard that retains the strongest feedback for every used letter
- Correct handling of repeated letters
- Dictionary validation without consuming an attempt for invalid guesses
- Automatic resume of submitted guesses and completed daily boards
- Spoiler-free result sharing through the terminal clipboard
- Daily and all-time win statistics with streaks and guess distribution
- SQLite persistence embedded with the application

## Install

### macOS, Linux, or Windows Git Bash

```sh
curl -fsSL https://raw.githubusercontent.com/nikhil25803/wordly/main/install.sh | sh
```

### Windows PowerShell

```powershell
curl.exe -fsSL https://raw.githubusercontent.com/nikhil25803/wordly/main/install.ps1 | Out-String | Invoke-Expression
```

Both installers detect the operating system and architecture, download the latest release with curl, verify its SHA-256 checksum, and install it to `~/.local/bin` by default.

To choose another directory:

```sh
curl -fsSL https://raw.githubusercontent.com/nikhil25803/wordly/main/install.sh | WORDLY_INSTALL_DIR="$HOME/bin" sh
```

```powershell
$env:WORDLY_INSTALL_DIR = "$HOME\bin"
curl.exe -fsSL https://raw.githubusercontent.com/nikhil25803/wordly/main/install.ps1 | Out-String | Invoke-Expression
```

Ensure the selected directory is on your `PATH`, then launch the game with `wordly`.

## Commands

| Command          | Purpose                                                                      | Example                                                                      |
| ---------------- | ---------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| `wordly --help`  | Show CLI usage and available flags.                                          | <img src="docs/assets/help.svg" alt="wordly help output" width="440">        |
| `wordly --stats` | Print the current user's completed-game statistics without starting the TUI. | <img src="docs/assets/stats.svg" alt="wordly stats output" width="440">      |
| `wordly --words` | Print the number of words in the embedded dictionary.                        | <img src="docs/assets/words.svg" alt="wordly word count output" width="440"> |
| `wordly --reset` | Delete only the current user's guesses and game history.                     | <img src="docs/assets/reset.svg" alt="wordly reset output" width="440">      |

`--stats`, `--words`, and `--reset` are mutually exclusive. Running `wordly` without a flag opens or resumes today's game.

## Controls

| Key             | Action                                           |
| --------------- | ------------------------------------------------ |
| Letters         | Fill the current five-letter guess               |
| Backspace       | Remove the last letter                            |
| Arrow keys, j/k | Navigate menus                                   |
| Enter           | Submit a guess or select a menu item              |
| S               | Copy a completed result from the Result screen    |
| Esc             | Return to the previous screen                     |
| Ctrl+C          | Quit while preserving submitted guesses          |

Correct letters are **bold and green**, present letters are <u>underlined and yellow</u>, and absent letters are dim and gray. The legend remains visible so meaning is not conveyed by color alone.

## Sharing

From the completed Result screen, select **Share Result** or press `s`. Wordly copies a spoiler-free grid using the UTC puzzle date and number of attempts:

```text
WORDLY 2026-10-06 5/6

⬛🟩⬛⬛⬛
⬛🟨🟨⬛⬛
⬛⬛⬛⬛⬛
🟩🟩⬛🟩🟩
🟩🟩🟩🟩🟩

I played today's Wordly — can you solve it too?
https://github.com/nikhil25803/wordly
```

Clipboard copying uses OSC52, supported by most modern terminals.

## Supported platforms

| Operating system | Architectures | Archive   | Installer              |
| ---------------- | ------------- | --------- | ---------------------- |
| Linux            | amd64, arm64  | `.tar.gz` | POSIX shell            |
| macOS            | amd64, arm64  | `.tar.gz` | POSIX shell            |
| Windows          | amd64, arm64  | `.zip`    | PowerShell or Git Bash |

## Local data

Wordly stores one SQLite database in the operating system's user configuration directory:

| Platform | Default location                                                    |
| -------- | ------------------------------------------------------------------- |
| Linux    | `$XDG_CONFIG_HOME/wordly/wordly.db` or `~/.config/wordly/wordly.db` |
| macOS    | `~/Library/Application Support/wordly/wordly.db`                    |
| Windows  | `%AppData%\wordly\wordly.db`                                        |

`wordly --reset` removes the current user's gameplay records while preserving the user registration, dictionary, and daily puzzles.

## Development

Wordly requires the Go version declared in [go.mod](go.mod).

```sh
go test ./...
go test -race ./...
go vet ./...
go run ./cmd/wordly
```

## License

Wordly is available under the [MIT License](LICENSE).
