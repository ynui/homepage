#!/usr/bin/env python3
import sys

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

services_html = ''
for s in services:
    url = s.get('url', '')
    local = s.get('local', url)
    services_html += f'''      <a href="{url or local}" data-name="{s['name'].lower()}" data-external="{url}" data-local="{local}" data-type="{s['type']}">
        <span class="icon">{s['icon']}</span>
        {s['name']}
      </a>
'''

with open('src/template.html') as f:
    template = f.read()

with open('src/style.css') as f:
    css = f.read()

with open('src/script.js') as f:
    js = f.read()

html = template.replace('{{title}}', title).replace('{{services}}', services_html).replace('{{css}}', css).replace('{{js}}', js)

with open(output_file, 'w') as f:
    f.write(html)

print(f'Generated {output_file} with {len(services)} services')
