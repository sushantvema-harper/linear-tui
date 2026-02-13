# linear-tui

A terminal user interface (TUI) for Linear built with Go and tview.

## Screenshots

![Main interface](docs/main.jpeg)

![Create issue](docs/create.png)

![Assign issue](docs/assign.png)

## Demo

![Agent demo](docs/agent-demo.gif)

## Features

- 3-pane layout (navigation tree + issues list + details view)
- Command palette for quick actions with keyboard shortcuts
- Vim-style keyboard navigation (Ctrl+N/P, Ctrl+D/U, h/l, gg/G, g-prefix motions)
- Mouse support (click to focus, scroll to navigate)
- Issue descriptions with markdown rendering
- Sub-issues support (expand/collapse, create, view parent)
- Issue management (create, edit title, edit labels, archive)
- Comments (view, add, browse with search/copy/open)
- Attachments browser (browse, search, copy URL, open)
- Status management (change status, assign/unassign)
- Search and filtering
- Sorting (by updated, created, or priority)
- My Issues vs Other Issues sections
- Agent runs via command palette (Claude or Cursor Agent)
- Agent prompt templates and streaming output with copy/resume
- Real-time issue fetching from Linear API
- Comprehensive logging system for debugging
- Settings modal with live config updates
- Themes (linear, high_contrast, color_blind) and density modes
- Status bar with context and search info
- Clipboard actions (issue ID, issue URL, agent output)

## Requirements

- Linear API key (set as `LINEAR_API_KEY` environment variable)
- Agent CLI for the agent command:
  - Claude provider: `claude`
  - Cursor provider: `cursor-agent` (preferred) or `agent`

## Configuration

- `LINEAR_API_KEY` is required (the API key is not stored on disk).
- Settings are stored in `~/.linear-tui/config.json` and created on first start.
- Use the Settings modal from the command palette (`:` -> `Settings`) to edit and apply settings immediately.
- UI settings in `config.json`: `theme` (`linear`, `high_contrast`, `color_blind`) and `density` (`comfortable`, `compact`).
- Agent settings live in `config.json`: `agent_provider` (`cursor` or `claude`), `agent_sandbox` (`enabled` or `disabled`), `agent_model` (optional), and `agent_workspace` (optional).
- Prompt templates are stored in `~/.linear-tui/prompts.json` and edited via the "Edit agent prompt templates" command.
- `agent_workspace` is the default workspace for agent runs and can be overridden per run in the Ask Agent modal (overrides are not persisted).

Example `~/.linear-tui/config.json`:

```json
{
  "api_endpoint": "https://api.linear.app/graphql",
  "timeout": "30s",
  "page_size": 50,
  "cache_ttl": "5m",
  "log_file": "/Users/you/.linear-tui/app.log",
  "log_level": "warning",
  "theme": "linear",
  "density": "comfortable",
  "agent_provider": "cursor",
  "agent_sandbox": "enabled",
  "agent_model": "",
  "agent_workspace": ""
}
```

## Installation

### From Source (Recommended)

Requires Go 1.24 or later.

```bash
git clone https://github.com/sushantvema-harper/linear-tui.git
cd linear-tui
./scripts/install.sh
```

This builds the binary and symlinks it to `~/.local/bin/linear-tui`. Make sure `~/.local/bin` is on your `$PATH`:

```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.) if not already present
export PATH="${HOME}/.local/bin:${PATH}"
```

### Updating

```bash
cd linear-tui
git pull && ./scripts/install.sh
```

### Download Binary

Download pre-built binaries from the [Releases](https://github.com/sushantvema-harper/linear-tui/releases) page.

## Usage

### Basic Usage

Set your Linear API key and run the application:

```bash
export LINEAR_API_KEY="your-api-key-here"
./linear-tui
```

### Advanced Configuration

Example `~/.linear-tui/config.json`:

```json
{
  "api_endpoint": "https://api.linear.app/graphql",
  "timeout": "30s",
  "page_size": 50,
  "cache_ttl": "5m",
  "log_file": "/Users/you/.linear-tui/app.log",
  "log_level": "warning",
  "theme": "linear",
  "density": "comfortable",
  "agent_provider": "cursor",
  "agent_sandbox": "enabled",
  "agent_model": "",
  "agent_workspace": ""
}
```

### Disable Logging

To disable logging, set `log_file` to an empty string in the settings file or via the Settings modal:

```json
{
  "log_file": ""
}
```

## Keyboard Shortcuts

### Global

| Key | Action |
|-----|--------|
| `:` | Open command palette |
| `/` | Open search palette |
| `Tab` / `Shift+Tab` | Cycle between panes |
| `Ctrl+B` | Toggle navigation pane |
| `Esc` | Close palette / Clear search |
| `q` | Quit |

### Navigation Tree (left pane)

| Key | Action |
|-----|--------|
| `Ctrl+N` | Next node |
| `Ctrl+P` | Previous node |
| `Ctrl+D` | Half-page down |
| `Ctrl+U` | Half-page up |
| `Enter` | Select team/project/status filter |
| `l` / `→` | Focus issues pane |

### Issues Table (center pane)

#### Movement

| Key | Action |
|-----|--------|
| `Ctrl+N` | Next issue |
| `Ctrl+P` | Previous issue |
| `Ctrl+D` | Half-page down |
| `Ctrl+U` | Half-page up |
| `j` | Focus Other Issues section (from My Issues) |
| `k` | Focus My Issues section (from Other Issues) |
| `gg` | Go to top of section |
| `G` | Go to bottom of section |
| `Space` | Toggle expand/collapse sub-issues |
| `Enter` | Toggle expand (parent) / Focus details (leaf) |

#### Pane Navigation

| Key | Action |
|-----|--------|
| `h` / `←` | Focus navigation pane |
| `l` / `→` | Focus details pane |

#### Motions (g-prefix)

| Key | Action |
|-----|--------|
| `gx` | Open selected issue in browser |
| `gy` | Copy issue identifier to clipboard |

#### Quick Commands (from issues pane)

| Key | Action |
|-----|--------|
| `r` | Refresh issues |
| `n` | Create new issue |
| `e` | Edit issue title |
| `E` | Edit issue description in `$EDITOR` |
| `Ctrl+L` | Edit issue labels |
| `s` | Change status |
| `a` | Assign to user |
| `m` | Assign to me |
| `u` | Unassign issue |
| `t` | Add comment |
| `o` | Open in browser |
| `y` | Copy issue ID |
| `w` | Copy issue URL |
| `x` | Archive issue |
| `b` | Create sub-issue |
| `p` | View parent issue |
| `i` | Set parent issue |
| `d` | Remove parent |
| `]` | Expand all sub-issues |
| `[` | Collapse all sub-issues |
| `c` | Browse comments |
| `A` | Browse attachments |
| `f` | Fullscreen issue description |

### Details Pane (right pane)

| Key | Action |
|-----|--------|
| `Ctrl+N` | Scroll down 1 line |
| `Ctrl+P` | Scroll up 1 line |
| `Ctrl+D` | Scroll down half-page |
| `Ctrl+U` | Scroll up half-page |
| `Tab` | Switch between description and comments |
| `h` / `←` | Focus issues pane |
| `f` | Fullscreen issue description |
| `E` | Edit issue description in `$EDITOR` |

### Comments Browser

| Key | Action |
|-----|--------|
| `n` / `j` | Next comment |
| `p` / `k` | Previous comment |
| `Ctrl+N` / `Ctrl+P` | Scroll up/down |
| `Ctrl+D` / `Ctrl+U` | Half-page down/up |
| `/` | Toggle search filter |
| `y` | Copy comment link |
| `o` | Open comment in browser |
| `t` | Add new comment |
| `Esc` / `q` | Close browser |

### Attachments Browser

| Key | Action |
|-----|--------|
| `n` / `j` | Next attachment |
| `p` / `k` | Previous attachment |
| `Ctrl+N` / `Ctrl+P` | Scroll up/down |
| `Ctrl+D` / `Ctrl+U` | Half-page down/up |
| `/` | Toggle search filter |
| `y` | Copy attachment URL |
| `o` | Open attachment URL |
| `Esc` / `q` | Close browser |

### Command Palette

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate commands |
| `Enter` | Execute selected command / Submit search |
| `Esc` | Close palette |

## Development

Run tests:

```bash
go test ./...
```

Build:

```bash
go build ./cmd/linear-tui
```
