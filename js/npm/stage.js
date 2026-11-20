#!/usr/bin/env node
// Stages the npm package `seedpoet` into a directory (default bin/npm).
// The package ships the library, the word data and a loader, because
// js/seedpoet.js alone cannot write a poem without data/.
//
//   node js/npm/stage.js [outDir]
//   cd bin/npm && npm pack --dry-run && npm publish
'use strict';

const fs = require('node:fs');
const path = require('node:path');

const root = path.join(__dirname, '..', '..');
const out = path.resolve(process.argv[2] || path.join(root, 'bin', 'npm'));

const lib = JSON.parse(fs.readFileSync(path.join(root, 'js', 'package.json'), 'utf8'));
const core = fs.readFileSync(path.join(root, 'js', 'seedpoet.js'), 'utf8');
const json = JSON.parse(fs.readFileSync(path.join(root, 'data', 'seedpoet.json'), 'utf8'));
const bip39 = fs.readFileSync(path.join(root, 'data', 'english.txt'), 'utf8').split(/\s+/).filter(Boolean);
if (bip39.length !== 2048) throw new Error('data/english.txt: expected 2048 words');

// Compact form understood by normalizeData: bip39 as one string, word lists as indices.
const index = new Map(bip39.map((w, i) => [w, i]));
const ids = (xs) => xs.map((w) => {
  if (!index.has(w)) throw new Error('word not in BIP39 list: ' + w);
  return index.get(w);
});
const data = {
  bip39: bip39.join(' '),
  verbs: ids(json.verbs),
  adjectives: ids(json.adjectives),
  nouns: ids(json.nouns),
  rhymes: json.rhymes.map((g) => [g[0], ids(g[1])])
};
const dataJSON = JSON.stringify(data);

const addOn = (target, dataExpr) => `
;(function (S, d) {
  'use strict';
  var gen;
  function g() { return gen || (gen = S.createGenerator(d)); }
  S.data = d;
  S.generate = function (wordCount) { return g().generate(wordCount || 24); };
  S.validateWords = function (words) { return g().validate(words); };
})(${target}, ${dataExpr});
`;

const pkg = {
  name: 'seedpoet',
  version: lib.version,
  description: lib.description,
  main: 'index.js',
  browser: 'seedpoet.full.js',
  unpkg: 'seedpoet.full.js',
  jsdelivr: 'seedpoet.full.js',
  files: ['index.js', 'seedpoet.js', 'seedpoet.full.js', 'data.json'],
  engines: lib.engines,
  keywords: lib.keywords,
  homepage: lib.homepage,
  bugs: { url: 'https://github.com/abkvme/seed-poet/issues' },
  repository: { type: 'git', url: 'git+https://github.com/abkvme/seed-poet.git' },
  author: lib.author,
  license: lib.license
};

const indexJS = `'use strict';
// seedpoet: BIP39 seed phrases written as rhyming couplets. https://seedpoet.com
// The library (seedpoet.js) plus the bundled word data (data.json).
const SeedPoet = require('./seedpoet.js');
${addOn('SeedPoet', "require('./data.json')").trim()}
module.exports = SeedPoet;
`;

const fullJS = core.trimEnd() + '\n' + addOn(
  "(typeof module === 'object' && module.exports && module.exports.createGenerator) ? module.exports : globalThis.SeedPoet",
  dataJSON
);

const readme = `# seedpoet

BIP39 seed phrases written as rhyming couplets: a valid 12 or 24 word mnemonic that is easier
to read back, check and remember. Zero dependencies, runs in browsers and Node.js (19+), and
produces exactly the same poem as the Go library from the same random bytes.

Try it in the browser: [seedpoet.com](https://seedpoet.com) (the poem is written on your device,
nothing is sent anywhere).

## Install

\`\`\`bash
npm install seedpoet
\`\`\`

\`\`\`js
const SeedPoet = require('seedpoet');

const poem = SeedPoet.generate(24);      // crypto.getRandomValues
console.log(poem.format());              // 8 rhyming lines
const words = poem.words();              // the 24-word BIP39 mnemonic
console.log(SeedPoet.validateWords(words)); // null = valid BIP39
\`\`\`

## In the browser, one script tag

\`\`\`html
<script src="https://cdn.jsdelivr.net/npm/seedpoet@${lib.version.split('.')[0]}/seedpoet.full.js"></script>
<script>
  const poem = SeedPoet.generate(12);
  console.log(poem.format());
</script>
\`\`\`

For a wallet seed, prefer a copy you host or saved yourself (pin an exact version and add
\`integrity=\` with the hash from the release's SHA256SUMS) over a live CDN link.

## Files

- \`seedpoet.full.js\`: library with the word data built in (browser global \`SeedPoet\`).
- \`seedpoet.js\`: library only; pass the data yourself with \`SeedPoet.createGenerator(data)\`.
- \`data.json\`: the word data (BIP39 English list, categories, rhyme groups).
- \`index.js\`: Node.js entry, library plus data.

## Security

Randomness comes only from \`crypto.getRandomValues\`. The library logs, stores and sends
nothing. Every poem is checked against the BIP39 checksum before it is returned. Source, Go
library, CLI and issue tracker: [github.com/abkvme/seed-poet](https://github.com/abkvme/seed-poet).

MIT licensed. Made by [OxiSoft](https://oxisoft.io).
`;

fs.rmSync(out, { recursive: true, force: true });
fs.mkdirSync(out, { recursive: true });
fs.writeFileSync(path.join(out, 'package.json'), JSON.stringify(pkg, null, 2) + '\n');
fs.writeFileSync(path.join(out, 'seedpoet.js'), core);
fs.writeFileSync(path.join(out, 'seedpoet.full.js'), fullJS);
fs.writeFileSync(path.join(out, 'data.json'), dataJSON + '\n');
fs.writeFileSync(path.join(out, 'index.js'), indexJS);
fs.writeFileSync(path.join(out, 'README.md'), readme);
fs.copyFileSync(path.join(root, 'LICENSE'), path.join(out, 'LICENSE'));
console.log('staged seedpoet@' + pkg.version + ' in ' + out);
