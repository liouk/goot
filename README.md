# goot

A minimal TUI for quickly adding tasks to Google Tasks.

## Setup

1. Create a Google Cloud project and enable the [Tasks API](https://console.cloud.google.com/apis/library/tasks.googleapis.com)
2. Create an OAuth 2.0 credential (Application type: **Desktop app**)
3. Download the credentials JSON and place it at `~/.config/goot/credentials.json`

On first run, `goot` will open your browser for OAuth consent. The token is cached at `~/.config/goot/token.json` for subsequent runs.

## Install

```
go install github.com/liouk/goot@latest
```

Or build from source:

```
go build -o goot .
```

## Usage

```
goot [list-name]
```

- No arguments: starts at the list selection screen with fuzzy filtering
- With a list name: skips selection and jumps straight to task creation (case-insensitive match; falls back to selection if no match)

## Keybindings

| Key | Action |
|-----|--------|
| `/` | Filter lists (picker screen) |
| `tab` / `shift+tab` | Navigate between form fields |
| `enter` | Select list / submit task |
| `esc` | Quit with confirmation (creator screen) |
| `ctrl+c` | Quit immediately |

## Task fields

| Field | Required | Format |
|-------|----------|--------|
| Title | yes | free text |
| Notes | no | free text |
| Due | no | `YYYY-MM-DD` (auto-fills today on first focus) |

Due dates are interpreted as CET/CEST.
