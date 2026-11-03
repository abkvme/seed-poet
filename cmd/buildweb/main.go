// Command buildweb builds the Seed Poet website into web/dist:
//
//	go run ./cmd/buildweb
//
// Output is deterministic (same sources, same bytes), so CI can rebuild and
// fail on any difference from the committed web/dist.
package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/abkvme/seed-poet/v2/data"
	"github.com/abkvme/seed-poet/v2/seedpoet"
	"github.com/abkvme/seed-poet/v2/web"
)

const dist = "web/dist"

func main() {
	log.SetFlags(0)
	if _, err := os.Stat("go.mod"); err != nil {
		log.Fatal("run from the repository root: go run ./cmd/buildweb")
	}
	must(os.RemoveAll(dist))
	must(os.MkdirAll(filepath.Join(dist, "assets"), 0o755))

	// 1. Sources that become fingerprinted assets.
	sources := map[string][]byte{
		"css/site.css":   read("web/static/css/site.css"),
		"js/app.js":      read("web/static/js/app.js"),
		"js/site.js":     read("web/static/js/site.js"),
		"js/pdf.js":      read("web/static/js/pdf.js"),
		"js/seedpoet.js": read("js/seedpoet.js"),
		"js/data.js":     dataJS(),
	}
	assets := map[string]web.Asset{}
	var names []string
	for name := range sources {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		b := sources[name]
		sum := sha256.Sum256(b)
		ext := path.Ext(name)
		file := strings.TrimSuffix(path.Base(name), ext) + "." + hex.EncodeToString(sum[:])[:10] + ext
		sri := sha512.Sum384(b)
		assets[name] = web.Asset{Path: "/assets/" + file, SRI: "sha384-" + base64.StdEncoding.EncodeToString(sri[:])}
		write(filepath.Join(dist, "assets", file), b)
	}
	aj, _ := json.MarshalIndent(assets, "", "  ")
	write(filepath.Join(dist, "assets.json"), append(aj, '\n'))

	// 2. Pages.
	r, err := web.NewRenderer(assets, func(name string) (string, error) {
		b, ok := sources[name]
		if !ok {
			return "", fmt.Errorf("no source %q", name)
		}
		return string(b), nil
	})
	must(err)
	var sitemap []string
	var pageURLs []string
	for _, p := range web.Pages() {
		var buf bytes.Buffer
		if err := r.Render(&buf, p); err != nil {
			log.Fatalf("%s: %v", p.File, err)
		}
		write(filepath.Join(dist, p.File), buf.Bytes())
		if p.InSitemap {
			sitemap = append(sitemap, p.URL)
		}
		if !p.Offline && p.Path != "/404" {
			pageURLs = append(pageURLs, p.Path)
		}
	}

	// 3. Static root files (icons, social image, screenshots).
	must(fs.WalkDir(os.DirFS("web/static/root"), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		write(filepath.Join(dist, p), read(filepath.Join("web/static/root", p)))
		return nil
	}))

	// 4. Manifest, service worker, robots, sitemap, llms.txt.
	write(filepath.Join(dist, "manifest.webmanifest"), manifest())
	write(filepath.Join(dist, "sw.js"), serviceWorker(assets, pageURLs))
	write(filepath.Join(dist, "robots.txt"), []byte(robots))
	write(filepath.Join(dist, "sitemap.xml"), sitemapXML(sitemap))
	write(filepath.Join(dist, "llms.txt"), []byte(llms()))
	log.Printf("built %s: %d assets, %d pages", dist, len(assets), len(web.Pages()))
}

// dataJS is the compact browser form of data/: the BIP39 list as one string and
// every category as indices into it.
func dataJS() []byte {
	idx := map[string]int{}
	for i, w := range data.BIP39 {
		idx[w] = i
	}
	ids := func(ws []string) []int {
		out := make([]int, len(ws))
		for i, w := range ws {
			out[i] = idx[w]
		}
		return out
	}
	rhymes := make([][]any, len(data.Rhymes))
	for i, g := range data.Rhymes {
		rhymes[i] = []any{g.Key, ids(g.Words)}
	}
	d := map[string]any{
		"bip39":      strings.Join(data.BIP39, " "),
		"verbs":      ids(data.Verbs),
		"adjectives": ids(data.Adjectives),
		"nouns":      ids(data.Nouns),
		"rhymes":     rhymes,
	}
	b, err := json.Marshal(d)
	must(err)
	return []byte("/*! Seed Poet word data (BIP39 English + categories + rhymes), generated from data/ */\n" +
		"globalThis.SeedPoetData=" + string(b) + ";\n")
}

func manifest() []byte {
	m := map[string]any{
		"id":               "/",
		"name":             "Seed Poet",
		"short_name":       "Seed Poet",
		"description":      "Your BIP39 seed phrase as a rhyming poem, written on your device. Works offline.",
		"start_url":        "/?source=pwa",
		"scope":            "/",
		"display":          "standalone",
		"display_override": []string{"standalone", "minimal-ui"},
		"orientation":      "any",
		"background_color": "#020a05",
		"theme_color":      "#020a05",
		"lang":             "en",
		"dir":              "ltr",
		"categories":       []string{"security", "utilities", "finance"},
		"icons": []map[string]string{
			{"src": "/icons/icon-192.png", "sizes": "192x192", "type": "image/png", "purpose": "any"},
			{"src": "/icons/icon-512.png", "sizes": "512x512", "type": "image/png", "purpose": "any"},
			{"src": "/icons/maskable-512.png", "sizes": "512x512", "type": "image/png", "purpose": "maskable"},
			{"src": "/favicon.svg", "sizes": "any", "type": "image/svg+xml", "purpose": "any"},
		},
		"screenshots": []map[string]string{
			{"src": "/screenshots/wide.png", "sizes": "1280x720", "type": "image/png", "form_factor": "wide", "label": "Seed Poet home"},
			{"src": "/screenshots/narrow.png", "sizes": "412x915", "type": "image/png", "form_factor": "narrow", "label": "Seed Poet on a phone"},
		},
		"shortcuts": []map[string]any{
			{"name": "12-word poem", "url": "/?words=12", "icons": []map[string]string{{"src": "/icons/icon-192.png", "sizes": "192x192"}}},
			{"name": "24-word poem", "url": "/?words=24", "icons": []map[string]string{{"src": "/icons/icon-192.png", "sizes": "192x192"}}},
		},
	}
	b, err := json.MarshalIndent(m, "", "  ")
	must(err)
	return append(b, '\n')
}

func serviceWorker(assets map[string]web.Asset, pages []string) []byte {
	var urls []string
	urls = append(urls, pages...)
	for _, a := range assets {
		urls = append(urls, a.Path)
	}
	urls = append(urls, "/favicon.svg", "/favicon.ico", "/icons/icon-192.png", "/manifest.webmanifest")
	sort.Strings(urls)
	list, _ := json.Marshal(urls)
	sum := sha256.Sum256(list)
	version := seedpoet.Version + "-" + hex.EncodeToString(sum[:])[:8]
	return []byte(`// Seed Poet service worker: makes the site work offline and installable.
// It caches only the site's own static files. Poems are never fetched, so
// there is nothing secret to cache. POST requests (no-JS server poems) pass through.
'use strict';
const CACHE = 'seedpoet-` + version + `';
const URLS = ` + string(list) + `;

self.addEventListener('install', (e) => {
  e.waitUntil(caches.open(CACHE).then((c) => c.addAll(URLS)).then(() => self.skipWaiting()));
});

self.addEventListener('activate', (e) => {
  e.waitUntil(caches.keys()
    .then((keys) => Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k))))
    .then(() => self.clients.claim()));
});

const PRECACHED = new Set(URLS);

self.addEventListener('fetch', (e) => {
  const req = e.request;
  if (req.method !== 'GET') return;
  const url = new URL(req.url);
  if (url.origin !== self.location.origin) return;
  if (req.mode === 'navigate') {
    // Network first so updates show at once; the cached page when offline.
    e.respondWith(fetch(req).catch(() =>
      caches.match(url.pathname).then((r) => r || caches.match('/')).then((r) => r || Response.error())));
    return;
  }
  // Only the site's own precached files are served from the cache; anything else
  // goes to the network untouched (the worker does not intercept it at all).
  if (!PRECACHED.has(url.pathname)) return;
  e.respondWith(caches.match(url.pathname).then((hit) => hit || fetch(req).catch(() => Response.error())));
});
`)
}

const robots = `# Seed Poet: https://seedpoet.com
User-agent: *
Allow: /
Disallow: /nojs/
Disallow: /expired

Sitemap: https://seedpoet.com/sitemap.xml
`

func sitemapXML(urls []string) []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n" + `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, u := range urls {
		pri := "0.8"
		if strings.HasSuffix(u, "/") {
			pri = "1.0"
		}
		fmt.Fprintf(&b, "  <url>\n    <loc>%s</loc>\n    <changefreq>monthly</changefreq>\n    <priority>%s</priority>\n  </url>\n", u, pri)
	}
	b.WriteString("</urlset>\n")
	return []byte(b.String())
}

func llms() string {
	return `# Seed Poet

> Seed Poet generates a valid BIP39 seed phrase (12 or 24 words) written as rhyming couplets, so it is easier to remember. The poem is generated in the user's browser and never sent or stored.

Seed Poet is free and open source (MIT), developed by OxiSoft (https://oxisoft.io).

## Facts
- Every poem, read word by word, is a valid BIP39 English mnemonic with a correct checksum; any BIP39 wallet accepts it.
- Lines follow a VERB + ADJECTIVE + NOUN pattern; the nouns of each couplet rhyme and the lines have matching syllable counts.
- 12 words = 4 lines (2 couplets); 24 words = 8 lines (4 couplets).
- Randomness: crypto.getRandomValues in the browser (crypto/rand in Go).
- Works with any wallet that imports a standard BIP39 English phrase: Bitcoin, Litecoin, Dogecoin, Bitcoin Cash, Ethereum and EVM networks (Polygon, BNB Chain, Arbitrum, Optimism, Base, Avalanche C-Chain), Solana, Cardano (24 words), Polkadot, Cosmos and Tron; wallets such as Ledger, Trezor, MetaMask, Trust Wallet, Phantom and Exodus. Not Monero (own 25-word format).
- Strength estimate: about 88 bits for 12 words and 176 bits for 24 words (a fully random BIP39 phrase has 128 and 256), because rhyme and grammar limit word choice. 24 words are recommended for real value.
- Privacy: the poem page may not make network requests (CSP connect-src 'none'); no cookies, analytics or third-party code; the poem is wiped after 15 minutes or when the page is left.
- Without JavaScript, the server can generate a poem instead; it is shown once and never stored or logged.
- Works offline, installs as an app (PWA), and has a single-file offline version.
- Two identical implementations: Go (library + CLI) and dependency-free JavaScript; the same random bytes give the same poem.
- Version ` + seedpoet.Version + `.

## Links
- Website: https://seedpoet.com
- About: https://seedpoet.com/about
- Offline version: https://seedpoet.com/seedpoet-offline.html
- Source code: https://github.com/abkvme/seed-poet
- Developer: https://oxisoft.io
`
}

func read(p string) []byte {
	b, err := os.ReadFile(p)
	must(err)
	return b
}

func write(p string, b []byte) {
	must(os.MkdirAll(filepath.Dir(p), 0o755))
	must(os.WriteFile(p, b, 0o644))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
