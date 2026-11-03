// Seed Poet: the poem lives only in this page's memory and DOM.
// It is fetched with a POST that no cache keeps, never put in the URL or in
// storage, and wiped when the time is up or when the user leaves the page
// (so Back, history and the back/forward cache never show an old poem).
(function () {
  'use strict';

  var box = document.getElementById('poem');
  if (!box) return;
  var text = document.getElementById('poem-text');
  var status = document.getElementById('poem-status');
  var copyBtn = document.getElementById('copy-btn');
  var ttl = parseInt(box.getAttribute('data-ttl'), 10) || 900;
  var words = box.getAttribute('data-words') === '24' ? 24 : 12;
  var expiresAt = 0;
  var timer = 0;
  var wiped = false;

  function wipe(message) {
    wiped = true;
    if (timer) { clearInterval(timer); timer = 0; }
    text.textContent = '';
    while (text.firstChild) text.removeChild(text.firstChild);
    copyBtn.hidden = true;
    status.className = 'status expired';
    status.textContent = message;
  }

  function tick() {
    if (wiped || !expiresAt) return;
    var left = Math.round((expiresAt - Date.now()) / 1000);
    if (left <= 0) {
      wipe('This poem has expired and was removed from the page. Generate a new one.');
      return;
    }
    var m = Math.floor(left / 60), s = left % 60;
    status.className = 'status';
    status.textContent = 'Disappears in ' + m + ':' + (s < 10 ? '0' : '') + s;
  }

  function show(poem) {
    text.textContent = poem;
    expiresAt = Date.now() + ttl * 1000;
    copyBtn.hidden = false;
    tick();
    timer = setInterval(tick, 1000);
  }

  // Leaving the page always removes the poem, so it never comes back from history.
  window.addEventListener('pagehide', function () {
    wipe('The poem was removed when you left the page. Generate a new one.');
  });
  window.addEventListener('pageshow', function (e) {
    if (e.persisted) wipe('The poem was removed when you left the page. Generate a new one.');
  });
  document.addEventListener('visibilitychange', tick);
  window.addEventListener('focus', tick);

  copyBtn.addEventListener('click', function () {
    if (wiped || !navigator.clipboard) return;
    navigator.clipboard.writeText(text.textContent).then(function () {
      copyBtn.textContent = 'Copied!';
      copyBtn.classList.add('copied');
      setTimeout(function () {
        copyBtn.textContent = 'Copy';
        copyBtn.classList.remove('copied');
      }, 2000);
    });
  });

  // No-JavaScript fallback page already holds the poem: just run the timer.
  if (box.getAttribute('data-static') === '1') {
    show(text.textContent);
    return;
  }

  status.textContent = 'Writing your poem...';
  fetch('/api/poem', {
    method: 'POST',
    cache: 'no-store',
    credentials: 'omit',
    referrerPolicy: 'no-referrer',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ words: words })
  }).then(function (r) {
    if (!r.ok) throw new Error('HTTP ' + r.status);
    return r.json();
  }).then(function (data) {
    if (wiped) return;
    if (data.ttl) ttl = data.ttl;
    show(data.lines.join('\n'));
    data = null;
  }).catch(function () {
    status.className = 'status expired';
    status.textContent = 'Could not generate a poem. Please try again.';
  });
})();
