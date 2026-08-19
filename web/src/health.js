window.healthStatus = {};

async function checkAllHealth() {
  const links = document.querySelectorAll('.grid a[data-url]');
  if (!links.length) return;

  links.forEach(a => {
    let dot = a.querySelector('.health-dot');
    if (!dot) {
      dot = document.createElement('span');
      a.appendChild(dot);
    }
    dot.className = 'health-dot checking';
  });

  const promises = [...links].map(async (a) => {
    const url = a.dataset.url;
    const name = a.dataset.name;
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), 5000);
    try {
      await fetch(url, { mode: 'no-cors', signal: controller.signal });
      window.healthStatus[name] = 'online';
    } catch {
      window.healthStatus[name] = 'offline';
    } finally {
      clearTimeout(timer);
    }
  });

  await Promise.allSettled(promises);
  renderHealthIndicators();
}

function renderHealthIndicators() {
  document.querySelectorAll('.grid a[data-url]').forEach(a => {
    let dot = a.querySelector('.health-dot');
    if (!dot) {
      dot = document.createElement('span');
      a.appendChild(dot);
    }
    const status = window.healthStatus[a.dataset.name];
    dot.className = 'health-dot ' + (status || 'unknown');
  });
}

document.addEventListener('DOMContentLoaded', function() {
  setTimeout(checkAllHealth, 500);
});
