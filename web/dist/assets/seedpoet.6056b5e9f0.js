/*! Seed Poet JavaScript library | MIT | https://github.com/abkvme/seed-poet
 *
 * Generates BIP39 seed phrases written as rhyming couplets. A line-for-line port of
 * the Go library (seedpoet/, internal/): from the same random bytes both produce the
 * same poem (checked against testdata/vectors.json). No dependencies; runs in
 * browsers and in Node.js (>= 19, for globalThis.crypto).
 *
 *   var gen = SeedPoet.createGenerator(data);   // data: see normalizeData
 *   var poem = gen.generate(24);                // crypto.getRandomValues
 *   poem.lines   -> [["verb","adj","noun"], ...]
 *   poem.words() -> 24-word BIP39 mnemonic
 *   SeedPoet.validate(words) -> null if valid, else an error message
 */
(function (root, factory) {
  'use strict';
  var api = factory();
  if (typeof module === 'object' && module.exports) module.exports = api;
  else root.SeedPoet = api;
})(typeof globalThis !== 'undefined' ? globalThis : this, function () {
  'use strict';

  // ---------- SHA-256 (FIPS 180-4), synchronous, for the BIP39 checksum ----------

  var K = new Uint32Array([
    0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
    0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
    0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
    0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
    0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
    0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
    0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
    0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2
  ]);

  function sha256(msg) {
    var len = msg.length;
    var padded = new Uint8Array(((len + 9 + 63) >> 6) << 6);
    padded.set(msg);
    padded[len] = 0x80;
    var bitsHi = Math.floor(len / 0x20000000), bitsLo = (len << 3) >>> 0;
    var end = padded.length;
    padded[end - 8] = bitsHi >>> 24; padded[end - 7] = bitsHi >>> 16;
    padded[end - 6] = bitsHi >>> 8; padded[end - 5] = bitsHi;
    padded[end - 4] = bitsLo >>> 24; padded[end - 3] = bitsLo >>> 16;
    padded[end - 2] = bitsLo >>> 8; padded[end - 1] = bitsLo;

    var h = new Uint32Array([0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
      0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19]);
    var w = new Uint32Array(64);
    for (var off = 0; off < end; off += 64) {
      for (var i = 0; i < 16; i++) {
        var p = off + i * 4;
        w[i] = (padded[p] << 24) | (padded[p + 1] << 16) | (padded[p + 2] << 8) | padded[p + 3];
      }
      for (i = 16; i < 64; i++) {
        var x = w[i - 15], y = w[i - 2];
        var s0 = ((x >>> 7) | (x << 25)) ^ ((x >>> 18) | (x << 14)) ^ (x >>> 3);
        var s1 = ((y >>> 17) | (y << 15)) ^ ((y >>> 19) | (y << 13)) ^ (y >>> 10);
        w[i] = (w[i - 16] + s0 + w[i - 7] + s1) | 0;
      }
      var a = h[0], b = h[1], c = h[2], d = h[3], e = h[4], f = h[5], g = h[6], hh = h[7];
      for (i = 0; i < 64; i++) {
        var S1 = ((e >>> 6) | (e << 26)) ^ ((e >>> 11) | (e << 21)) ^ ((e >>> 25) | (e << 7));
        var t1 = (hh + S1 + ((e & f) ^ (~e & g)) + K[i] + w[i]) | 0;
        var S0 = ((a >>> 2) | (a << 30)) ^ ((a >>> 13) | (a << 19)) ^ ((a >>> 22) | (a << 10));
        var t2 = (S0 + ((a & b) ^ (a & c) ^ (b & c))) | 0;
        hh = g; g = f; f = e; e = (d + t1) | 0; d = c; c = b; b = a; a = (t1 + t2) | 0;
      }
      h[0] += a; h[1] += b; h[2] += c; h[3] += d; h[4] += e; h[5] += f; h[6] += g; h[7] += hh;
    }
    var out = new Uint8Array(32);
    for (i = 0; i < 8; i++) {
      out[i * 4] = h[i] >>> 24; out[i * 4 + 1] = h[i] >>> 16;
      out[i * 4 + 2] = h[i] >>> 8; out[i * 4 + 3] = h[i];
    }
    return out;
  }

  // ---------- Random numbers: the same rules as internal/rng (Go) ----------
  // Uint32 = next 4 bytes big-endian; intN = rejection sampling; shuffle = Fisher-Yates
  // from the high index down.

  function RNG(fill) {
    this._fill = fill;
    this._buf = new Uint8Array(256);
    this._pos = 0;
    this._n = 0;
  }
  RNG.prototype.uint32 = function () {
    var v = 0;
    for (var i = 0; i < 4; i++) {
      if (this._pos === this._n) {
        this._fill(this._buf);
        this._pos = 0;
        this._n = this._buf.length;
      }
      v = v * 256 + this._buf[this._pos++];
    }
    return v;
  };
  RNG.prototype.intN = function (n) {
    if (!(n > 0)) throw new Error('rng: intN with n <= 0');
    if (n === 1) return 0;
    var span = 4294967296, limit = span - (span % n);
    for (;;) {
      var x = this.uint32();
      if (x < limit) return x % n;
    }
  };
  RNG.prototype.shuffle = function (arr) {
    for (var i = arr.length - 1; i > 0; i--) {
      var j = this.intN(i + 1), t = arr[i];
      arr[i] = arr[j];
      arr[j] = t;
    }
  };
  RNG.prototype.wipe = function () {
    this._buf.fill(0);
    this._pos = this._n = 0;
  };

  function cryptoRNG() {
    var c = typeof globalThis !== 'undefined' && globalThis.crypto;
    if (!c || !c.getRandomValues) throw new Error('No secure random source (crypto.getRandomValues)');
    return new RNG(function (buf) { c.getRandomValues(buf); });
  }

  function utf8(s) {
    if (typeof s !== 'string') return s;
    if (typeof TextEncoder !== 'undefined') return new TextEncoder().encode(s);
    return new Uint8Array(Buffer.from(s, 'utf8'));
  }

  // Reproducible stream for tests only: block k = SHA-256(seed || uint32_be(k)).
  function deterministicRNG(seed) {
    seed = utf8(seed);
    var block = 0, pending = new Uint8Array(0), at = 0;
    return new RNG(function (buf) {
      for (var i = 0; i < buf.length; i++) {
        if (at === pending.length) {
          var input = new Uint8Array(seed.length + 4);
          input.set(seed);
          input[seed.length] = block >>> 24; input[seed.length + 1] = block >>> 16;
          input[seed.length + 2] = block >>> 8; input[seed.length + 3] = block;
          pending = sha256(input);
          at = 0;
          block = (block + 1) >>> 0;
        }
        buf[i] = pending[at++];
      }
    });
  }

  // ---------- Data ----------
  // Accepts words or indices into bip39 (the compact browser form). bip39 may be an
  // array or a space-separated string.
  function normalizeData(d) {
    var list = typeof d.bip39 === 'string' ? d.bip39.split(' ') : d.bip39.slice();
    if (list.length !== 2048) throw new Error('data: BIP39 list must have 2048 words');
    function words(xs) {
      return xs.map(function (x) { return typeof x === 'number' ? list[x] : x; });
    }
    return {
      bip39: list,
      verbs: words(d.verbs),
      adjectives: words(d.adjectives),
      nouns: words(d.nouns),
      rhymes: d.rhymes.map(function (g) { return [g[0], words(g[1])]; })
    };
  }

  // ---------- BIP39 checksum ----------

  function makeValidator(list) {
    var index = new Map();
    list.forEach(function (w, i) { index.set(w, i); });
    return function validate(words) {
      var n = words.length;
      if (n < 12 || n > 24 || n % 3 !== 0) return 'bip39: unsupported word count ' + n;
      var totalBits = n * 11, csBits = totalBits / 33, entBits = totalBits - csBits;
      var bits = new Uint8Array(totalBits);
      for (var i = 0; i < n; i++) {
        var idx = index.get(words[i]);
        if (idx === undefined) return 'bip39: "' + words[i] + '" is not in the wordlist';
        for (var b = 10; b >= 0; b--) bits[i * 11 + (10 - b)] = (idx >> b) & 1;
      }
      var entropy = new Uint8Array(entBits / 8);
      for (i = 0; i < entBits; i++) entropy[i >> 3] |= bits[i] << (7 - (i & 7));
      var sum = sha256(entropy);
      for (i = 0; i < csBits; i++) {
        if (((sum[i >> 3] >> (7 - (i & 7))) & 1) !== bits[entBits + i]) return 'bip39: checksum mismatch';
      }
      return null;
    };
  }

  // ---------- Syllables: same heuristic as internal/syllable ----------

  function countSyllables(word) {
    word = word.toLowerCase();
    if (word.length === 0) return 0;
    var vowels = 'aeiouy', count = 0, prev = false;
    for (var i = 0; i < word.length; i++) {
      var r = word[i], isVowel = vowels.indexOf(r) >= 0;
      if (isVowel && !prev) count++;
      if (r === 'y' && i === 0) {
        count--;
        if (count < 0) count = 0;
      }
      prev = isVowel;
    }
    if (word.length > 2 && word.slice(-1) === 'e' && word.slice(-2) !== 'le') count--;
    if (word.length > 2 && word.slice(-2) === 'ed') {
      var c = word[word.length - 3];
      if (c !== 't' && c !== 'd') count--;
    }
    return count < 1 ? 1 : count;
  }

  // ---------- Generator: same algorithm as internal/generator ----------

  var MAX_ATTEMPTS = 10000;

  function groupBy(list, syl) {
    var m = new Map();
    list.forEach(function (w) {
      var s = syl(w);
      if (!m.has(s)) m.set(s, []);
      m.get(s).push(w);
    });
    return m;
  }

  function sortedKeys(m) {
    return Array.from(m.keys()).sort(function (a, b) { return a - b; });
  }

  function Poem(lines, wordCount) {
    this.lines = lines;
    this.wordCount = wordCount;
  }
  Poem.prototype.words = function () {
    return this.lines.reduce(function (acc, l) { return acc.concat(l); }, []);
  };
  Poem.prototype.format = function () {
    return this.lines.map(function (l) { return l.join(' '); }).join('\n');
  };

  function createGenerator(rawData) {
    var d = normalizeData(rawData);
    var validate = makeValidator(d.bip39);

    // Missing words count as 0 syllables, exactly like the Go map lookup.
    var syllables = new Map();
    d.verbs.concat(d.adjectives, d.nouns).forEach(function (w) { syllables.set(w, countSyllables(w)); });
    function syl(w) { return syllables.has(w) ? syllables.get(w) : 0; }

    var verbsBySyl = groupBy(d.verbs, syl);
    var adjsBySyl = groupBy(d.adjectives, syl);
    var combo = new Map();
    sortedKeys(verbsBySyl).forEach(function (vs) {
      sortedKeys(adjsBySyl).forEach(function (as) {
        var total = vs + as;
        verbsBySyl.get(vs).forEach(function (v) {
          adjsBySyl.get(as).forEach(function (a) {
            if (v !== a) {
              if (!combo.has(total)) combo.set(total, []);
              combo.get(total).push([v, a]);
            }
          });
        });
      });
    });

    var groups = new Map();
    var keys = [];
    d.rhymes.forEach(function (g) {
      groups.set(g[0], g[1]);
      if (g[1].length >= 2) keys.push(g[0]);
    });
    keys.sort(function (a, b) { return a < b ? -1 : a > b ? 1 : 0; });
    keys.sort(function (a, b) { return groups.get(b).length - groups.get(a).length; }); // stable

    function pickUnusedFrom(r, list, used) {
      var n = list.length;
      for (var t = 0; t < 50; t++) {
        var w = list[r.intN(n)];
        if (!used.has(w)) return w;
      }
      for (var i = 0; i < n; i++) if (!used.has(list[i])) return list[i];
      return list[r.intN(n)];
    }

    function pickVerbAndAdj(r, target, used) {
      var targets = [target, target - 1, target + 1];
      for (var k = 0; k < targets.length; k++) {
        var pairs = combo.get(targets[k]);
        if (pairs && pairs.length > 0) {
          for (var t = 0; t < 20; t++) {
            var p = pairs[r.intN(pairs.length)];
            if (!used.has(p[0]) && !used.has(p[1])) return p;
          }
        }
      }
      var v = pickUnusedFrom(r, d.verbs, used);
      var a = pickUnusedFrom(r, d.adjectives, used);
      return [v, a];
    }

    function pickTwoRhyming(r, group, used) {
      var available = group.filter(function (w) { return !used.has(w); });
      if (available.length < 2) return available;
      r.shuffle(available);
      var bySyl = groupBy(available, syl);
      var valid = sortedKeys(bySyl).filter(function (c) { return bySyl.get(c).length >= 2; });
      if (valid.length > 0) {
        var words = bySyl.get(valid[r.intN(valid.length)]);
        r.shuffle(words);
        return words.slice(0, 2);
      }
      return available.slice(0, 2);
    }

    function pickTwoFromGroup(r, group) {
      if (group.length < 2) return group;
      var s = group.slice();
      r.shuffle(s);
      return [s[0], s[1]];
    }

    function generateOnce(r, wordCount) {
      var couplets = wordCount === 24 ? 4 : 2;
      var used = new Set();
      var lines = [];
      var validKeys = keys.slice();
      if (validKeys.length === 0) return new Poem([], 0);
      while (validKeys.length < couplets) validKeys = validKeys.concat(validKeys);
      r.shuffle(validKeys);
      validKeys.slice(0, couplets).forEach(function (key) {
        var group = groups.get(key);
        if (group.length < 2) return;
        var rw = pickTwoRhyming(r, group, used);
        if (rw.length < 2) rw = pickTwoFromGroup(r, group);
        used.add(rw[0]);
        used.add(rw[1]);

        var v1 = pickUnusedFrom(r, d.verbs, used);
        used.add(v1);
        var a1 = pickUnusedFrom(r, d.adjectives, used);
        used.add(a1);
        lines.push([v1, a1, rw[0]]);

        used.add(rw[1]);
        var target = syl(rw[0]) + syl(v1) + syl(a1) - syl(rw[1]);
        var pair = pickVerbAndAdj(r, target, used);
        used.add(pair[0]);
        used.add(pair[1]);
        lines.push([pair[0], pair[1], rw[1]]);
      });
      return new Poem(lines, wordCount);
    }

    return {
      // generate(12 | 24, [rng]) -> Poem with a valid BIP39 checksum.
      generate: function (wordCount, rng) {
        if (wordCount !== 12 && wordCount !== 24) wordCount = 12;
        var own = !rng;
        var r = rng || cryptoRNG();
        try {
          for (var i = 0; i < MAX_ATTEMPTS; i++) {
            var p = generateOnce(r, wordCount);
            if (validate(p.words()) === null) return p;
          }
        } finally {
          if (own) r.wipe();
        }
        throw new Error('generator: no valid BIP39 poem found');
      },
      validate: validate,
      wordlist: function () { return d.bip39.slice(); }
    };
  }

  return {
    version: '2.1.2',
    createGenerator: createGenerator,
    normalizeData: normalizeData,
    cryptoRNG: cryptoRNG,
    deterministicRNG: deterministicRNG,
    countSyllables: countSyllables,
    sha256: sha256,
    validate: function (data, words) { return makeValidator(normalizeData(data).bip39)(words); },
    RNG: RNG
  };
});
