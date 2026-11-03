// Seed Poet service worker: makes the site work offline and installable.
// It caches only the site's own static files. Poems are never fetched, so
// there is nothing secret to cache. POST requests (no-JS server poems) pass through.
'use strict';
const CACHE = 'seedpoet-2.1.0-8e2a1c45';
const URLS = ["/","/about","/assets/app.94382dfcaf.js","/assets/data.6d82e27841.js","/assets/pdf.3b53484cde.js","/assets/seedpoet.60de8fb12b.js","/assets/site.335c2f3a01.css","/assets/site.8bce1e819a.js","/expired","/favicon.ico","/favicon.svg","/icons/icon-192.png","/manifest.webmanifest"];

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
