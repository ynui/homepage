# Homelab UI

A minimal, clean homepage for homelab services.

## Features

- Clean, dark-themed UI
- External/Local toggle for different network environments
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
title: My Homelab     # Page title (optional, default: Homelab)

services:
  - name: Radarr
    url: https://radarr.example.com       # external URL
    local: http://192.168.1.100:7878     # local URL (optional)
    icon: 🎬                              # emoji icon
```

### Service Types

- **Both url + local**: Shows in both modes, switches URL based on toggle
- **url only**: Shows only in external mode
- **local only**: Shows only in local mode

### Interaction

- **Click/Tap**: Navigate to service
- **Long Press** (500ms): Copy URL to clipboard
- **Scroll**: Normal page scrolling (won't trigger copy)

## Deployment

1. Run `python3 build.py`
2. Upload `index.html` to your web server (nginx, apache, etc.)

## Files

```
.
├── services.yml    # Service definitions
├── build.py       # Build script
├── index.html     # Generated output
└── README.md      # This file
```
