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

const modeToggle = document.getElementById('modeToggle');
const grid = document.querySelector('.grid');
const links = [...document.querySelectorAll('.grid a')];

const MODES = ['External', 'Local', 'All'];
let currentMode = 0;

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

function setMode(mode) {
  const isLocal = mode === 1;
  const isAll = mode === 2;
  links.forEach(link => {
    const type = link.dataset.type;
    if (type === 'both') {
      link.href = isLocal ? link.dataset.local : link.dataset.external;
      link.classList.remove('hidden');
    } else if (type === 'local') {
      link.href = link.dataset.local;
      link.classList.toggle('hidden', !isLocal && !isAll);
    } else if (type === 'external') {
      link.href = link.dataset.external;
      link.classList.toggle('hidden', isLocal && !isAll);
    }
  });
}

const modeIndicator = document.getElementById('modeIndicator');

function cycleMode() {
  currentMode = (currentMode + 1) % 3;
  modeToggle.querySelector('.value').textContent = MODES[currentMode];
  modeIndicator.textContent = MODES[currentMode];
  localStorage.setItem('homepage-mode', currentMode);
  setMode(currentMode);
  filterServices(search.value);
  showToast(`Mode: ${MODES[currentMode]}`);
}

currentMode = parseInt(localStorage.getItem('homepage-mode')) || 0;
modeToggle.querySelector('.value').textContent = MODES[currentMode];
modeIndicator.textContent = MODES[currentMode];
setMode(currentMode);

modeToggle.addEventListener('click', cycleMode);

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

const hintMode = document.getElementById('hintMode');
const hintSettings = document.getElementById('hintSettings');
const hintHelp = document.getElementById('hintHelp');
const hintsOverlay = document.getElementById('hints');

hintMode.addEventListener('click', (e) => {
  e.stopPropagation();
  cycleMode();
});
hintSettings.addEventListener('click', (e) => {
  e.stopPropagation();
  settingsDropdown.classList.toggle('show');
  if (settingsDropdown.classList.contains('show')) {
    settingsFocusIndex = 0;
    settingsOptions.forEach((opt, i) => opt.classList.toggle('focused', i === 0));
    settingsDropdown.focus();
  }
});
hintHelp.addEventListener('click', (e) => {
  e.stopPropagation();
  hintsOverlay.classList.toggle('show');
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
  const isFooterHint = e.target.closest('.hints-footer');
  if (!settingsDropdown.contains(e.target) && e.target !== settingsBtn && !isFooterHint) {
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
  themeToggle.querySelector('.value').textContent = light ? '☀️' : '🌙';
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
  clockToggle.querySelector('.value').textContent = is24 ? '24h' : '12h';
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
  dateToggle.querySelector('.value').textContent = visible ? 'On' : 'Off';
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
  searchToggle.querySelector('.value').textContent = visible ? 'On' : 'Off';
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
  iconToggle.querySelector('.value').textContent = visible ? 'On' : 'Off';
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
  compactToggle.querySelector('.value').textContent = compact ? 'On' : 'Off';
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
  const isLocal = currentMode === 1;
  const isAll = currentMode === 2;
  let visibleLinks = [];
  links.forEach(link => {
    const type = link.dataset.type;
    let isTypeMatch = true;
    if (type === 'local') isTypeMatch = isLocal || isAll;
    else if (type === 'external') isTypeMatch = !isLocal || isAll;
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
    if (e.key === ',') {
      e.preventDefault();
      settingsDropdown.classList.remove('show');
      return;
    }
    return;
  }
  if (e.target.tagName === 'INPUT') {
    if (e.key === '/') {
      e.preventDefault();
      cycleMode();
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
    cycleMode();
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
