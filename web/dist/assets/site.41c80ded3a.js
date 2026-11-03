// Matrix rain background and service worker registration. Pure decoration plus
// offline support; nothing here touches the poem.
(function () {
  'use strict';

  if ('serviceWorker' in navigator && (location.protocol === 'https:' || location.hostname === 'localhost')) {
    window.addEventListener('load', function () {
      navigator.serviceWorker.register('/sw.js').catch(function () {});
    });
  }

  var canvas = document.getElementById('rain');
  if (!canvas || !canvas.getContext) return;
  if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;

  var ctx = canvas.getContext('2d');
  var glyphs = 'ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉﾊﾋﾌﾍﾎﾏﾐﾑﾒﾓﾔﾕﾖﾗﾘﾙﾚﾛﾜ0123456789';
  var size = 16, cols = 0, drops = [], last = 0, frame = 1000 / 18, running = true;

  // Decorative only, so Math.random is fine here (the poem uses crypto.getRandomValues).
  function resize() {
    var w = window.innerWidth, h = window.innerHeight;
    size = w < 600 ? 14 : 16;
    canvas.width = w;
    canvas.height = h;
    cols = Math.ceil(w / size);
    drops = [];
    for (var i = 0; i < cols; i++) drops.push(Math.floor(Math.random() * -h / size));
    ctx.fillStyle = '#020a05';
    ctx.fillRect(0, 0, w, h);
  }

  function draw(t) {
    if (!running) return;
    requestAnimationFrame(draw);
    if (t - last < frame) return;
    last = t;
    ctx.fillStyle = 'rgba(2, 10, 5, 0.14)';
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    ctx.font = size + 'px monospace';
    for (var i = 0; i < cols; i++) {
      var y = drops[i] * size;
      if (y > 0) {
        var ch = glyphs.charAt(Math.floor(Math.random() * glyphs.length));
        ctx.fillStyle = Math.random() < 0.06 ? '#d9ffe9' : '#00ff9c';
        ctx.fillText(ch, i * size, y);
      }
      if (y > canvas.height && Math.random() > 0.975) drops[i] = 0;
      drops[i]++;
    }
  }

  document.addEventListener('visibilitychange', function () {
    var was = running;
    running = !document.hidden;
    if (running && !was) requestAnimationFrame(draw);
  });
  window.addEventListener('resize', resize);
  resize();
  requestAnimationFrame(draw);
})();
