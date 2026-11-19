// Seed Poet page logic. The poem is generated right here, in this page, by the
// Seed Poet library with crypto.getRandomValues. It is never sent anywhere (the
// site's Content-Security-Policy has connect-src 'none'), never stored, and it is
// wiped after the time limit, on "Hide now", and whenever the page is left.
(function () {
  'use strict';

  var box = document.getElementById('poem');
  if (!box) return;

  var $ = function (id) { return document.getElementById(id); };
  var lines = $('poem-lines');
  var placeholder = $('placeholder');
  var numbered = $('numbered');
  var numberedList = $('numbered-list');
  var status = $('poem-status');
  var timer = $('timer');
  var timerValue = $('timer-value');
  var actions = $('actions');
  var copyBtn = $('copy-btn');
  var pdfBtn = $('pdf-btn');
  var hideBtn = $('hide-btn');
  var drillBtn = $('drill-btn');
  var checkBtn = $('check-btn');
  var Drill = window.SeedPoetDrill;
  var wordsLabel = $('words-label');
  var pickers = document.querySelectorAll('[data-generate]');

  var ttl = parseInt(box.getAttribute('data-ttl'), 10) || 900;
  var words = box.getAttribute('data-words') === '24' ? 24 : 12;
  var current = null; // array of lines, each an array of 3 words
  var expiresAt = 0;
  var ticker = 0;
  var gen = null;
  // Drill state, in memory only and wiped with the poem: level (words hidden per
  // line), hidden word positions, the input elements and whether the round is graded.
  var drill = null;

  function clear(el) {
    while (el.firstChild) el.removeChild(el.firstChild);
  }

  function setStatus(msg, bad) {
    status.textContent = msg;
    status.className = bad ? 'poem-status expired' : 'poem-status';
  }

  function markPicker() {
    Array.prototype.forEach.call(pickers, function (b) {
      b.setAttribute('aria-pressed', current && b.getAttribute('data-generate') === String(words) ? 'true' : 'false');
    });
  }

  function render(poemLines) {
    clear(lines);
    clear(numberedList);
    poemLines.forEach(function (line, i) {
      var li = document.createElement('li');
      var ln = document.createElement('span');
      ln.className = 'ln';
      ln.textContent = (i < 9 ? '0' : '') + (i + 1);
      li.appendChild(ln);
      li.appendChild(document.createTextNode(line[0] + ' ' + line[1] + ' '));
      var r = document.createElement('span');
      r.className = 'rhyme';
      r.textContent = line[2];
      li.appendChild(r);
      lines.appendChild(li);
      line.forEach(function (w) {
        var n = document.createElement('li');
        n.textContent = w;
        numberedList.appendChild(n);
      });
    });
  }

  function setDrillUI(on) {
    if (!drillBtn) return;
    drillBtn.textContent = on ? 'Show poem' : 'Drill';
    drillBtn.setAttribute('aria-pressed', on ? 'true' : 'false');
    checkBtn.hidden = !on;
    checkBtn.textContent = 'Check';
    if (current) {
      copyBtn.hidden = on || !(navigator.clipboard && navigator.clipboard.writeText);
      pdfBtn.hidden = on || !window.SeedPoetPDF;
      numbered.hidden = on;
      if (on) numbered.open = false;
    }
  }

  function drillRound() {
    drill.hidden = Drill.blanks(current, drill.level);
    drill.inputs = [];
    drill.graded = false;
    clear(lines);
    current.forEach(function (line, i) {
      var li = document.createElement('li');
      var ln = document.createElement('span');
      ln.className = 'ln';
      ln.textContent = (i < 9 ? '0' : '') + (i + 1);
      li.appendChild(ln);
      var row = [];
      line.forEach(function (w, k) {
        if (k) li.appendChild(document.createTextNode(' '));
        if (drill.hidden[i].indexOf(k) < 0) {
          var sp = document.createElement('span');
          if (k === 2) sp.className = 'rhyme';
          sp.textContent = w;
          li.appendChild(sp);
          return;
        }
        var inp = document.createElement('input');
        inp.type = 'text';
        inp.className = 'drill-word' + (k === 2 ? ' rhyme' : '');
        inp.size = 8;
        inp.maxLength = 8;
        inp.setAttribute('autocomplete', 'off');
        inp.setAttribute('autocapitalize', 'none');
        inp.setAttribute('autocorrect', 'off');
        inp.setAttribute('spellcheck', 'false');
        inp.setAttribute('aria-label', 'Line ' + (i + 1) + ', word ' + (k + 1));
        li.appendChild(inp);
        row.push(inp);
      });
      drill.inputs.push(row);
      lines.appendChild(li);
    });
    var all = [].concat.apply([], drill.inputs);
    all.forEach(function (inp, n) {
      inp.addEventListener('keydown', function (e) {
        if (e.key !== 'Enter') return;
        e.preventDefault();
        if (n + 1 < all.length) all[n + 1].focus(); else drillCheck();
      });
    });
    var per = drill.level === Drill.MAX_LEVEL ? 'every word' : drill.level + (drill.level === 1 ? ' word' : ' words') + ' per line';
    setStatus('Drill, round ' + drill.level + ' of ' + Drill.MAX_LEVEL + ': type the missing words (' + per + '). Nothing is stored.', false);
    checkBtn.textContent = 'Check';
    if (all.length) all[0].focus();
  }

  function drillCheck() {
    if (!drill || !current) return;
    if (drill.graded) { drillRound(); return; }
    var answers = drill.inputs.map(function (row) { return row.map(function (inp) { return inp.value; }); });
    var res = Drill.grade(current, drill.hidden, answers);
    answers = null;
    drill.inputs.forEach(function (row, i) {
      row.forEach(function (inp, k) {
        inp.readOnly = true;
        if (res.ok[i][k]) {
          inp.classList.add('good');
        } else {
          inp.classList.add('bad');
          inp.value = current[i][drill.hidden[i][k]];
        }
      });
    });
    drill.graded = true;
    var lvl = Drill.next(drill.level, res);
    var msg = res.right + ' of ' + res.total + ' right. ';
    if (res.right === res.total && drill.level === Drill.MAX_LEVEL) msg += 'You know the whole poem by heart.';
    else if (lvl > drill.level) msg += 'Next round hides more words.';
    else msg += 'The right words are shown in red: try again.';
    drill.level = lvl;
    setStatus(msg, res.right !== res.total);
    checkBtn.textContent = 'Next round';
    checkBtn.focus();
  }

  function drillStop() {
    drill = null;
    setDrillUI(false);
    if (current) {
      render(current);
      setStatus('Drill closed. Nothing you typed was kept.', false);
    }
  }

  function wipe(message) {
    drill = null;
    setDrillUI(false);
    current = null;
    expiresAt = 0;
    if (ticker) { clearInterval(ticker); ticker = 0; }
    clear(lines);
    clear(numberedList);
    if (placeholder) placeholder.hidden = false;
    numbered.hidden = true;
    numbered.open = false;
    timer.hidden = true;
    actions.hidden = true;
    hideBtn.classList.remove('attention');
    hideBtn.textContent = 'I saved it, hide now';
    markPicker();
    setStatus(message, true);
  }

  function tick() {
    if (!current) return;
    var left = Math.round((expiresAt - Date.now()) / 1000);
    if (left <= 0) {
      wipe('Time is up: the poem was removed from this page. Generate a new one.');
      return;
    }
    var m = Math.floor(left / 60), s = left % 60;
    timerValue.textContent = (m < 10 ? '0' : '') + m + ':' + (s < 10 ? '0' : '') + s;
    timerValue.className = left <= 60 ? 'timer-value low' : 'timer-value';
  }

  function show(poemLines) {
    current = poemLines;
    render(poemLines);
    if (placeholder) placeholder.hidden = true;
    wordsLabel.textContent = String(words);
    expiresAt = Date.now() + ttl * 1000;
    timer.hidden = false;
    actions.hidden = false;
    numbered.hidden = false;
    pdfBtn.hidden = !window.SeedPoetPDF;
    copyBtn.hidden = !(navigator.clipboard && navigator.clipboard.writeText);
    hideBtn.classList.remove('attention');
    hideBtn.textContent = 'I saved it, hide now';
    drill = null;
    setDrillUI(false);
    if (drillBtn) drillBtn.hidden = !Drill;
    markPicker();
    setStatus('Generated on this device. Nothing was sent or stored.', false);
    tick();
    if (ticker) clearInterval(ticker);
    ticker = setInterval(tick, 1000);
    // Narrow screens stack the poem under the controls: bring it into view.
    if (window.matchMedia && window.matchMedia('(max-width: 860px)').matches) {
      box.querySelector('.out').scrollIntoView({ block: 'start' });
    }
  }

  function generate(n) {
    words = n === 24 ? 24 : 12;
    try {
      if (!gen) gen = window.SeedPoet.createGenerator(window.SeedPoetData);
      var poem = gen.generate(words);
      show(poem.lines);
      poem = null;
    } catch (e) {
      wipe('Could not generate a poem in this browser (' + e.message + ').');
    }
  }

  function text() {
    return current ? current.map(function (l) { return l.join(' '); }).join('\n') : '';
  }

  function nudge(label) {
    hideBtn.classList.add('attention');
    hideBtn.textContent = label + ': hide the poem now';
  }

  // Leaving the page always removes the poem, so Back or history never shows it.
  window.addEventListener('pagehide', function () {
    wipe('The poem was removed when you left the page.');
  });
  window.addEventListener('pageshow', function (e) {
    if (e.persisted) wipe('The poem was removed when you left the page.');
  });
  document.addEventListener('visibilitychange', tick);
  window.addEventListener('focus', tick);

  hideBtn.addEventListener('click', function (e) {
    e.preventDefault();
    wipe('Poem hidden. It was never stored and cannot be shown again.');
  });

  if (drillBtn) {
    drillBtn.addEventListener('click', function () {
      if (!current || !Drill) return;
      if (drill) { drillStop(); return; }
      drill = { level: 1 };
      setDrillUI(true);
      drillRound();
    });
    checkBtn.addEventListener('click', drillCheck);
  }

  copyBtn.addEventListener('click', function () {
    if (!current || drill) return;
    navigator.clipboard.writeText(text()).then(function () {
      copyBtn.textContent = 'Copied';
      setTimeout(function () { copyBtn.textContent = 'Copy'; }, 2000);
      nudge('Copied');
    });
  });

  pdfBtn.addEventListener('click', function () {
    if (!current || drill || !window.SeedPoetPDF) return;
    var url = URL.createObjectURL(window.SeedPoetPDF.build(current.map(function (l) { return l.join(' '); })));
    var a = document.createElement('a');
    a.href = url;
    a.download = 'seed-poem-' + words + '-words.pdf';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    setTimeout(function () { URL.revokeObjectURL(url); }, 1500);
    nudge('PDF saved');
  });

  // The 12 / 24 buttons submit to the server only without JavaScript; here they
  // generate on this device instead.
  Array.prototype.forEach.call(pickers, function (el) {
    el.addEventListener('click', function (e) {
      e.preventDefault();
      generate(parseInt(el.getAttribute('data-generate'), 10));
    });
  });

  if (box.getAttribute('data-static') === '1') {
    // Server-rendered poem (no-JavaScript fallback page): take over the timer.
    var server = [];
    Array.prototype.forEach.call(lines.querySelectorAll('li'), function (li) {
      server.push(li.textContent.replace(/^\d+/, '').trim().split(/\s+/));
    });
    if (server.length) show(server);
    return;
  }

  // App shortcuts and old links (/12, /24 redirect here as /?words=N) start a poem at once.
  var asked = /[?&]words=(12|24)\b/.exec(location.search);
  if (asked) {
    if (history.replaceState) history.replaceState(null, '', location.pathname);
    generate(parseInt(asked[1], 10));
  }
})();
