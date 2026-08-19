# AGENTS.md - Agentic Coding Guidelines

This document provides guidelines for AI agents working in this repository.

## Project Overview

Homepage dashboard with two frontends:
- **Web**: Static HTML/JS generated from YAML — service links with group filtering, search, drag-and-drop, keyboard navigation, customizable settings
- **CLI TUI**: Go terminal UI (bubbletea) — service list with health checks, search, color themes, settings panel

## Build Commands

```bash
make build                     # Builds both web and CLI
make build-html                # Generates output/index.html from services.yml
make build-cmd                 # Builds CLI to output/tui (from cli/)
python3 web/build.py           # Same as make build-html
python3 web/build.py --example # Generates output/example.html
python3 web/build.py --open    # Opens generated file in browser
python3 web/build.py --no-health # Build without health check feature
```

Open `output/index.html` directly in a browser. No build server required.

## Linting/Validation

- **HTML**: Use [W3C Validator](https://validator.w3.org/) or browser DevTools
- **YAML**: Ensure valid YAML syntax in `services.yml`
- **Python**: Run `python3 -m py_compile web/build.py` to check for syntax errors
- **Go**: `cd cli && go build -ldflags="-s -w" -o ../output/tui ./tui` (compile check)

## Testing

- No test framework configured
- Manual browser testing for web, manual terminal testing for TUI

## Code Style Guidelines

### General

- Vanilla HTML/CSS/JS for web — no frameworks
- Go (bubbletea/lipgloss) for CLI TUI
- Keep code in `<script>` tags within HTML files
- Use ES6+ JavaScript features

### HTML

- 2-space indentation
- Include proper meta tags: `<meta charset="UTF-8">`, `<meta name="viewport" content="width=device-width, initial-scale=1.0">`
- Keep CSS in `<style>` tags in `<head>`

### CSS

- Use CSS custom properties (`:root { --variable: value; }`) for theming
- Follow existing color scheme (dark theme: `#0a0a0b` background, `#6366f1` accent)
- Light theme uses `.light` class on `:root`
- Use flexbox for layout
- Use `rem` for sizing, `em` for font-relative sizing

### JavaScript

- Use `const` by default, `let` only when reassignment needed
- Use arrow functions for callbacks
- Use template literals for string interpolation
- Prefer `addEventListener` over inline event handlers
- Use `[...querySelectorAll(...)]` to convert NodeLists to arrays

### Go (TUI)

- Use `github.com/charmbracelet/x/ansi` for string width measurement, **not** `go-runewidth`
- Always use `realStringWidth()` (strips VS16/VS15 variation selectors) for emoji width — `ansi.StringWidth` miscounts emoji with variation selectors
- Use `padIcon()` to normalize icon widths to `iconPadWidth` (2 cells)
- Use `lipgloss.DoubleBorder()` with `muted` foreground for consistent border styling
- Manual border drawing for detail box (avoids lipgloss `.Width()` emoji bugs)
- Settings persisted to `~/.config/homepage/tui.json`

### Python (build.py)

- Use f-strings for string formatting
- 4-space indentation
- Import standard library modules first, then third-party (PyYAML)
- Keep the YAML parser simple - this is a minimal build script

### Naming Conventions

- IDs/Classes: kebab-case (e.g., `groupToggle`, `toggle-switch`)
- JavaScript variables: camelCase (e.g., `draggedItem`, `touchStartY`)
- Data attributes: kebab-case (e.g., `data-name`, `data-groups`)
- Python: snake_case (e.g., `yaml_file`, `output_file`)

### Error Handling

- Use `try/catch` for operations that may fail (e.g., `localStorage` access)
- Always handle promise rejections (e.g., `navigator.clipboard.writeText`)
- Check for null/undefined before accessing properties
- For settings toggles, read current state from DOM/localStorage on click (not captured constants)

## YAML Configuration (services.yml)

```yaml
title: Home Page              # Browser tab title
header: Dashboard             # Header display (optional, defaults to title)

# Define groups (dynamic, displayed in group selector)
groups:
  - id: external
    name: External
    icon: 🌐
  - id: local
    name: Local
    icon: 🏠

# List your services
services:
  - name: ServiceName
    url: https://example.com
    icon: 🛠
    groups: [external]
```

- `title` - Browser tab title (default: "Home Page")
- `header` - Header display text (optional, defaults to title value)
- `groups` - List of group definitions (id, name, icon)
  - `icon` - Emoji for group headers in TUI (optional, defaults to ▸)
  - An implicit "all" group is added automatically (shows all services)
- `services` - List of service entries:
  - `name` (required) - Display name
  - `url` (required) - Service URL
  - `icon` - Emoji or text icon (use clean emoji without variation selectors for best TUI compatibility)
  - `groups` (required) - List of group IDs the service belongs to
  - Services with no groups (or invalid group IDs) default to a "No Group" category

## File Structure

```
/homepage
├── Makefile                   # Build orchestration
├── services.yml               # Service configuration (gitignored)
├── services.example.yml       # Example config for CI demo
├── web/
│   ├── build.py               # Build script (requires PyYAML)
│   └── src/
│       ├── template.html      # HTML template
│       ├── style.css          # Styles
│       ├── script.js          # JavaScript
│       └── health.js          # Health check module
├── cli/
│   ├── tui/
│   │   └── main.go            # Go TUI (bubbletea/lipgloss)
│   ├── go.mod
│   └── go.sum
├── output/                    # Generated files (gitignored except example.html)
│   ├── index.html
│   ├── example.html
│   └── tui
├── .gitignore
├── .github/workflows/
│   └── deploy-example.yml     # CI: builds + deploys example.html
├── AGENTS.md
└── README.md
```

## Important Notes

1. **Do not edit output/index.html directly** - it is generated by `build.py`
2. Edit `services.yml` to add/remove services
3. Edit `template.html`, `style.css`, or `script.js` for web changes
4. Edit `cli/tui/main.go` for TUI changes
5. Run `make build` after any changes to regenerate output
6. Settings that persist in localStorage should use `homepage-` prefix
7. **Emoji in services.yml**: Use clean emoji without VS16/VS15 variation selectors (e.g., 🛰 not 🛰️) for correct TUI alignment

## Common Tasks

### Add a new service
1. Edit `services.yml`, add service entry with name, url, icon, groups
2. Run `make build`
3. Test in browser and/or TUI

### Add a new group
1. Add group entry to `groups` section in `services.yml` (with optional icon)
2. Assign services to the new group via `groups` array
3. Run `make build`
4. Test group selector in browser/TUI

### Add a new TUI setting
1. Add entry to `settingItems` in `cli/tui/main.go`
2. Add rendering case in `renderSettings()`
3. Add toggle logic in `Update()` settings handler
4. Add field to `Settings` struct and `loadSettings()` default
5. Settings auto-save on Enter, auto-revert on Escape

### Modify web styling or JavaScript
1. Edit `style.css` for CSS changes
2. Edit `script.js` for JavaScript changes
3. Edit `template.html` for HTML structure changes
4. Run `make build-html` to regenerate HTML
5. Test in browser

## Features

### Web Features

#### Group Filtering
- Dynamic groups defined in `services.yml`
- Click group indicator (left of footer) to open group selector
- Cycle groups with `/` key
- Services belong to one or more groups via `groups` array
- "All" group shows all services (added automatically)

#### Search & Keyboard Navigation
- Type any key to focus search and filter
- Arrow keys to navigate filtered results
- Page auto-scrolls to keep the selected service in view
- Enter to open selected service
- Escape to clear search
- `/` to cycle group
- `?` to show keyboard shortcuts overlay

#### Drag-and-Drop Reordering
- Desktop: Drag cards to reorder
- Mobile: Touch and drag (distinguished from tap/long-press)
- Order persists in localStorage under key `homepage-order`

#### Long-Press to Copy (Mobile)
- Long-press (500ms) provides haptic feedback (vibration)
- Shows "Release to copy" toast during hold
- Copies URL to clipboard on release
- Requires HTTPS context (notified if unavailable)

#### Grouped View (All Mode)
- In "All" group mode, services are categorized under group headers
- Headers are dynamically hidden if no services match the current search

#### Settings Panel (Web)
- Theme: Toggle light/dark mode (stored in `homepage-theme`)
- Group: Cycle through groups (stored in `homepage-group`)
- Clock: Toggle 12h/24h format (stored in `homepage-clock24`)
- Date: Show/hide date display (stored in `homepage-date`)
- Search: Show/hide search bar (stored in `homepage-search`)
- Icons: Show/hide service icons (stored in `homepage-icons`)
- Compact: Smaller cards mode (stored in `homepage-compact`)
- Clear custom order: Reset service order to default
- Reset all settings: Clear all localStorage and reload

#### Health Checks (Web)
- Auto-runs on page load (500ms delay)
- Press `h` to re-check all services
- Uses `fetch` with `mode: 'no-cors'` (works with local/CORS-blocked services)
- Green dot = online, red dot = offline, gray pulsing = checking
- Summary in header (e.g. "21/23 online")

### CLI TUI Features

#### Service List
- Navigable service list grouped by YAML group (YAML-like style with group headers)
- Cursor skips group rows, wraps around at top/bottom
- Column widths computed from ALL services (stable alignment, no jumping)
- Compact mode toggle (smaller rows)
- Health indicators: green ● online, red ● offline, gray spinner while checking

#### Detail Box
- Shows selected service: icon, name, health status, URL, groups
- Manual double-line borders matching header style
- Uses `realStringWidth()` for correct emoji width measurement

#### Header Box
- 4-line double-border box: title + datetime (line 1), stats (line 2)
- Services count, health summary (e.g. "28✓ 3✗"), spinner during checks
- DateTime auto-updates via tick loop

#### Search
- Press `/` to enter search mode
- Substring matching (case-insensitive) on name, URL, and groups
- Real-time filtering as you type
- Escape to exit search

#### Settings Panel
- `,` to open (centered overlay), `Esc` to cancel (reverts changes), `Enter` to save
- Arrow keys to navigate, left/right to change values
- **Theme**: Cycles through 6 color themes (Indigo, Teal, Rose, Gold, Lime, Sky)
- **DateTime**: Toggle date+time display (default: On)
- **Compact**: Toggle compact mode
- **Install CLI**: Copies binary to `~/.local/bin/<alias>`, ensures `~/.local/bin` is in PATH
  - Customizable alias (default: "hp")
  - Alias editable via text input in settings panel
  - Auto-detects shell (zsh/bash/fish) for PATH setup

#### Keyboard Shortcuts (TUI)
- `↑↓` Navigate service list
- `/` Enter search mode
- `Enter` / `o` Open selected service in browser
- `,` Open settings panel
- `h` Re-check all service health
- `q` Quit

## JavaScript API Reference

### Core Functions
- `updateTime()` - Updates clock and date display
- `saveOrder()` / `loadOrder()` - localStorage persistence for order
- `setGroup(groupId)` - Filter services by group; restores grouped order in 'all' mode
- `showToast(msg)` - Show toast notification
- `copyLink(url)` - Copy URL to clipboard with safety checks
- `filterServices(query)` - Filter services by name and toggle group headers
- `updateSelection(visibleLinks)` - Highlight selected service and scroll into view
- `openSelected()` - Navigate to selected service

### Settings Functions
- `setTheme(light)` - Toggle light/dark theme
- `setClockFormat(is24)` - Toggle 12h/24h clock
- `setDateVisible(visible)` - Toggle date display
- `setSearchVisible(visible)` - Toggle search bar
- `setIconsVisible(visible)` - Toggle service icons
- `setCompact(compact)` - Toggle compact mode

### Health Check Functions
- `checkAllHealth()` - Check all service URLs (async, uses `no-cors` fetch)
- `renderHealthIndicators()` - Update health dots on service cards
- `updateHealthSummary()` - Update header summary (e.g. "21/23 online")

### Global State
- `groupToggle`, `groupIndicator`, `groupSelector`, `groupOptions`
- `GROUPS` - Array of group objects from `window.GROUPS`
- `currentGroup` - Currently selected group ID
- `grid`, `links`, `draggedItem`, `selectedIndex`
- `settingsBtn`, `settingsDropdown`, `settingsFocusIndex`

### localStorage Keys
- `homepage-order` - Service order
- `homepage-group` - Selected group ID (e.g., "external", "local", "all")
- `homepage-theme` - Theme preference (light/dark)
- `homepage-clock24` - Clock format (true/false)
- `homepage-date` - Date visibility (true/false)
- `homepage-search` - Search visibility (true/false)
- `homepage-icons` - Icons visibility (true/false)
- `homepage-compact` - Compact mode (true/false)

### TUI Settings File
- `~/.config/homepage/tui.json` - TUI persistent settings
- Fields: `theme` (int), `compact` (bool), `datetime` (bool), `alias` (string)

## Browser Compatibility

- ES6+ features, localStorage, Clipboard API
- CSS custom properties (not IE11)
- Target: Modern browsers (Chrome, Firefox, Safari, Edge)

## Go Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - TUI styling
- `github.com/charmbracelet/bubbles` - TUI components (textinput)
- `github.com/charmbracelet/x/ansi` - ANSI string width measurement
- `gopkg.in/yaml.v3` - YAML parsing
