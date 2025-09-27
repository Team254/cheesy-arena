// ===== Command Palette (Ctrl/Cmd + K) =====
(function () {
  const modalEl = document.getElementById('cmdPalette');
  const inputEl = document.getElementById('cmdInput');
  const listEl = document.getElementById('cmdResults');
  const openLink = document.getElementById('openCmdPalette');

  // Catalog of destinations (mirrors your menus)
  const CATALOG = [
    {g: 'Setup', t: 'Settings', u: '/setup/settings'},
    {g: 'Setup', t: 'Team List', u: '/setup/teams'},
    {g: 'Setup', t: 'Match Scheduling', u: '/setup/schedule'},
    {g: 'Setup', t: 'Judge Scheduling', u: '/setup/judging'},
    {g: 'Setup', t: 'Awards', u: '/setup/awards'},
    {g: 'Setup', t: 'Lower Thirds', u: '/setup/lower_thirds'},
    {g: 'Setup', t: 'Sponsor Slides', u: '/setup/sponsor_slides'},
    {g: 'Setup', t: 'Scheduled Breaks', u: '/setup/breaks'},
    {g: 'Setup', t: 'Display Configuration', u: '/setup/displays'},
    {g: 'Setup', t: 'Field Testing', u: '/setup/field_testing'},

    {g: 'Run', t: 'Match Play', u: '/match_play'},
    {g: 'Run', t: 'Match Review', u: '/match_review'},
    {g: 'Run', t: 'Match Logs', u: '/match_logs'},
    {g: 'Run', t: 'Alliance Selection', u: '/alliance_selection'},

    {g: 'Panels', t: 'Head Referee', u: '/panels/referee'},
    {g: 'Panels', t: 'Referee', u: '/panels/referee?hr=false'},
    {g: 'Scoring', t: 'Red Near', u: '/panels/scoring/red_near'},
    {g: 'Scoring', t: 'Red Far', u: '/panels/scoring/red_far'},
    {g: 'Scoring', t: 'Blue Near', u: '/panels/scoring/blue_near'},
    {g: 'Scoring', t: 'Blue Far', u: '/panels/scoring/blue_far'},

    {g: 'Reports (PDF)', t: 'Team List', u: '/reports/pdf/teams'},
    {g: 'Reports (PDF)', t: 'Practice Schedule', u: '/reports/pdf/schedule/practice'},
    {g: 'Reports (PDF)', t: 'Qualification Schedule', u: '/reports/pdf/schedule/qualification'},
    {g: 'Reports (PDF)', t: 'Playoff Schedule', u: '/reports/pdf/schedule/playoff'},
    {g: 'Reports (PDF)', t: 'Judging Schedule', u: '/reports/pdf/judging_schedule'},
    {g: 'Reports (PDF)', t: 'Standings', u: '/reports/pdf/rankings'},
    {g: 'Reports (PDF)', t: 'Playoff Alliances', u: '/reports/pdf/alliances'},
    {g: 'Reports (PDF)', t: 'Playoff Bracket', u: '/reports/pdf/bracket'},
    {g: 'Reports (PDF)', t: 'Backup Teams', u: '/reports/pdf/backups'},
    {g: 'Reports (PDF)', t: 'Playoff Alliance Coupons', u: '/reports/pdf/coupons'},
    {g: 'Reports (PDF)', t: 'Team Connection Status', u: '/reports/pdf/teams?showHasConnected=true'},
    {g: 'Reports (PDF)', t: 'Practice Cycle Report', u: '/reports/pdf/cycle/practice'},
    {g: 'Reports (PDF)', t: 'Qualification Cycle Report', u: '/reports/pdf/cycle/qualification'},
    {g: 'Reports (PDF)', t: 'Playoff Cycle Report', u: '/reports/pdf/cycle/playoff'},

    {g: 'Reports (CSV)', t: 'Team List', u: '/reports/csv/teams'},
    {g: 'Reports (CSV)', t: 'FTA Report', u: '/reports/csv/fta'},
    {g: 'Reports (CSV)', t: 'Practice Schedule', u: '/reports/csv/schedule/practice'},
    {g: 'Reports (CSV)', t: 'Qualification Schedule', u: '/reports/csv/schedule/qualification'},
    {g: 'Reports (CSV)', t: 'Playoff Schedule', u: '/reports/csv/schedule/playoff'},
    {g: 'Reports (CSV)', t: 'Standings', u: '/reports/csv/rankings'},
    {g: 'Reports (CSV)', t: 'Backup Teams', u: '/reports/csv/backups'},

    {g: 'Displays', t: 'Placeholder', u: '/display'},
    {g: 'Displays', t: 'Announcer', u: '/displays/announcer'},
    {g: 'Displays', t: 'Audience', u: '/displays/audience'},
    {g: 'Displays', t: 'Bracket', u: '/displays/bracket'},
    {g: 'Displays', t: 'Field Monitor', u: '/displays/field_monitor'},
    {g: 'Displays', t: 'Field Monitor (FTA)', u: '/displays/field_monitor?fta=true'},
    {g: 'Displays', t: 'Field Monitor (Blue DS)', u: '/displays/field_monitor?ds=true&reversed=true'},
    {g: 'Displays', t: 'Field Monitor (Red DS)', u: '/displays/field_monitor?ds=true&reversed=false'},
    {g: 'Displays', t: 'Logo', u: '/displays/logo'},
    {g: 'Displays', t: 'Queueing', u: '/displays/queueing'},
    {g: 'Displays', t: 'Standings', u: '/displays/rankings'},
    {g: 'Displays', t: 'Wall', u: '/displays/wall'},
    {g: 'Displays', t: 'Web Page', u: '/displays/webpage'},
    {g: 'Displays', t: 'Alliance Station: Red 1', u: '/displays/alliance_station?station=R1'},
    {g: 'Displays', t: 'Alliance Station: Red 2', u: '/displays/alliance_station?station=R2'},
    {g: 'Displays', t: 'Alliance Station: Red 3', u: '/displays/alliance_station?station=R3'},
    {g: 'Displays', t: 'Alliance Station: Blue 1', u: '/displays/alliance_station?station=B1'},
    {g: 'Displays', t: 'Alliance Station: Blue 2', u: '/displays/alliance_station?station=B2'},
    {g: 'Displays', t: 'Alliance Station: Blue 3', u: '/displays/alliance_station?station=B3'},
    {g: 'Displays', t: 'Alliance Station: Clock', u: '/displays/alliance_station?station=N2'},
    {g: 'Displays', t: 'Alliance Station: Red Score', u: '/displays/alliance_station?station=N3'},
    {g: 'Displays', t: 'Alliance Station: Blue Score', u: '/displays/alliance_station?station=N1'},
  ];

  function openModal() {
    const modal = new bootstrap.Modal(modalEl, {backdrop: true, keyboard: true});
    modal.show();
    setTimeout(() => inputEl.focus(), 100);
    inputEl.select();
    render('');
  }

  function closeModal() {
    const instance = bootstrap.Modal.getInstance(modalEl);
    if (instance) instance.hide();
  }

  // Basic fuzzy match: score title by occurrences (order-insensitive), prefer prefix hits
  function score(item, q) {
    if (!q) return 1;
    const hay = (item.g + ' ' + item.t).toLowerCase();
    const terms = q.toLowerCase().split(/\s+/).filter(Boolean);
    let s = 0;
    for (const term of terms) {
      const idx = hay.indexOf(term);
      if (idx === -1) return -1;
      s += (idx === 0 ? 3 : idx < 10 ? 2 : 1);
    }
    return s;
  }

  function render(q) {
    const rows = [];
    const scored = CATALOG.map(it => ({it, s: score(it, q)}))
      .filter(x => x.s >= 0)
      .sort((a, b) => b.s - a.s || a.it.t.localeCompare(b.it.t));

    // Group by 'g'
    let lastGroup = '';
    let idx = 0;
    for (const {it} of scored.slice(0, 50)) {
      if (it.g !== lastGroup) {
        rows.push(`<div class="group-label">${it.g}</div>`);
        lastGroup = it.g;
      }
      rows.push(
        `<a href="${it.u}" class="list-group-item list-group-item-action" role="option" data-idx="${idx}">
             <div class="d-flex justify-content-between">
               <span>${it.t}</span>
               <small class="text-muted">${it.u}</small>
             </div>
           </a>`);
      idx++;
    }
    listEl.innerHTML = rows.join('') || `<div class="group-label">No results</div>`;
    activeIndex = 0;
    updateActive();
  }

  let activeIndex = 0;

  function updateActive() {
    const items = listEl.querySelectorAll('.list-group-item');
    items.forEach((el, i) => el.classList.toggle('active', i === activeIndex));
    if (items[activeIndex]) items[activeIndex].scrollIntoView({block: 'nearest'});
  }

  function openActive() {
    const items = listEl.querySelectorAll('.list-group-item');
    const el = items[activeIndex];
    if (el) window.location.assign(el.getAttribute('href'));
  }

  // Events
  inputEl.addEventListener('input', () => render(inputEl.value));
  inputEl.addEventListener('keydown', (e) => {
    const items = listEl.querySelectorAll('.list-group-item');
    if (e.key === 'ArrowDown') {
      activeIndex = Math.min(activeIndex + 1, items.length - 1);
      updateActive();
      e.preventDefault();
    }
    if (e.key === 'ArrowUp') {
      activeIndex = Math.max(activeIndex - 1, 0);
      updateActive();
      e.preventDefault();
    }
    if (e.key === 'Enter') {
      openActive();
    }
    if (e.key === 'Escape') {
      closeModal();
    }
  });
  listEl.addEventListener('click', (e) => {
    const a = e.target.closest('a.list-group-item');
    if (a) {
      window.location.assign(a.getAttribute('href'));
      e.preventDefault();
    }
  });

  // Keyboard shortcut: Ctrl/Cmd + K
  document.addEventListener('keydown', (e) => {
    const tag = (document.activeElement && document.activeElement.tagName) || '';
    if (['INPUT', 'TEXTAREA', 'SELECT'].includes(tag) && !(e.metaKey || e.ctrlKey)) return;
    if ((e.key === 'k' || e.key === 'K') && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      openModal();
    }
  });

  // Navbar link
  if (openLink) openLink.addEventListener('click', (e) => {
    e.preventDefault();
    openModal();
  });
})();