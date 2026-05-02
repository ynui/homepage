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

const groupToggle = document.getElementById('groupToggle');
const grid = document.querySelector('.grid');
const links = [...document.querySelectorAll('.grid a')];
const initialElements = [...grid.children];

const GROUPS = window.GROUPS || [{id: 'all', name: 'All'}];
let currentGroup = GROUPS[0].id;

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

function setGroup(groupId) {
  const isAll = groupId === 'all';
  if (isAll) {
    initialElements.forEach(el => grid.appendChild(el));
  } else {
    loadOrder();
  }
  document.querySelectorAll('.group-header').forEach(h => h.classList.toggle('hidden', !isAll));

  links.forEach(link => {
    const groups = link.dataset.groups.split(' ');
    if (isAll) {
      link.classList.remove('hidden');
      link.href = link.dataset.defaultUrl || link.href;
    } else if (groups.includes(groupId)) {
      link.classList.remove('hidden');
      link.href = link.dataset.defaultUrl || link.href;
    } else {
      link.classList.add('hidden');
    }
  });
}

const groupIndicator = document.getElementById('groupIndicator');
const groupValue = document.getElementById('groupValue');
let groupSelector = document.getElementById('groupSelector');
let groupOptions = [];

function cycleGroup() {
  const currentIndex = GROUPS.findIndex(g => g.id === currentGroup);
  const nextIndex = (currentIndex + 1) % GROUPS.length;
  currentGroup = GROUPS[nextIndex].id;
  groupToggle.querySelector('.value').textContent = GROUPS[nextIndex].name;
  const nextText = GROUPS[nextIndex].icon ? `${GROUPS[nextIndex].icon} ${GROUPS[nextIndex].name}` : GROUPS[nextIndex].name;
  groupValue.textContent = nextText;
  localStorage.setItem('homepage-group', currentGroup);
  setGroup(currentGroup);
  filterServices(search.value);
  showToast(`Group: ${GROUPS[nextIndex].name}`);
}

// Initialize group selector options from GROUPS
groupSelector = document.getElementById('groupSelector');
GROUPS.forEach(g => {
  const btn = document.createElement('button');
  btn.className = 'group-option';
  btn.dataset.group = g.id;
  btn.innerHTML = g.icon ? `${g.icon} ${g.name}` : g.name;
  groupSelector.appendChild(btn);
});

// Re-query groupOptions after dynamic creation
groupOptions = [...document.querySelectorAll('.group-option')];

const savedGroup = localStorage.getItem('homepage-group');
const initialGroup = GROUPS.find(g => g.id === savedGroup) ? savedGroup : GROUPS[0].id;
currentGroup = initialGroup;
const initialGroupObj = GROUPS.find(g => g.id === currentGroup);
const initialText = initialGroupObj.icon ? `${initialGroupObj.icon} ${initialGroupObj.name}` : initialGroupObj.name;
groupToggle.querySelector('.value').textContent = initialGroupObj.name;
groupValue.textContent = initialText;
setGroup(currentGroup);

groupToggle.addEventListener('click', cycleGroup);

groupIndicator.addEventListener('click', (e) => {
  e.stopPropagation();
  groupSelector.classList.toggle('active');
});

groupOptions.forEach(opt => {
  opt.addEventListener('click', (e) => {
    e.stopPropagation();
    currentGroup = opt.dataset.group;
    const groupObj = GROUPS.find(g => g.id === currentGroup);
    groupToggle.querySelector('.value').textContent = groupObj.name;
    const displayText = groupObj.icon ? `${groupObj.icon} ${groupObj.name}` : groupObj.name;
    groupValue.textContent = displayText;
    localStorage.setItem('homepage-group', currentGroup);
    setGroup(currentGroup);
    filterServices(search.value);
    groupSelector.classList.remove('active');
    showToast(`Group: ${groupObj.name}`);
  });
});

document.addEventListener('click', (e) => {
  if (!groupIndicator.contains(e.target)) {
    groupSelector.classList.remove('active');
  }
});

const toast = document.getElementById('toast');
let longPressTimer;

let toastTimeout = null;
function showToast(msg) {
  if (toastTimeout) clearTimeout(toastTimeout);
  toast.textContent = msg;
  toast.classList.add('show');
  toastTimeout = setTimeout(() => {
    toast.classList.remove('show');
    toastTimeout = null;
  }, 1500);
}

function copyLink(url) {
  if (!navigator.clipboard) {
    showToast('Use HTTPS to copy links');
    return;
  }
  navigator.clipboard.writeText(url).then(() => {
    showToast('Link copied!');
  }).catch(() => {
    showToast('Copy failed');
  });
}

links.forEach(link => {
  let longPressed = false;
  let touchMoved = false;
  let startX, startY;

  const startPress = (e) => {
    if (e.type === 'mousedown' && e.button !== 0) return;
    longPressed = false;
    touchMoved = false;
    if (e.type === 'touchstart') {
      startX = e.touches[0].clientX;
      startY = e.touches[0].clientY;
    }
    longPressTimer = setTimeout(() => {
      if (!touchMoved) {
        longPressed = true;
        if (navigator.vibrate) navigator.vibrate(50);
        showToast('Release to copy');
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
    } else if (e.type === 'mousemove') {
      // No movement threshold for mouse yet, could add if needed
    }
  };

  const endPress = (e) => {
    clearTimeout(longPressTimer);
    if (longPressed) {
      copyLink(link.href);
      // We set a flag to prevent the upcoming click
      setTimeout(() => { longPressed = false; }, 10);
    }
  };

  const handleClick = (e) => {
    if (longPressed) {
      e.preventDefault();
      e.stopPropagation();
      // longPressed is reset in endPress's timeout
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

const hintSettings = document.getElementById('hintSettings');
const hintHelp = document.getElementById('hintHelp');
const hintsOverlay = document.getElementById('hints');

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
  const isAll = currentGroup === 'all';
  let visibleLinks = [];
  links.forEach(link => {
    const groups = link.dataset.groups.split(' ');
    const isGroupMatch = isAll || groups.includes(currentGroup);
    const name = link.textContent.toLowerCase();
    const isMatch = isGroupMatch && (!q || name.includes(q));
    link.classList.toggle('hidden', !isMatch);
    if (isMatch) visibleLinks.push(link);
  });

  if (isAll) {
    document.querySelectorAll('.group-header').forEach(header => {
      let next = header.nextElementSibling;
      let hasVisible = false;
      while (next && !next.classList.contains('group-header')) {
        if (!next.classList.contains('hidden')) {
          hasVisible = true;
          break;
        }
        next = next.nextElementSibling;
      }
      header.classList.toggle('hidden', !hasVisible);
    });
  }

  selectedIndex = visibleLinks.length > 0 ? 0 : -1;
  updateSelection(visibleLinks);
}

function updateSelection(visibleLinks) {
  visibleLinks.forEach((link, i) => link.classList.remove('selected'));
  if (selectedIndex >= 0 && visibleLinks[selectedIndex]) {
    const selected = visibleLinks[selectedIndex];
    selected.classList.add('selected');
    selected.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
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
      cycleGroup();
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
    cycleGroup();
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
