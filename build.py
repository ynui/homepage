#!/usr/bin/env python3
import sys
import json
import yaml

example = None
if len(sys.argv) > 1 and sys.argv[1].startswith('--'):
    example = sys.argv[1][2:]

yaml_file = f'services.{example}.yml' if example else 'services.yml'
output_file = f'{example}.html' if example else 'index.html'

with open(yaml_file) as f:
    data = yaml.safe_load(f) or {}

title = data.get('title', 'Home Page')
header = data.get('header', title)

# Parse groups (explicit + implicit 'all')
groups = data.get('groups', [])
groups_list = [{'id': g['id'], 'name': g['name'], 'icon': g.get('icon', '')} for g in groups]
# Add implicit 'all' group
groups_list.append({'id': 'all', 'name': 'All', 'icon': ''})

# Build group lookup
group_ids = {g['id'] for g in groups_list}

# Parse services (list format)
services_list = data.get('services', [])
services = []
for svc in services_list:
    # Default: belongs to 'all' if no groups specified
    svc_groups = svc.get('groups', ['all'])
    svc['groups'] = [g for g in svc_groups if g in group_ids or g == 'all']
    if not svc['groups']:
        svc['groups'] = ['all']
    services.append(svc)

# Generate services HTML
services_html = ''
for s in services:
    url = s.get('url', '')
    groups_attr = ' '.join(s['groups'])
    icon = s.get('icon', '⚙')
    # Store default URL for the service
    services_html += f'''      <a href="{url}" data-name="{s['name'].lower()}" data-groups="{groups_attr}" data-default-url="{url}">
        <span class="icon">{icon}</span>
        {s['name']}
      </a>
'''

# JSON for JS
groups_json = json.dumps(groups_list)

with open('src/template.html') as f:
    template = f.read()

with open('src/style.css') as f:
    css = f.read()

with open('src/script.js') as f:
    js = f.read()

html = (template
    .replace('{{title}}', title)
    .replace('{{header}}', header)
    .replace('{{services}}', services_html)
    .replace('{{css}}', css)
    .replace('{{js}}', js)
    .replace('{{groups_json}}', groups_json))

with open(output_file, 'w') as f:
    f.write(html)

print(f'Generated {output_file} with {len(services)} services')
