#!/usr/bin/env node
// Regenerate demo/web.gif: records the web UI (output/example.html) with headless Chrome.
// Sandbox: temp Chrome profile + throwaway server, real services.yml/localStorage untouched.
// Run: make demo-web   (needs Chrome + ffmpeg)
import { spawn } from 'node:child_process';
import { mkdtempSync, writeFileSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { createServer } from 'node:http';

const CHROME = '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
const ROOT = new URL('..', import.meta.url).pathname;
const W = 1280, H = 800;

const log = m => console.error('[web-demo]', m);
const sleep = ms => new Promise(r => setTimeout(r, ms));

// ── static file server for output/example.html ──────────
const page = readFileSync(join(ROOT, 'output', 'example.html'));
const server = createServer((req, res) => {
  res.writeHead(200, { 'content-type': 'text/html; charset=utf-8' });
  res.end(page);
});
server.listen(0, '127.0.0.1');
const port = await new Promise(resolve => server.once('listening', () => resolve(server.address().port)));
log('serving on :' + port);

// ── chrome + CDP ─────────────────────────────────────────
const profile = mkdtempSync(join(tmpdir(), 'homepage-demo-'));
const chrome = spawn(CHROME, [
  '--headless=new', `--remote-debugging-port=0`, `--user-data-dir=${profile}`,
  '--no-first-run', '--window-size=' + W + ',' + H, 'about:blank',
]);
const wsUrl = await new Promise((resolve, reject) => {
  let buf = '';
  chrome.stderr.on('data', d => {
    buf += d.toString();
    const m = buf.match(/DevTools listening on (ws:\/\/\S+)/);
    if (m) resolve(m[1]);
  });
  setTimeout(() => reject(new Error('chrome did not start')), 10000);
});

log('chrome up');
const ws = new WebSocket(wsUrl);
await new Promise(resolve => { ws.addEventListener('open', resolve, { once: true }); });
let msgId = 0;
const pending = new Map();
const events = [];
ws.onmessage = ev => {
  const msg = JSON.parse(ev.data);
  if (msg.id && pending.has(msg.id)) { pending.get(msg.id)(msg); pending.delete(msg.id); }
  else events.push(msg);
};
const send = (method, params = {}, sessionId) => new Promise(resolve => {
  const id = ++msgId;
  pending.set(id, resolve);
  ws.send(JSON.stringify({ id, method, params, sessionId }));
});
const waitEvent = (name, sessionId, timeout = 15000) => new Promise((resolve, reject) => {
  const t0 = Date.now();
  const iv = setInterval(() => {
    const i = events.findIndex(e => e.method === name && (!sessionId || e.sessionId === sessionId));
    if (i >= 0) { clearInterval(iv); resolve(events.splice(i, 1)[0]); }
    else if (Date.now() - t0 > timeout) { clearInterval(iv); reject(new Error('timeout: ' + name)); }
  }, 25);
});

log('cdp connected');
const { result: { targetId } } = await send('Target.createTarget', { url: 'about:blank' });
const { result: { sessionId } } = await send('Target.attachToTarget', { targetId, flatten: true });
const cdp = (m, p) => send(m, p, sessionId);

await cdp('Page.enable');
await cdp('Emulation.setDeviceMetricsOverride', { width: W, height: H, deviceScaleFactor: 2, mobile: false });

// ── helpers ──────────────────────────────────────────────
const framesDir = mkdtempSync(join(tmpdir(), 'homepage-frames-'));
let frameN = 0;
const timeline = [];
async function shot(duration) {
  const { result } = await cdp('Page.captureScreenshot', { format: 'png' });
  const file = join(framesDir, `f${String(++frameN).padStart(2, '0')}.png`);
  writeFileSync(file, Buffer.from(result.data, 'base64'));
  timeline.push({ file, duration });
}
const evalJs = async expression => {
  const { result } = await cdp('Runtime.evaluate', { expression, returnByValue: true });
  return result.result?.value;
};
const pressKey = async (key, { code = '', vk = 0, text, modifiers } = {}) => {
  await cdp('Input.dispatchKeyEvent', {
    type: text ? 'keyDown' : 'rawKeyDown', key, code, windowsVirtualKeyCode: vk,
    ...(text ? { text, unmodifiedText: text } : {}), ...(modifiers ? { modifiers } : {}),
  });
  await cdp('Input.dispatchKeyEvent', { type: 'keyUp', key, code, windowsVirtualKeyCode: vk });
};

// ── the demo ─────────────────────────────────────────────
// mirrors the TUI demo: health dots -> arrow navigation -> group cycling ->
// progressive search -> settings tour (light flash) -> shortcuts overlay.
log('attached, navigating');
await cdp('Page.navigate', { url: `http://127.0.0.1:${port}/example.html` });
await waitEvent('Page.loadEventFired', sessionId);
await sleep(400); await shot(0.6);            // default dark dashboard
await sleep(2200); await shot(2.0);           // health dots settled

// arrow-key navigation across cards
await pressKey('ArrowRight', { vk: 39 }); await shot(0.5);
await pressKey('ArrowRight', { vk: 39 }); await shot(0.5);
await pressKey('ArrowRight', { vk: 39 }); await shot(0.7);
await pressKey('ArrowDown', { vk: 40 }); await shot(0.9);

// "/" cycles group filter: News -> Media -> Social -> Tools -> All
await pressKey('/', { code: 'Slash', text: '/' }); await shot(1.0);
await pressKey('/', { code: 'Slash', text: '/' }); await shot(1.0);
await pressKey('/', { code: 'Slash', text: '/' }); await shot(0.6);
await pressKey('/', { code: 'Slash', text: '/' }); await shot(0.6);
await pressKey('/', { code: 'Slash', text: '/' }); await shot(0.9);

// progressive search: i -> in -> int
await pressKey('i', { text: 'i' }); await shot(0.8);
await pressKey('n', { text: 'n' }); await shot(0.8);
await pressKey('t', { text: 't' }); await shot(1.4);
await pressKey('Escape', { vk: 27 }); await shot(0.9);

// settings tour — fully keyboard-driven: arrows move the cursor, Enter toggles
await pressKey(',', { text: ',' }); await shot(1.4);            // "," opens settings, Theme focused
await pressKey('ArrowDown', { vk: 40 }); await pressKey('ArrowDown', { vk: 40 }); await shot(0.9); // -> Clock
await pressKey('Enter', { vk: 13, text: '\r' }); await shot(1.2); // 24h -> 12h
await pressKey('ArrowDown', { vk: 40 }); await pressKey('ArrowDown', { vk: 40 }); await pressKey('ArrowDown', { vk: 40 }); await shot(0.9); // -> Icons
await pressKey('Enter', { vk: 13, text: '\r' }); await shot(1.2); // Icons -> Off
await pressKey('Enter', { vk: 13, text: '\r' });                  // Icons -> On
await pressKey('ArrowDown', { vk: 40 });                          // -> Compact
await pressKey('Enter', { vk: 13, text: '\r' }); await shot(1.3); // Compact -> On
for (let i = 0; i < 6; i++) await pressKey('ArrowUp', { vk: 38 }); await shot(0.7); // back to Theme
await pressKey('Enter', { vk: 13, text: '\r' }); await shot(1.4); // light flash
await pressKey('Enter', { vk: 13, text: '\r' }); await shot(1.2); // back to dark
await pressKey('Escape', { vk: 27 }); await shot(1.3);            // close settings

// keyboard shortcuts overlay
await pressKey('?', { code: 'Slash', text: '?', modifiers: 8 }); await shot(2.0);
await pressKey('Escape', { vk: 27 }); await shot(1.6);

// ── assemble gif ─────────────────────────────────────────
log(`captured ${frameN} frames, assembling gif`);
const ffconcat = ['ffconcat version 1.0'];
for (const f of timeline) ffconcat.push(`file '${f.file}'`, `duration ${f.duration}`);
const last = timeline[timeline.length - 1];
ffconcat.push(`file '${last.file}'`);   // concat demuxer needs final frame twice
const listFile = join(framesDir, 'frames.txt');
writeFileSync(listFile, ffconcat.join('\n'));

await new Promise((resolve, reject) => {
  const ff = spawn('ffmpeg', [
    '-y', '-v', 'error', '-f', 'concat', '-safe', '0', '-i', listFile,
    '-vf', 'fps=12,scale=1152:-1:flags=lanczos,split[a][b];[a]palettegen=max_colors=128[p];[b][p]paletteuse=dither=bayer:bayer_scale=4',
    '-loop', '0', join(ROOT, 'demo', 'web.gif'),
  ], { stdio: 'inherit' });
  ff.on('exit', c => (c === 0 ? resolve() : reject(new Error('ffmpeg failed'))));
});

ws.close(); server.close();
await new Promise(resolve => { chrome.on('exit', resolve); chrome.kill(); });
rmSync(framesDir, { recursive: true, force: true });
rmSync(profile, { recursive: true, force: true });
console.log('wrote demo/web.gif');
