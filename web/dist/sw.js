// Seed Poet service worker: makes the site work offline and installable.
// It caches only the site's own static files. Poems are never fetched, so
// there is nothing secret to cache. POST requests (no-JS server poems) pass through.
'use strict';
const CACHE = 'seedpoet-2.2.0-c7f62939';
const URLS = ["/","/about","/assets/app.e2322b4be5.js","/assets/data.6d82e27841.js","/assets/drill.d5a4dfcecd.js","/assets/pdf.3b53484cde.js","/assets/seedpoet.0e3d83a3e9.js","/assets/site.8bce1e819a.js","/assets/site.f882ddd009.css","/es/about","/es/guide/memorize-seed-phrase","/expired","/favicon.ico","/favicon.svg","/guide/memorize-seed-phrase","/icons/icon-192.png","/manifest.webmanifest","/pt/about","/pt/guide/memorize-seed-phrase","/wordlist"];

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
