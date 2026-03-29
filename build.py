#!/usr/bin/env python3
import sys
import os

example = None
if len(sys.argv) > 1 and sys.argv[1].startswith('--'):
    example = sys.argv[1][2:]

yaml_file = f'services.{example}.yml' if example else 'services.yml'
output_file = f'{example}.html' if example else 'index.html'

with open(yaml_file) as f:
    content = f.read()

title = 'Home Page'
services = []

for line in content.split('\n'):
    if line.startswith('title:'):
        title = line.split('title:')[1].strip()

current = {}
for line in content.split('\n'):
    if line.startswith('  - name:'):
        if current:
            services.append(current)
        current = {'name': line.split('name:')[1].strip()}
    elif line.startswith('    url:'):
        current['url'] = line.split('url:')[1].strip()
    elif line.startswith('    local:'):
        current['local'] = line.split('local:')[1].strip()
    elif line.startswith('    icon:'):
        current['icon'] = line.split('icon:')[1].strip()
if current:
    services.append(current)

for s in services:
    has_url = 'url' in s and s['url']
    has_local = 'local' in s and s['local']
    if has_url and has_local:
        s['type'] = 'both'
    elif has_local:
        s['type'] = 'local'
    else:
        s['type'] = 'external'

html = f'''<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{title}</title>
  <link rel="icon" href="data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22><text y=%22.9em%22 font-size=%2290%22>🏠</text></svg>">
  <style>
    * {{
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }}

    :root {{
      --bg: #0a0a0b;
      --card-bg: #131316;
      --card-border: #1f1f24;
      --text: #7a7a85;
      --text-hover: #e8e8e8;
      --accent: #6366f1;
      --accent-glow: rgba(99, 102, 241, 0.15);
    }}

    :root.light {{
      --bg: #f5f5f5;
      --card-bg: #ffffff;
      --card-border: #e0e0e0;
      --text: #666666;
      --text-hover: #1a1a1a;
      --accent: #6366f1;
      --accent-glow: rgba(99, 102, 241, 0.1);
    }}

    body {{
      font-family: 'SF Mono', 'Fira Code', 'JetBrains Mono', monospace;
      background: var(--bg);
      color: var(--text);
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 3rem 2rem;
    }}

    main {{
      max-width: 720px;
      width: 100%;
    }}

    header {{
      position: relative;
      text-align: center;
      margin-bottom: 3rem;
    }}

    h1 {{
      font-size: 0.75rem;
      font-weight: 500;
      letter-spacing: 0.4em;
      text-transform: uppercase;
      color: #3d3d45;
    }}

    .time {{
      font-size: 2.5rem;
      font-weight: 300;
      color: var(--text-hover);
      letter-spacing: 0.05em;
    }}

    .date {{
      font-size: 0.7rem;
      color: var(--text);
      margin-top: 0.25rem;
      text-transform: uppercase;
      letter-spacing: 0.1em;
    }}

    .date.hidden {{
      display: none;
    }}

    .toggle {{
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 0.5rem;
      margin-top: 1rem;
      cursor: pointer;
      user-select: none;
    }}

    .toggle-input {{
      display: none;
    }}

    .toggle-switch {{
      width: 44px;
      height: 24px;
      background: #2a2a30;
      border-radius: 12px;
      position: relative;
      transition: background 0.3s;
      border: 1px solid #333;
      cursor: pointer;
    }}

    .toggle-switch::after {{
      content: '';
      position: absolute;
      width: 18px;
      height: 18px;
      background: #666;
      border-radius: 50%;
      top: 2px;
      left: 2px;
      transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1), background 0.3s;
    }}

    .toggle-input:checked + .toggle-switch {{
      background: var(--accent);
      border-color: var(--accent);
    }}

    .toggle-input:checked + .toggle-switch::after {{
      transform: translateX(20px);
      background: #fff;
    }}

    .toggle-label {{
      font-size: 0.65rem;
      text-transform: uppercase;
      letter-spacing: 0.1em;
      color: #444;
      margin-top: 0.25rem;
    }}

    .search {{
      margin-top: 1.5rem;
      width: 100%;
      max-width: 280px;
      padding: 0.75rem 1rem;
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 8px;
      color: var(--text);
      font-family: inherit;
      font-size: 0.8rem;
      text-align: center;
      outline: none;
      transition: border-color 0.2s;
    }}

    .search::placeholder {{
      color: #444;
    }}

    .search:focus {{
      border-color: var(--accent);
    }}

    .settings {{
      position: absolute;
      top: 1rem;
      right: 1rem;
    }}

    .settings-btn {{
      background: none;
      border: none;
      color: var(--text);
      cursor: pointer;
      padding: 0.5rem;
      font-size: 1rem;
      opacity: 0.5;
      transition: opacity 0.2s;
    }}

    button:focus, .settings-dropdown:focus {{
      outline: none;
    }}

    .settings-btn:hover {{
      opacity: 1;
    }}

    .settings-dropdown {{
      position: absolute;
      top: 100%;
      right: 0;
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 8px;
      padding: 0.5rem;
      min-width: 160px;
      display: none;
      z-index: 100;
    }}

    .settings-dropdown.show {{
      display: block;
    }}

    .settings-option {{
      display: block;
      width: 100%;
      background: none;
      border: none;
      color: var(--text);
      padding: 0.5rem 0.75rem;
      text-align: left;
      font-family: inherit;
      font-size: 0.75rem;
      cursor: pointer;
      border-radius: 4px;
    }}

    .settings-option:hover {{
      background: var(--card-border);
      color: var(--text-hover);
    }}

    .settings-option.focused {{
      background: var(--card-border);
      color: var(--accent);
    }}

    a.hidden {{
      display: none;
    }}

    .grid {{
      display: flex;
      flex-wrap: wrap;
      gap: 0.75rem;
      justify-content: center;
    }}

    .grid a {{
      width: 140px;
      flex-shrink: 0;
      cursor: grab;
    }}

    a {{
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 0.75rem;
      padding: 1.5rem 1rem;
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 12px;
      text-decoration: none;
      color: var(--text);
      font-size: 0.7rem;
      text-transform: uppercase;
      letter-spacing: 0.15em;
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }}

    a.hidden {{
      display: none;
    }}

    a:hover {{
      transform: translateY(-2px);
      border-color: var(--accent);
      box-shadow: 0 8px 30px var(--accent-glow), 0 0 0 1px var(--accent);
      color: var(--text-hover);
    }}

    a.selected {{
      transform: translateY(-2px);
      border-color: var(--accent);
      box-shadow: 0 8px 30px var(--accent-glow), 0 0 0 1px var(--accent);
      color: var(--text-hover);
    }}

    .icon {{
      font-size: 1.5rem;
      opacity: 0.7;
      transition: opacity 0.3s;
    }}

    a:hover .icon {{
      opacity: 1;
    }}

    .search.hidden {{
      display: none;
    }}

    .icon.hidden {{
      display: none;
    }}

    .compact .grid a {{
      width: 80px;
      padding: 0.75rem 0.5rem;
    }}

    .compact .icon {{
      font-size: 1rem;
    }}

    .compact a {{
      padding: 0.75rem 0.5rem;
      gap: 0.25rem;
    }}

    .toast {{
      position: fixed;
      bottom: 2rem;
      left: 50%;
      transform: translateX(-50%) translateY(100px);
      background: var(--card-bg);
      border: 1px solid var(--accent);
      padding: 0.75rem 1.5rem;
      border-radius: 8px;
      font-size: 0.75rem;
      color: var(--text-hover);
      opacity: 0;
      transition: all 0.3s ease;
      pointer-events: none;
    }}

    .toast.show {{
      transform: translateX(-50%) translateY(0);
      opacity: 1;
    }}

    .hints {{
      display: none;
      position: fixed;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      background: var(--bg);
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 1.5rem 2rem;
      z-index: 1000;
      box-shadow: 0 4px 24px rgba(0, 0, 0, 0.4);
    }}

    .hints.show {{
      display: block;
    }}

    .hints h2 {{
      margin: 0 0 1rem;
      font-size: 1rem;
      color: var(--accent);
    }}

    .hint-row {{
      display: flex;
      justify-content: space-between;
      gap: 2rem;
      padding: 0.25rem 0;
      color: var(--text);
      font-size: 0.875rem;
    }}

    .hint-row .key {{
      background: var(--border);
      padding: 0.125rem 0.5rem;
      border-radius: 4px;
      font-family: monospace;
      color: var(--accent);
    }}
  </style>
</head>
<body>
  <main>
    <div class="settings">
      <button id="settingsBtn" class="settings-btn">⚙</button>
      <div id="settingsDropdown" class="settings-dropdown" tabindex="0">
        <button id="themeToggle" class="settings-option">Theme: 🌙</button>
        <button id="clockToggle" class="settings-option">Clock: 24h</button>
        <button id="dateToggle" class="settings-option">Date: On</button>
        <button id="searchToggle" class="settings-option">Search: On</button>
        <button id="iconToggle" class="settings-option">Icons: On</button>
        <button id="compactToggle" class="settings-option">Compact: Off</button>
        <button id="clearOrder" class="settings-option">Clear custom order</button>
        <button id="resetAll" class="settings-option">Reset all settings</button>
      </div>
    </div>
    <header>
      <h1>{title}</h1>
      <div class="time" id="time"></div>
      <div class="date" id="date"></div>
      <label class="toggle">
        <input type="checkbox" id="modeToggle" class="toggle-input">
        <span class="toggle-switch"></span>
        <span id="modeLabel" class="toggle-label">External</span>
      </label>
      <input type="text" id="search" class="search" placeholder="Type to jump...">
    </header>
    <div class="grid">
'''

for s in services:
    url = s.get('url', '')
    local = s.get('local', url)
    html += f'''      <a href="{url or local}" data-name="{s['name'].lower()}" data-external="{url}" data-local="{local}" data-type="{s['type']}">
        <span class="icon">{s['icon']}</span>
        {s['name']}
      </a>
'''

html += '''    </div>
  </main>
  <div class="toast" id="toast">Copied!</div>
    <div class="hints" id="hints">
    <h2>Keyboard Shortcuts</h2>
    <div class="hint-row"><span class="key">↑↓←→</span><span>Navigate apps</span></div>
    <div class="hint-row"><span class="key">Enter</span><span>Open selected</span></div>
    <div class="hint-row"><span class="key">/</span><span>Toggle Local/External</span></div>
    <div class="hint-row"><span class="key">,</span><span>Open settings</span></div>
    <div class="hint-row"><span class="key">Type</span><span>Search apps</span></div>
    <div class="hint-row"><span class="key">Esc</span><span>Clear / Close</span></div>
    <div class="hint-row"><span class="key">?</span><span>Show shortcuts</span></div>
  </div>
  <script>
    function updateTime() {
      const now = new Date();
      const is24h = localStorage.getItem('homepage-clock24') !== 'false';
      document.getElementById('time').textContent = now.toLocaleTimeString('en-US', { 
        hour12: !is24h, 
        hour: '2-digit', 
        minute: '2-digit' 
      });
      const dateEl = document.getElementById('date');
      if (dateEl) {
        dateEl.textContent = now.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
      }
    }
    updateTime();
    setInterval(updateTime, 1000);

    const toggle = document.getElementById('modeToggle');
    const grid = document.querySelector('.grid');
    const links = [...document.querySelectorAll('.grid a')];

    let draggedItem = null;

    function saveOrder() {
      const order = [...grid.children].map(el => el.dataset.name);
      localStorage.setItem('homepage-order', JSON.stringify(order));
    }

    function loadOrder() {
      const saved = localStorage.getItem('homepage-order');
      if (saved) {
        const order = JSON.parse(saved);
        order.forEach(name => {
          const link = links.find(l => l.dataset.name === name);
          if (link) grid.appendChild(link);
        });
      }
    }
    loadOrder();

    links.forEach(link => {
      link.draggable = true;
      link.addEventListener('dragstart', (e) => {
        draggedItem = link;
        link.style.opacity = '0.5';
      });
      link.addEventListener('dragend', () => {
        link.style.opacity = '1';
        draggedItem = null;
        saveOrder();
      });
      link.addEventListener('dragover', (e) => {
        e.preventDefault();
        if (draggedItem && draggedItem !== link) {
          const rect = link.getBoundingClientRect();
          const midY = rect.top + rect.height / 2;
          if (e.clientY < midY) {
            grid.insertBefore(draggedItem, link);
          } else {
            grid.insertBefore(draggedItem, link.nextSibling);
          }
        }
      });
    });

    function setMode(local) {
      links.forEach(link => {
        const type = link.dataset.type;
        if (type === 'both') {
          link.href = local ? link.dataset.local : link.dataset.external;
          link.classList.remove('hidden');
        } else if (type === 'local') {
          link.href = link.dataset.local;
          link.classList.toggle('hidden', !local);
        } else if (type === 'external') {
          link.href = link.dataset.external;
          link.classList.toggle('hidden', local);
        }
      });
    }

    const modeLabel = document.getElementById('modeLabel');
    toggle.addEventListener('change', (e) => {
      modeLabel.textContent = e.target.checked ? 'Local' : 'External';
      setMode(e.target.checked);
    });

    const toast = document.getElementById('toast');
    let longPressTimer;

    function showToast(msg) {
      toast.textContent = msg;
      toast.classList.add('show');
      setTimeout(() => toast.classList.remove('show'), 1500);
    }

    function copyLink(url) {
      navigator.clipboard.writeText(url).then(() => {
        showToast('Link copied!');
      });
    }

    links.forEach(link => {
      let longPressed = false;
      let touchMoved = false;
      let startX, startY;

      const startPress = (e) => {
        longPressed = false;
        touchMoved = false;
        if (e.type === 'touchstart') {
          startX = e.touches[0].clientX;
          startY = e.touches[0].clientY;
        }
        longPressTimer = setTimeout(() => {
          if (!touchMoved) {
            longPressed = true;
            copyLink(link.href);
          }
        }, 500);
      };

      const movePress = (e) => {
        if (e.type === 'touchmove') {
          const dx = Math.abs(e.touches[0].clientX - startX);
          const dy = Math.abs(e.touches[0].clientY - startY);
          if (dx > 10 || dy > 10) {
            touchMoved = true;
            clearTimeout(longPressTimer);
          }
        }
      };

      const endPress = () => {
        touchMoved = false;
        clearTimeout(longPressTimer);
      };

      const handleClick = (e) => {
        if (longPressed) {
          e.preventDefault();
          e.stopPropagation();
          longPressed = false;
        }
      };

      const handleContextMenu = (e) => {
        e.preventDefault();
      };

      link.addEventListener('mousedown', startPress);
      link.addEventListener('mouseup', endPress);
      link.addEventListener('mouseleave', () => clearTimeout(longPressTimer));
      link.addEventListener('click', handleClick);
      link.addEventListener('touchstart', startPress, { passive: true });
      link.addEventListener('touchmove', movePress, { passive: true });
      link.addEventListener('touchend', endPress);
      link.addEventListener('contextmenu', handleContextMenu);
    });

    setMode(false);

    const settingsBtn = document.getElementById('settingsBtn');
    const settingsDropdown = document.getElementById('settingsDropdown');
    const clearOrder = document.getElementById('clearOrder');
    const resetAll = document.getElementById('resetAll');

    settingsBtn.addEventListener('click', (e) => {
      e.stopPropagation();
      settingsDropdown.classList.toggle('show');
      if (settingsDropdown.classList.contains('show')) {
        settingsFocusIndex = 0;
        settingsOptions.forEach((opt, i) => opt.classList.toggle('focused', i === 0));
        settingsDropdown.focus();
      }
    });

    let settingsFocusIndex = -1;
    const settingsOptions = [...document.querySelectorAll('.settings-option')];

    settingsDropdown.addEventListener('keydown', (e) => {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        settingsFocusIndex = Math.min(settingsFocusIndex + 1, settingsOptions.length - 1);
        settingsOptions.forEach((opt, i) => opt.classList.toggle('focused', i === settingsFocusIndex));
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        settingsFocusIndex = Math.max(settingsFocusIndex - 1, 0);
        settingsOptions.forEach((opt, i) => opt.classList.toggle('focused', i === settingsFocusIndex));
      } else if (e.key === 'Enter') {
        e.preventDefault();
        if (settingsFocusIndex >= 0) settingsOptions[settingsFocusIndex].click();
      } else if (e.key === 'Escape') {
        settingsDropdown.classList.remove('show');
      }
    });

    settingsOptions.forEach((opt, i) => {
      opt.addEventListener('mouseenter', () => {
        settingsFocusIndex = i;
        settingsOptions.forEach((o, j) => o.classList.toggle('focused', j === i));
      });
    });

    document.addEventListener('click', (e) => {
      if (!settingsDropdown.contains(e.target) && e.target !== settingsBtn) {
        settingsDropdown.classList.remove('show');
      }
    });

    clearOrder.addEventListener('click', () => {
      localStorage.removeItem('homepage-order');
      links.forEach(link => grid.appendChild(link));
      showToast('Order reset');
    });

    resetAll.addEventListener('click', () => {
      localStorage.clear();
      location.reload();
    });

    const themeToggle = document.getElementById('themeToggle');
    const savedTheme = localStorage.getItem('homepage-theme');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const isLight = savedTheme ? savedTheme === 'light' : !prefersDark;

    function setTheme(light) {
      document.documentElement.classList.toggle('light', light);
      themeToggle.textContent = `Theme: ${light ? '☀️' : '🌙'}`;
      localStorage.setItem('homepage-theme', light ? 'light' : 'dark');
    }

    setTheme(isLight);

    themeToggle.addEventListener('click', () => {
      const isCurrentlyLight = document.documentElement.classList.contains('light');
      setTheme(!isCurrentlyLight);
      showToast(`Theme: ${!isCurrentlyLight ? 'Light' : 'Dark'}`);
    });

    const search = document.getElementById('search');

    const clockToggle = document.getElementById('clockToggle');

    function setClockFormat(is24) {
      clockToggle.textContent = `Clock: ${is24 ? '24h' : '12h'}`;
      localStorage.setItem('homepage-clock24', is24 ? 'true' : 'false');
      updateTime();
    }

    setClockFormat(localStorage.getItem('homepage-clock24') !== 'false');

    clockToggle.addEventListener('click', () => {
      const is24 = localStorage.getItem('homepage-clock24') !== 'false';
      setClockFormat(!is24);
      showToast(`Clock: ${!is24 ? '24h' : '12h'}`);
    });

    const dateToggle = document.getElementById('dateToggle');

    function setDateVisible(visible) {
      document.getElementById('date').style.display = visible ? 'block' : 'none';
      dateToggle.textContent = `Date: ${visible ? 'On' : 'Off'}`;
      localStorage.setItem('homepage-date', visible ? 'true' : 'false');
    }

    setDateVisible(localStorage.getItem('homepage-date') !== 'false');

    dateToggle.addEventListener('click', () => {
      const isVisible = document.getElementById('date').style.display !== 'none';
      setDateVisible(!isVisible);
      showToast(`Date: ${!isVisible ? 'On' : 'Off'}`);
    });

    const searchToggle = document.getElementById('searchToggle');

    function setSearchVisible(visible) {
      search.classList.toggle('hidden', !visible);
      searchToggle.textContent = `Search: ${visible ? 'On' : 'Off'}`;
      localStorage.setItem('homepage-search', visible ? 'true' : 'false');
    }

    setSearchVisible(localStorage.getItem('homepage-search') !== 'false');

    searchToggle.addEventListener('click', () => {
      const isVisible = search.classList.contains('hidden') === false;
      setSearchVisible(!isVisible);
      showToast(`Search: ${!isVisible ? 'On' : 'Off'}`);
    });

    const iconToggle = document.getElementById('iconToggle');

    function setIconsVisible(visible) {
      document.querySelectorAll('.icon').forEach(icon => icon.classList.toggle('hidden', !visible));
      iconToggle.textContent = `Icons: ${visible ? 'On' : 'Off'}`;
      localStorage.setItem('homepage-icons', visible ? 'true' : 'false');
    }

    setIconsVisible(localStorage.getItem('homepage-icons') !== 'false');

    iconToggle.addEventListener('click', () => {
      const isVisible = document.querySelector('.icon.hidden') === null;
      setIconsVisible(!isVisible);
      showToast(`Icons: ${!isVisible ? 'On' : 'Off'}`);
    });

    const compactToggle = document.getElementById('compactToggle');

    function setCompact(compact) {
      document.body.classList.toggle('compact', compact);
      compactToggle.textContent = `Compact: ${compact ? 'On' : 'Off'}`;
      localStorage.setItem('homepage-compact', compact ? 'true' : 'false');
    }

    setCompact(localStorage.getItem('homepage-compact') === 'true');

    compactToggle.addEventListener('click', () => {
      const isCompact = document.body.classList.contains('compact');
      setCompact(!isCompact);
      showToast(`Compact: ${!isCompact ? 'On' : 'Off'}`);
    });

    let selectedIndex = -1;

    function filterServices(query) {
      const q = query.toLowerCase();
      const isLocal = toggle.checked;
      let visibleLinks = [];
      links.forEach(link => {
        const type = link.dataset.type;
        let isTypeMatch = true;
        if (type === 'local') isTypeMatch = isLocal;
        else if (type === 'external') isTypeMatch = !isLocal;
        const name = link.textContent.toLowerCase();
        const isMatch = isTypeMatch && (!q || name.includes(q));
        link.classList.toggle('hidden', !isMatch);
        if (isMatch) visibleLinks.push(link);
      });
      selectedIndex = visibleLinks.length > 0 ? 0 : -1;
      updateSelection(visibleLinks);
    }

    function updateSelection(visibleLinks) {
      visibleLinks.forEach((link, i) => link.classList.remove('selected'));
      if (selectedIndex >= 0 && visibleLinks[selectedIndex]) {
        visibleLinks[selectedIndex].classList.add('selected');
      }
    }

    function openSelected() {
      const visibleLinks = [...links].filter(l => !l.classList.contains('hidden'));
      if (selectedIndex >= 0 && visibleLinks[selectedIndex]) {
        window.location.href = visibleLinks[selectedIndex].href;
      }
    }

    search.addEventListener('input', (e) => filterServices(e.target.value));

    search.addEventListener('keydown', (e) => {
      if (e.key === '/') {
        e.preventDefault();
        return;
      }
      if (e.key === ',') {
        e.preventDefault();
        search.blur();
        settingsDropdown.classList.toggle('show');
        if (settingsDropdown.classList.contains('show')) {
          settingsFocusIndex = -1;
          settingsDropdown.focus();
        }
        return;
      }
      const visibleLinks = [...links].filter(l => !l.classList.contains('hidden'));
      if (visibleLinks.length === 0) return;
      const gridWidth = grid.clientWidth - 24;
      const itemWidth = 147;
      const cols = Math.floor(gridWidth / itemWidth) || 1;
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        selectedIndex = Math.min(selectedIndex + cols, visibleLinks.length - 1);
        updateSelection(visibleLinks);
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        selectedIndex = Math.max(selectedIndex - cols, 0);
        updateSelection(visibleLinks);
      } else if (e.key === 'ArrowRight') {
        e.preventDefault();
        selectedIndex = Math.min(selectedIndex + 1, visibleLinks.length - 1);
        updateSelection(visibleLinks);
      } else if (e.key === 'ArrowLeft') {
        e.preventDefault();
        selectedIndex = Math.max(selectedIndex - 1, 0);
        updateSelection(visibleLinks);
      } else if (e.key === 'Enter') {
        e.preventDefault();
        openSelected();
      } else if (e.key === 'Escape') {
        search.value = '';
        filterServices('');
        search.blur();
      }
    });

    document.addEventListener('keydown', (e) => {
      const hints = document.getElementById('hints');
      if (e.key === '?' || (e.key === '/' && e.shiftKey)) {
        e.preventDefault();
        hints.classList.toggle('show');
        return;
      }
      if (hints.classList.contains('show')) {
        hints.classList.remove('show');
      }
      if (settingsDropdown.classList.contains('show')) {
        return;
      }
      if (e.target.tagName === 'INPUT') {
        if (e.key === '/') {
          e.preventDefault();
          toggle.checked = !toggle.checked;
          modeLabel.textContent = toggle.checked ? 'Local' : 'External';
          setMode(toggle.checked);
          filterServices(search.value);
          return;
        }
        if (e.key === ',') {
          e.preventDefault();
          search.blur();
          settingsDropdown.classList.toggle('show');
          if (settingsDropdown.classList.contains('show')) {
            settingsFocusIndex = 0;
            settingsOptions.forEach((opt, i) => opt.classList.toggle('focused', i === 0));
            settingsDropdown.focus();
          }
          return;
        }
        return;
      }
      const visibleLinks = [...links].filter(l => !l.classList.contains('hidden'));
      if (visibleLinks.length === 0) return;
      const gridWidth = grid.clientWidth - 24;
      const itemWidth = 147;
      const cols = Math.floor(gridWidth / itemWidth) || 1;
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        selectedIndex = Math.min(selectedIndex + cols, visibleLinks.length - 1);
        updateSelection(visibleLinks);
        return;
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        selectedIndex = Math.max(selectedIndex - cols, 0);
        updateSelection(visibleLinks);
        return;
      } else if (e.key === 'ArrowRight') {
        e.preventDefault();
        selectedIndex = Math.min(selectedIndex + 1, visibleLinks.length - 1);
        updateSelection(visibleLinks);
        return;
      } else if (e.key === 'ArrowLeft') {
        e.preventDefault();
        selectedIndex = Math.max(selectedIndex - 1, 0);
        updateSelection(visibleLinks);
        return;
      } else if (e.key === 'Enter' && selectedIndex >= 0) {
        e.preventDefault();
        openSelected();
        return;
      }
      if (e.key === 'Escape') {
        if (settingsDropdown.classList.contains('show')) {
          settingsDropdown.classList.remove('show');
          return;
        }
        search.value = '';
        filterServices('');
        search.blur();
        links.forEach(l => l.classList.remove('selected'));
        return;
      }
      if (e.key === ',') {
        e.preventDefault();
        settingsDropdown.classList.toggle('show');
        if (settingsDropdown.classList.contains('show')) {
          settingsFocusIndex = 0;
          settingsOptions.forEach((opt, i) => opt.classList.toggle('focused', i === 0));
          settingsDropdown.focus();
        }
        return;
      }
      if (e.key === '/') {
        e.preventDefault();
        toggle.checked = !toggle.checked;
        modeLabel.textContent = toggle.checked ? 'Local' : 'External';
        setMode(toggle.checked);
        filterServices(search.value);
        return;
      }
      if (e.key.length === 1 && !e.ctrlKey && !e.metaKey) {
        e.preventDefault();
        if (document.activeElement !== search) {
          search.value = '';
          search.focus();
        }
        search.value += e.key;
        filterServices(search.value);
      }
    });
  </script>
</body>
</html>'''

with open(output_file, 'w') as f:
    f.write(html)

print(f'Generated {output_file} with {len(services)} services')
