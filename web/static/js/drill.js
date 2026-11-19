// Seed Poet drill: an offline quiz on the poem shown on this page. Pure logic, no
// DOM, no storage, no network: app.js keeps the quiz in memory and wipes it with the
// poem. Rounds blank 1, then 2, then all 3 words of every line ("by heart").
(function (root, factory) {
  if (typeof module === 'object' && module.exports) module.exports = factory();
  else root.SeedPoetDrill = factory();
})(typeof self !== 'undefined' ? self : this, function () {
  'use strict';

  var MAX_LEVEL = 3;

  // randInt(n) returns an integer in [0, n). Defaults to crypto.getRandomValues.
  function cryptoInt(n) {
    var c = (typeof globalThis !== 'undefined' && globalThis.crypto) || null;
    if (!c || !c.getRandomValues) throw new Error('crypto.getRandomValues is not available');
    var buf = new Uint32Array(1);
    var limit = Math.floor(0x100000000 / n) * n;
    do { c.getRandomValues(buf); } while (buf[0] >= limit);
    return buf[0] % n;
  }

  // blanks picks `level` word positions per line (1..3) and returns, per line,
  // a sorted array of the hidden word indexes.
  function blanks(lines, level, randInt) {
    randInt = randInt || cryptoInt;
    level = Math.max(1, Math.min(MAX_LEVEL, level | 0));
    return lines.map(function (line) {
      var idx = line.map(function (_, i) { return i; });
      for (var i = idx.length - 1; i > 0; i--) {
        var j = randInt(i + 1);
        var t = idx[i]; idx[i] = idx[j]; idx[j] = t;
      }
      return idx.slice(0, Math.min(level, line.length)).sort(function (a, b) { return a - b; });
    });
  }

  function normalize(s) {
    return String(s == null ? '' : s).trim().toLowerCase();
  }

  // grade compares answers (same shape as the blanks: per line, one string per
  // hidden word) with the poem. Returns {right, total, ok} where ok mirrors answers.
  function grade(lines, hidden, answers) {
    var right = 0, total = 0;
    var ok = hidden.map(function (pos, li) {
      return pos.map(function (wi, k) {
        total++;
        var good = normalize(answers[li] && answers[li][k]) === lines[li][wi];
        if (good) right++;
        return good;
      });
    });
    return { right: right, total: total, ok: ok };
  }

  // next is the level after a graded round: up after a perfect round, else the same.
  function next(level, result) {
    return result.right === result.total ? Math.min(MAX_LEVEL, level + 1) : level;
  }

  return { MAX_LEVEL: MAX_LEVEL, blanks: blanks, normalize: normalize, grade: grade, next: next };
});
