# Homepage

A minimal, keyboard-friendly homepage dashboard.

## Features

- Clean, dark-themed UI
- External/Local toggle for different network environments
- Search/filter services with type-to-jump
- Drag to reorder services (persists in browser)
- Long-press (500ms) to copy service URL on mobile

## Quick Start

```bash
# Edit services
vim services.yml

# Build
python3 build.py
```

## Configuration

Edit `services.yml`:

```yaml
title: My Homepage

services:
  - name: ServiceName
    url: https://service.example.com    # external URL
    local: http://localhost:8080        # local URL (optional)
    icon: 🛠
```

### Service Types

- **url + local**: Shows in both modes, switches based on toggle
- **url only**: Shows only in external mode
- **local only**: Shows only in local mode

### Keyboard Shortcuts

- **Type**: Filter services and jump to first result
- **Arrow Keys**: Navigate filtered results
- **Enter**: Open selected service
- **Esc**: Clear search

### Mobile

- **Tap**: Navigate to service
- **Long Press** (500ms): Copy URL to clipboard
- **Drag**: Reorder services

## Deployment

```bash
python3 build.py
# Upload index.html to your web server
```

## Files

```
.
├── services.yml         # Service definitions
├── services.example.yml # Example config
├── build.py             # Build script
├── index.html           # Generated output (do not edit)
└── README.md
```
