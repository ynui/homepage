# Homepage

A minimal, keyboard-friendly homepage dashboard.

## Features

- Clean, dark-themed UI with light mode option
- Dynamic group filtering (External/Local/All or custom groups)
- Search/filter services with type-to-jump
- Drag to reorder services (persists in browser)
- Long-press (500ms) to copy service URL on mobile
- Customizable via settings panel

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
- **groups** (required): List of group IDs the service belongs to
- Services with no groups default to appearing in "All" group only

### Groups

Groups are defined in the `groups` section with:
- **id**: Unique identifier (used in service `groups` array)
- **name**: Display name (shown in group selector)

An implicit "All" group is automatically added (shows all services).

### Settings

Click the ⚙ button or press `,` to access:
- **Theme**: Toggle light/dark mode
- **Group**: Cycle through groups (External → Local → All)
- **Clock**: Switch between 12h/24h format
- **Date**: Show/hide date display
- **Search**: Show/hide search bar
- **Icons**: Show/hide service icons
- **Compact**: Smaller cards for more services
- **Clear custom order**: Reset service order to default
- **Reset all settings**: Clear all saved preferences

### Keyboard Shortcuts

- **Type**: Filter services and jump to first result
- **Arrow Keys**: Navigate filtered results (or settings when open)
- **Enter**: Open selected service (or activate settings option)
- **Esc**: Clear search / Close settings / Close hints
- **,**: Toggle settings panel
- **/**: Cycle group (External → Local → All)
- **?**: Show keyboard shortcuts overlay

### Mobile

- **Tap**: Navigate to service
- **Long Press** (500ms): Copy URL to clipboard
- **Drag**: Reorder services

## Deployment

```bash
python3 build.py
# Upload index.html to your web server

# For example config
python3 build.py --example
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

## Configuration

Edit `services.yml`:

```yaml
title: My Homepage     # Browser tab title
header: Dashboard       # Header display (optional, defaults to title)

services:
  - name: ServiceName
    url: https://service.example.com    # external URL
    local: http://localhost:8080        # local URL
    icon: 🛠
```

### Service Types

- **url + local**: Shows in all modes, switches URL based on mode
- **url only**: Shows only in external mode
- **local only**: Shows only in local mode

### Settings

Click the ⚙ button or press `,` to access:
- **Theme**: Toggle light/dark mode
- **Mode**: Cycle External → Local → All
- **Clock**: Switch between 12h/24h format
- **Date**: Show/hide date display
- **Search**: Show/hide search bar
- **Icons**: Show/hide service icons
- **Compact**: Smaller cards for more services
- **Clear custom order**: Reset service order to default
- **Reset all settings**: Clear all saved preferences

### Keyboard Shortcuts

- **Type**: Filter services and jump to first result
- **Arrow Keys**: Navigate filtered results (or settings when open)
- **Enter**: Open selected service (or activate settings option)
- **Esc**: Clear search / Close settings / Close hints
- **,**: Toggle settings panel
- **/**: Cycle mode (External → Local → All)
- **?**: Show keyboard shortcuts overlay

### Mobile

- **Tap**: Navigate to service
- **Long Press** (500ms): Copy URL to clipboard
- **Drag**: Reorder services

## Deployment

```bash
python3 build.py
# Upload index.html to your web server

# For example config
python3 build.py --example
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
├── build.py             # Build script
├── index.html           # Generated output (do not edit)
├── AGENTS.md            # Agent guidelines
└── README.md
```
