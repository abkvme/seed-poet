'use strict';
// Run: node --test js/   (no dependencies)
const test = require('node:test');
const assert = require('node:assert');
const fs = require('node:fs');
const path = require('node:path');
const nodeCrypto = require('node:crypto');
const SeedPoet = require('../seedpoet.js');

const root = path.join(__dirname, '..', '..');
const json = JSON.parse(fs.readFileSync(path.join(root, 'data', 'seedpoet.json'), 'utf8'));
const bip39 = fs.readFileSync(path.join(root, 'data', 'english.txt'), 'utf8').split(/\s+/).filter(Boolean);
const data = { bip39, verbs: json.verbs, adjectives: json.adjectives, nouns: json.nouns, rhymes: json.rhymes };
const vectors = JSON.parse(fs.readFileSync(path.join(root, 'testdata', 'vectors.json'), 'utf8'));
const gen = SeedPoet.createGenerator(data);

test('sha256 matches node:crypto', () => {
  for (const len of [0, 1, 55, 56, 63, 64, 65, 119, 120, 1000]) {
    const msg = nodeCrypto.randomBytes(len);
    const want = nodeCrypto.createHash('sha256').update(msg).digest('hex');
    assert.strictEqual(Buffer.from(SeedPoet.sha256(new Uint8Array(msg))).toString('hex'), want, 'len ' + len);
  }
});

test('BIP39 reference vectors', () => {
  const valid = [
    'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about',
    'legal winner thank year wave sausage worth useful legal winner thank yellow',
    'letter advice cage absurd amount doctor acoustic avoid letter advice cage above',
    'zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong',
    'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art',
    'zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo vote',
    'ozone drill grab fiber curtain grace pudding thank cruise elder eight picnic',
    'hamster diagram private dutch cause delay private meat slide toddler razor book happy fancy gospel tennis maple dilemma loan word shrug inflict delay length'
  ];
  for (const m of valid) assert.strictEqual(gen.validate(m.split(' ')), null, m);
  const invalid = [
    'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon',
    'zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo',
    'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about',
    'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon notaword'
  ];
  for (const m of invalid) assert.notStrictEqual(gen.validate(m.split(' ')), null, m);
});

test('random number rules match Go (testdata/vectors.json rng)', () => {
  const r = SeedPoet.deterministicRNG(vectors.rng.seed);
  assert.deepStrictEqual(vectors.rng.uint32.map(() => r.uint32()), vectors.rng.uint32);
  for (const [n, want] of vectors.rng.intn) assert.strictEqual(r.intN(n), want, 'intN(' + n + ')');
  const xs = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11];
  r.shuffle(xs);
  assert.deepStrictEqual(xs, vectors.rng.shuffle);
});

test('poems match the Go library exactly', () => {
  assert.ok(vectors.poems.length >= 50);
  for (const v of vectors.poems) {
    const p = gen.generate(v.words, SeedPoet.deterministicRNG(v.seed));
    assert.deepStrictEqual(p.format().split('\n'), v.lines, v.seed + '/' + v.words);
  }
});

test('syllable counter matches Go tests', () => {
  for (const [w, n] of [['apple', 2], ['banana', 3], ['cat', 1], ['syllable', 3], ['generation', 4]]) {
    assert.strictEqual(SeedPoet.countSyllables(w), n, w);
  }
});

test('crypto poems are valid, rhyme and do not repeat', () => {
  const keysOf = new Map();
  for (const [k, ws] of data.rhymes) for (const w of ws) (keysOf.get(w) || keysOf.set(w, new Set()).get(w)).add(k);
  const rhymes = (a, b) => [...(keysOf.get(a) || [])].some((k) => (keysOf.get(b) || new Set()).has(k));
  for (const n of [12, 24]) {
    const seen = new Set();
    for (let i = 0; i < 300; i++) {
      const p = gen.generate(n);
      const words = p.words();
      assert.strictEqual(words.length, n);
      assert.strictEqual(gen.validate(words), null, p.format());
      for (let l = 0; l + 1 < p.lines.length; l += 2) assert.ok(rhymes(p.lines[l][2], p.lines[l + 1][2]), p.format());
      assert.ok(!seen.has(p.format()));
      seen.add(p.format());
    }
  }
});

test('compact browser data (indices) gives the same poems', () => {
  const idx = new Map(bip39.map((w, i) => [w, i]));
  const compact = {
    bip39: bip39.join(' '),
    verbs: json.verbs.map((w) => idx.get(w)),
    adjectives: json.adjectives.map((w) => idx.get(w)),
    nouns: json.nouns.map((w) => idx.get(w)),
    rhymes: json.rhymes.map(([k, ws]) => [k, ws.map((w) => idx.get(w))])
  };
  const g2 = SeedPoet.createGenerator(compact);
  for (const v of vectors.poems.slice(0, 10)) {
    assert.deepStrictEqual(g2.generate(v.words, SeedPoet.deterministicRNG(v.seed)).format().split('\n'), v.lines);
  }
});
