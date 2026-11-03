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
  var numbered = $('numbered');
  var numberedList = $('numbered-list');
  var status = $('poem-status');
  var timer = $('timer');
  var timerValue = $('timer-value');
  var actions = $('actions');
  var copyBtn = $('copy-btn');
  var pdfBtn = $('pdf-btn');
  var hideBtn = $('hide-btn');
  var title = $('poem-title');

  var ttl = parseInt(box.getAttribute('data-ttl'), 10) || 900;
  var words = box.getAttribute('data-words') === '24' ? 24 : 12;
  var current = null; // array of lines, each an array of 3 words
  var expiresAt = 0;
  var ticker = 0;
  var gen = null;

  function clear(el) {
    while (el.firstChild) el.removeChild(el.firstChild);
  }

  function setStatus(msg, bad) {
    status.textContent = msg;
    status.className = bad ? 'poem-status expired' : 'poem-status';
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

  function wipe(message) {
    current = null;
    expiresAt = 0;
    if (ticker) { clearInterval(ticker); ticker = 0; }
    clear(lines);
    clear(numberedList);
    numbered.hidden = true;
    numbered.open = false;
    timer.hidden = true;
    actions.hidden = true;
    hideBtn.classList.remove('attention');
    hideBtn.textContent = 'I saved it, hide now';
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
    expiresAt = Date.now() + ttl * 1000;
    timer.hidden = false;
    actions.hidden = false;
    numbered.hidden = false;
    pdfBtn.hidden = !window.SeedPoetPDF;
    copyBtn.hidden = !(navigator.clipboard && navigator.clipboard.writeText);
    setStatus('Generated on this device. Nothing was sent or stored.', false);
    tick();
    if (ticker) clearInterval(ticker);
    ticker = setInterval(tick, 1000);
  }

  function generate(n) {
    words = n === 24 ? 24 : 12;
    box.setAttribute('data-words', String(words));
    if (title) title.textContent = words + '-word poem';
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

  copyBtn.addEventListener('click', function () {
    if (!current) return;
    navigator.clipboard.writeText(text()).then(function () {
      copyBtn.textContent = 'Copied';
      setTimeout(function () { copyBtn.textContent = 'Copy'; }, 2000);
      nudge('Copied');
    });
  });

  pdfBtn.addEventListener('click', function () {
    if (!current || !window.SeedPoetPDF) return;
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

  // "New poem" / "12 words" / "24 words" generate in place; the links still work without JS.
  Array.prototype.forEach.call(document.querySelectorAll('[data-generate]'), function (el) {
    el.addEventListener('click', function (e) {
      e.preventDefault();
      var n = parseInt(el.getAttribute('data-generate'), 10);
      if (box.getAttribute('data-path') === 'page' && history.replaceState) {
        history.replaceState(null, '', '/' + n);
      }
      generate(n);
      var other = $('switch-btn');
      if (other) {
        var to = n === 24 ? 12 : 24;
        other.setAttribute('data-generate', String(to));
        other.setAttribute('href', '/' + to);
        other.textContent = 'Switch to ' + to;
      }
    });
  });

  var noscript = $('noscript-box');
  if (noscript) noscript.hidden = true;

  if (box.getAttribute('data-static') === '1') {
    // Server-rendered poem (no-JavaScript fallback page): take over the timer.
    var server = [];
    Array.prototype.forEach.call(lines.querySelectorAll('li'), function (li) {
      server.push(li.textContent.replace(/^\d+/, '').trim().split(/\s+/));
    });
    if (server.length) show(server);
  } else if (box.getAttribute('data-autostart') === '1') {
    generate(words);
  } else {
    setStatus('Choose 12 or 24 words. The poem is generated on this device.', false);
  }
})();
