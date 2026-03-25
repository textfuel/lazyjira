# Changelog

All notable changes to this project will be documented in this file.

Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.4.0] - 2026-03-25

### Added

- Git integration: create branches from issues with configurable name templates (b key)
- Git integration: search and checkout existing branches by issue key (B key)
- Branch format rules with conditions by issue type (feat/*, fix/*, fallback)
- Auto-detect current issue from branch name
- CHANGELOG.md

## [2.3.0] - 2026-03-24

### Added

- JQL search modal with two-panel UI (input + suggestions/history) (s key)
- JQL autocomplete: field names and values from Jira API
- JQL syntax highlighting in the search input
- JQL history persistence (plain text file, max 50 entries)
- JQL search results appear as a temporary tab in the issues panel
- Custom readline-style text input with cursor, Home/End, Ctrl+A/E/W/K/U
- `make check` target (lint + vet + build)

## [2.2.0] - 2026-03-21

### Added

- Edit fields: transition, priority, assignee changes from TUI (t/p/a keys)
- Comment viewing and posting (c/n keys)
- Input modal component for text entry
- Diff view component for description change history
- ADF-to-Markdown renderer for rich text display in edit/comment workflows

## [2.1.0] - 2026-03-20

### Added

- Rich ADF (Atlassian Document Format) rendering in issue detail
- Support for mentions, emoji, lists, links, code blocks, inline cards
- Windows installation guide in README

## [1.0.0] - 2026-03-18

### Added

- Panel layout inspired by lazygit: Status, Issues, Projects, Detail
- Jira Cloud REST API v3 integration
- Interactive setup wizard on first launch
- Issue list with All/Assigned tabs
- Issue detail with tabs: Body, Sub, Cmt, Lnk, Info, Hist
- Project switcher with auto-fetch from Jira API
- Transition issues (t key) with modal picker
- URL picker (u key) with in-app navigation for Jira links
- History tab with diff for large field changes
- Author color coding consistent across all views
- Search/filter with / key (per-panel)
- Prefetch and cache all issue details for instant navigation
- Auto-refresh every 30 seconds
- Open in browser (o key), copy URL (y key)
- Mouse support: click panels, scroll, click tabs
- Vertical layout for narrow terminals (< 80 cols)
- Responsive side panel width
- Cross-platform: macOS, Linux, Windows
- Homebrew install via tap

[Unreleased]: https://github.com/textfuel/lazyjira/compare/v2.4.0...HEAD
[2.4.0]: https://github.com/textfuel/lazyjira/compare/v2.3.0...v2.4.0
[2.3.0]: https://github.com/textfuel/lazyjira/compare/v2.2.0...v2.3.0
[2.2.0]: https://github.com/textfuel/lazyjira/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/textfuel/lazyjira/compare/v2.0.3...v2.1.0
[1.0.0]: https://github.com/textfuel/lazyjira/releases/tag/v1.1.0
