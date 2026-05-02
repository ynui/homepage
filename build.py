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

# Parse groups
groups = data.get('groups', [])
groups_list = [{'id': g['id'], 'name': g['name'], 'icon': g.get('icon', '')} for g in groups]
group_ids = {g['id'] for g in groups_list}

# Parse services
services_list = data.get('services', [])
services = []
has_nogroup = False
for svc in services_list:
    svc_groups = svc.get('groups', [])
    valid_groups = [g for g in svc_groups if g in group_ids]
    if not valid_groups:
        valid_groups = ['nogroup']
        has_nogroup = True
    svc['groups'] = valid_groups
    services.append(svc)

if has_nogroup:
    groups_list.append({'id': 'nogroup', 'name': 'No Group', 'icon': ''})

# Add implicit 'all' group
groups_list.append({'id': 'all', 'name': 'All', 'icon': ''})

# Generate services HTML
services_html = ''
rendered_names = set()

for g in groups_list:
    if g['id'] == 'all':
        continue
    
    group_services = [s for s in services if g['id'] in s['groups'] and s['name'] not in rendered_names]
    if group_services:
        services_html += f'      <div class="group-header" data-group="{g["id"]}">{g["name"]}</div>\n'
        for s in group_services:
            url = s.get('url', '')
            groups_attr = ' '.join(s['groups'])
            icon = s.get('icon', '⚙')
            services_html += f'''      <a href="{url}" data-name="{s['name'].lower()}" data-groups="{groups_attr}" data-default-url="{url}">
        <span class="icon">{icon}</span>
        {s['name']}
      </a>\n'''
            rendered_names.add(s['name'])

# JSON for JS
groups_json = json.dumps(groups_list)

with open('src/template.html') as f:
    template = f.read()

with open('src/style.css') as f:
    css = f.read()

with open('src/script.js') as f:
    js = f.read()

html = (template
    .replace('{{__TITLE__}}', title)
    .replace('{{__HEADER__}}', header)
    .replace('{{__SERVICES__}}', services_html)
    .replace('{{__CSS__}}', css)
    .replace('{{__JS__}}', js)
    .replace('{{__GROUPS_JSON__}}', groups_json))

with open(output_file, 'w') as f:
    f.write(html)

print(f'Generated {output_file} with {len(services)} services')
