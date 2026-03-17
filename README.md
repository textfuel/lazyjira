# lazyjira

A terminal UI for Jira. Like [lazygit](https://github.com/jesseduffield/lazygit) but for Jira.

Navigate issues, switch projects, read descriptions and comments, all without leaving the terminal.

## Install

```
go install github.com/textfuel/lazyjira/cmd/lazyjira@latest
```

Or build from source:

```
git clone https://github.com/textfuel/lazyjira.git
cd lazyjira
go build -o lazyjira ./cmd/lazyjira
```

## Setup

Run `lazyjira`. On first launch it will ask for your Jira host, email, and API token.

You can create an API token at https://id.atlassian.com/manage-profile/security/api-tokens

Credentials are saved to `~/.config/lazyjira/auth.json` (or `$XDG_CONFIG_HOME/lazyjira/`).

To re-authenticate: `lazyjira auth`

To logout: `lazyjira logout`

## Usage

```
lazyjira              # start the TUI
lazyjira --log app.log   # log API requests to file
lazyjira --dry-run       # read-only mode, no write requests
```

### Layout

Left side has three panels stacked vertically:
- `[1]` Status - connection info
- `[2]` Issues - issue list for current project
- `[3]` Projects - project switcher

Right side:
- Detail panel with tabs (description, subtasks, comments, links, info)
- Command log showing API requests

### Keybindings

| Key | Action |
|-----|--------|
| `1` `2` `3` | focus left panel |
| `0` | focus right panel |
| `tab` | switch left/right |
| `j/k` | navigate or scroll |
| `enter` `l` | open issue / select project |
| `h` | back to left panel |
| `[` `]` | switch detail tabs |
| `i` | jump to info tab |
| `/` | search / filter |
| `r` | refresh |
| `?` | show all keybindings |
| `q` | quit |

Press `?` in any context for the full list.

## Config

Optional config file at `~/.config/lazyjira/config.yml`:

```yaml
gui:
  sidePanelWidth: 40
  mouse: true

cache:
  enabled: true
  ttl: 5m

refresh:
  autoRefresh: true
  interval: 30s
```

## License

MIT
