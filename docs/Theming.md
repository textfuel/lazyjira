# Theming

lazyjira ships with five built-in themes (`default`, `catppuccin-latte`,
`catppuccin-frappe`, `catppuccin-macchiato`, `catppuccin-mocha`) and can also
load themes from YAML files you write yourself.

A theme is a generic, shareable description of how the app should look. It
does not know anything about your particular Jira instance - that's what
`config.yml` is for (see [Config](./Config.md) for instance-specific colors
like `projectKeyColors`, `assigneeColors`, `typeColors`, and
`selectedForeground`).

## Where themes live

User-defined themes are loaded by basename from:

```
<config-dir>/themes/<name>.yml
```

The config directory is the same one that holds `config.yml`. Set
`gui.theme: my-theme` to load `<config-dir>/themes/my-theme.yml`.

If `<name>` matches a built-in (`default`, `catppuccin-*`) the built-in
theme is used and no file is read.

## Schema

```yaml
name: catppuccin-mocha     # informational; filename is what counts

palette:                    # free-form color variables named by you
  rosewater: "#f5e0dc"
  flamingo:  "#f2cdcd"
  pink:      "#f5c2e7"
  mauve:     "#cba6f7"
  red:       "#f38ba8"
  peach:     "#fab387"
  yellow:    "#f9e2af"
  green:     "#a6e3a1"
  teal:      "#94e2d5"
  blue:      "#89b4fa"
  text:      "#cdd6f4"
  overlay0:  "#6c7086"
  surface2:  "#585b70"

styles:                     # fixed set of semantic slots
  title:          green
  subtitle:       overlay0
  hintBar:        overlay0     # optional, defaults to muted
  accent:         green        # active borders/tabs/markers, key labels
  muted:          overlay0     # separators, placeholders, ellipsis
  errorText:      red
  successText:    green
  warningText:    yellow
  selectedItemBg: surface2
  issueKey:       teal
  activeBorder:   green        # optional, defaults to accent
  inactiveBorder: "-1"         # optional, defaults to terminal default

types:                      # optional; missing entries fall back to _fallback
  Bug:         red
  Story:       green
  Epic:        mauve
  Task:        blue
  Sub-task:    overlay0
  Improvement: teal
  New Feature: peach
  _fallback:   overlay0

priorities:                 # optional
  Highest: red
  High:    peach
  Medium:  yellow
  Low:     green
  Lowest:  overlay0

statuses:                   # optional; keyed on Jira status category key
  done:          green
  indeterminate: yellow
  new:           blue

authorPalette:              # optional; falls back to default if omitted
  - rosewater
  - flamingo
  - pink
  - mauve
  - red
  - peach
  - yellow
  - green
  - teal
  - blue
```

### Required style slots

`title`, `subtitle`, `accent`, `muted`, `errorText`, `successText`,
`warningText`, `selectedItemBg`, `issueKey`. All others are optional.

### `_fallback`

In `types:`, `priorities:`, and `statuses:`, an entry named `_fallback`
sets the color used for values not enumerated in the map. If absent, the
`muted` style is used as the fallback.

## Color value resolution

Wherever a color value appears (in `styles:`, `types:`, anywhere), the
value is resolved as follows:

1. If it matches a palette variable name, use that palette color.
2. Otherwise, pass the value through to lipgloss (`#rrggbb`, ANSI numeric
   codes, or named colors all work).

There is no chaining: a style slot value cannot reference another style
slot.

## Resolution chain at render time

For any colored cell, the rendered color is the first match in:

1. User config override in `config.yml`, if the cell's name is in the
   relevant override map.
2. The corresponding theme entry, if the cell's name is in the relevant
   theme map.
3. The fallback (`_fallback` for type/priority/status maps; the relevant
   styles slot for non-mapped fields).

## Example

See [`docs/examples/themes/catppuccin-mocha.yml`](./examples/themes/catppuccin-mocha.yml)
for a complete worked example you can copy as a starting point.

## Out of scope

The following call sites are not yet themed via YAML and still use the
built-in ANSI 16 palette: JQL syntax highlighting, ADF rendering, diff
views, and detail-view cursor/link styles.
