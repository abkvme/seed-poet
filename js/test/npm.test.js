'use strict';
// The npm package (js/npm/stage.js): staged files load, carry the bundled data and
// write the same poems as the Go library.
const test = require('node:test');
const assert = require('node:assert');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const vm = require('node:vm');
const { execFileSync } = require('node:child_process');

const root = path.join(__dirname, '..', '..');
const vectors = JSON.parse(fs.readFileSync(path.join(root, 'testdata', 'vectors.json'), 'utf8'));
const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'seedpoet-npm-'));
execFileSync(process.execPath, [path.join(root, 'js', 'npm', 'stage.js'), dir]);
test.after(() => fs.rmSync(dir, { recursive: true, force: true }));

test('package.json: name, version, files', () => {
  const pkg = JSON.parse(fs.readFileSync(path.join(dir, 'package.json'), 'utf8'));
  const lib = JSON.parse(fs.readFileSync(path.join(root, 'js', 'package.json'), 'utf8'));
  assert.strictEqual(pkg.name, 'seedpoet');
  assert.strictEqual(pkg.version, lib.version);
  for (const f of pkg.files.concat(['README.md', 'LICENSE'])) assert.ok(fs.existsSync(path.join(dir, f)), f);
});

test('index.js: generate + validate with bundled data', () => {
  const S = require(path.join(dir, 'index.js'));
  for (const n of [12, 24]) {
    const p = S.generate(n);
    assert.strictEqual(p.words().length, n);
    assert.strictEqual(S.validateWords(p.words()), null);
  }
  assert.notStrictEqual(S.validateWords('zoo '.repeat(12).trim().split(' ')), null);
});

test('bundled data writes the Go vectors', () => {
  const S = require(path.join(dir, 'index.js'));
  const gen = S.createGenerator(S.data);
  for (const v of vectors.poems) {
    const p = gen.generate(v.words, S.deterministicRNG(v.seed));
    assert.deepStrictEqual(p.format().split('\n'), v.lines, v.seed + '/' + v.words);
  }
});

test('seedpoet.full.js works as a browser script (global SeedPoet)', () => {
  const ctx = { crypto: globalThis.crypto, Uint8Array, Uint32Array, Map, Math, Error };
  ctx.globalThis = ctx;
  vm.createContext(ctx);
  vm.runInContext(fs.readFileSync(path.join(dir, 'seedpoet.full.js'), 'utf8'), ctx);
  const p = ctx.SeedPoet.generate(12);
  assert.strictEqual(p.words().length, 12);
  assert.strictEqual(ctx.SeedPoet.validateWords(p.words()), null);
});
