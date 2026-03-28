# Homepage

A minimal, clean homepage.

## Features

- Clean, dark-themed UI
- External/Local toggle for different network environments
- Search/filter services
- Type-to-jump (just start typing)
- Drag to reorder services (saved in browser)
- Long-press (500ms) to copy service URL
- Auto-detect service type from YAML
- Mobile-friendly with scroll detection

## Quick Start

```bash
# Edit services
vim services.yml

# Build
python3 build.py

# Serve locally
python3 -m http.server 8080
```

## Configuration

Edit `services.yml`:

```yaml
title: My Homepage     # Page title (optional, default: Home Page)

services:
  - name: service
    url: https://service.example.com       # external URL
    local: http://localhost:8080           # local URL (optional)
    icon: 🛠                               # emoji icon
```

### Service Types

- **Both url + local**: Shows in both modes, switches URL based on toggle
- **url only**: Shows only in external mode
- **local only**: Shows only in local mode

### Interaction

- **Click/Tap**: Navigate to service
- **Long Press** (500ms): Copy URL to clipboard
- **Scroll**: Normal page scrolling (won't trigger copy)
- **Type**: Start typing to filter and jump to first result
- **Arrow Keys**: Navigate filtered results
- **Enter**: Open selected service
- **Esc**: Clear search

## Deployment

1. Run `python3 build.py`
2. Upload `index.html` to your web server (nginx, apache, etc.)

## Files

```
.
├── services.yml    # Service definitions
├── build.py        # Build script
├── index.html      # Generated output
└── README.md       # This file
```
