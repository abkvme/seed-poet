'use strict';
// Tests for the website's drill (quiz) logic in web/static/js/drill.js.
const test = require('node:test');
const assert = require('node:assert');
const Drill = require('../../web/static/js/drill.js');

const poem = [
  ['abandon', 'ability', 'able'],
  ['about', 'above', 'absent'],
  ['absorb', 'abstract', 'absurd'],
  ['abuse', 'access', 'accident'],
];

test('blanks hides exactly `level` distinct words per line, sorted', () => {
  for (let level = 1; level <= 3; level++) {
    for (let run = 0; run < 50; run++) {
      const h = Drill.blanks(poem, level);
      assert.strictEqual(h.length, poem.length);
      h.forEach((pos) => {
        assert.strictEqual(pos.length, level);
        assert.strictEqual(new Set(pos).size, level);
        pos.forEach((p) => assert.ok(p >= 0 && p < 3));
        assert.deepStrictEqual(pos, pos.slice().sort((a, b) => a - b));
      });
    }
  }
});

test('blanks clamps the level to 1..3', () => {
  assert.ok(Drill.blanks(poem, 0).every((p) => p.length === 1));
  assert.ok(Drill.blanks(poem, 9).every((p) => p.length === 3));
});

test('blanks uses crypto by default and covers every position', () => {
  const seen = new Set();
  for (let i = 0; i < 200; i++) Drill.blanks(poem, 1).forEach((p) => seen.add(p[0]));
  assert.deepStrictEqual([...seen].sort(), [0, 1, 2]);
});

test('blanks takes an injected randInt (deterministic)', () => {
  const h = Drill.blanks(poem, 1, () => 0);
  // Fisher-Yates with j=0 always: [0,1,2] -> [1,2,0] -> [2,1,0]... first = 1 then sorted.
  assert.strictEqual(h.length, 4);
  assert.ok(h.every((p) => p.length === 1));
});

test('grade counts right answers, case and space insensitive', () => {
  const hidden = [[0], [2], [1], [0, 2]];
  const answers = [['  Abandon '], ['absent'], ['wrong'], ['abuse', '']];
  const r = Drill.grade(poem, hidden, answers);
  assert.strictEqual(r.total, 5);
  assert.strictEqual(r.right, 3);
  assert.deepStrictEqual(r.ok, [[true], [true], [false], [true, false]]);
});

test('grade treats missing answers as wrong', () => {
  const r = Drill.grade(poem, [[0], [0], [0], [0]], []);
  assert.strictEqual(r.right, 0);
  assert.strictEqual(r.total, 4);
});

test('next levels up only after a perfect round, capped at MAX_LEVEL', () => {
  assert.strictEqual(Drill.next(1, { right: 4, total: 4 }), 2);
  assert.strictEqual(Drill.next(2, { right: 3, total: 8 }), 2);
  assert.strictEqual(Drill.next(3, { right: 12, total: 12 }), Drill.MAX_LEVEL);
});
