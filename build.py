#!/usr/bin/env python3

with open('services.yml') as f:
    content = f.read()

title = 'Homelab'
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

    .grid {{
      display: flex;
      flex-wrap: wrap;
      gap: 0.75rem;
      justify-content: center;
    }}

    .grid a {{
      width: 140px;
      flex-shrink: 0;
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

    .icon {{
      font-size: 1.5rem;
      opacity: 0.7;
      transition: opacity 0.3s;
    }}

    a:hover .icon {{
      opacity: 1;
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
  </style>
</head>
<body>
  <main>
    <header>
      <h1>{title}</h1>
      <div class="time" id="time"></div>
      <label class="toggle">
        <input type="checkbox" id="modeToggle" class="toggle-input">
        <span class="toggle-switch"></span>
        <span id="modeLabel" class="toggle-label">External</span>
      </label>
    </header>
    <div class="grid">
'''

for s in services:
    url = s.get('url', '')
    local = s.get('local', url)
    html += f'''      <a href="{url or local}" data-external="{url}" data-local="{local}" data-type="{s['type']}">
        <span class="icon">{s['icon']}</span>
        {s['name']}
      </a>
'''

html += '''    </div>
  </main>
  <div class="toast" id="toast">Copied!</div>
  <script>
    function updateTime() {
      const now = new Date();
      document.getElementById('time').textContent = now.toLocaleTimeString('en-US', { 
        hour12: false, 
        hour: '2-digit', 
        minute: '2-digit' 
      });
    }
    updateTime();
    setInterval(updateTime, 1000);

    const toggle = document.getElementById('modeToggle');
    const links = document.querySelectorAll('.grid a');

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
  </script>
</body>
</html>'''

with open('index.html', 'w') as f:
    f.write(html)

print(f'Generated index.html with {len(services)} services')
