# Homepage

A minimal, keyboard-friendly homepage dashboard.

## Features

- Clean, dark-themed UI with light mode option
- Dynamic group filtering (External/Local/All or custom groups)
- **Categorized View**: Services are grouped with headers in "All" mode
- **Intelligent Search**: Type-to-jump with auto-scroll to selection
- **Drag-and-Drop**: Reorder services (persists in browser)
- **Mobile Friendly**: Long-press (500ms) with haptic feedback to copy URL
- **Customizable**: Extensive settings panel for theme, clock, and layout

## Quick Start

```bash
# Edit services
vim services.yml

# Build
python3 build.py

# Generate example
python3 build.py --example
```

## Configuration

Edit `services.yml`:

```yaml
title: My Homepage     # Browser tab title
header: Dashboard       # Header display (optional, defaults to title)

# Define groups
groups:
  - id: external
    name: External
  - id: local
    name: Local

# List your services
services:
  - name: ServiceName
    url: https://service.example.com
    icon: 🛠
    groups: [external]

  - name: LocalService
    url: http://localhost:8080
    icon: 🏠
    groups: [local]
```

### Service Configuration

- **url** (required): Service URL
- **icon**: Emoji or text icon
- **groups**: List of group IDs. Services without groups fall into "No Group"
- Services with no groups default to appearing in "All" group only

### Groups

Groups are defined in the `groups` section with:
- **id**: Unique identifier (used in service `groups` array)
- **name**: Display name (shown in group selector)

An implicit "All" group is automatically added, showing all services categorized by their primary group.

### Settings

Click the ⚙ button or press `,` to access:
- **Theme**: Toggle light/dark mode
- **Group**: Cycle through groups
- **Clock**: Switch between 12h/24h format
- **Date**: Show/hide date display
- **Search**: Show/hide search bar
- **Icons**: Show/hide service icons
- **Compact**: Smaller cards for more services
- **Clear custom order**: Reset service order to default
- **Reset all settings**: Clear all saved preferences

### Keyboard Shortcuts

- **Type**: Filter services and jump to first result (auto-scrolls to view)
- **Arrow Keys**: Navigate filtered results (or settings when open)
- **Enter**: Open selected service (or activate settings option)
- **Esc**: Clear search / Close settings / Close hints
- **,**: Toggle settings panel
- **/**: Cycle group
- **?**: Show keyboard shortcuts overlay

### Mobile

- **Tap**: Navigate to service
- **Long Press** (500ms): Copy URL to clipboard (includes haptic feedback)
- **Drag**: Reorder services

## Deployment

```bash
python3 build.py
# Upload index.html to your web server
```

## Files

```
.
├── src/
│   ├── template.html    # HTML template
│   ├── style.css        # Styles
│   └── script.js        # JavaScript
├── services.yml         # Service definitions
├── services.example.yml # Example config
├── build.py             # Build script (requires PyYAML)
├── index.html           # Generated output (do not edit)
├── AGENTS.md            # Agent guidelines
└── README.md
```
