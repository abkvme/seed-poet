// Seed Poet service worker: makes the site work offline and installable.
// It caches only the site's own static files. Poems are never fetched, so
// there is nothing secret to cache. POST requests (no-JS server poems) pass through.
'use strict';
const CACHE = 'seedpoet-2.0.0-e0150cd6';
const URLS = ["/","/12","/24","/about","/assets/app.6186cf0fc4.js","/assets/data.6d82e27841.js","/assets/pdf.3b53484cde.js","/assets/seedpoet.7e9db08f38.js","/assets/site.41c80ded3a.js","/assets/site.59a02ef349.css","/expired","/favicon.ico","/favicon.svg","/icons/icon-192.png","/manifest.webmanifest"];

self.addEventListener('install', (e) => {
  e.waitUntil(caches.open(CACHE).then((c) => c.addAll(URLS)).then(() => self.skipWaiting()));
});

self.addEventListener('activate', (e) => {
  e.waitUntil(caches.keys()
    .then((keys) => Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k))))
    .then(() => self.clients.claim()));
});

self.addEventListener('fetch', (e) => {
  const req = e.request;
  if (req.method !== 'GET') return;
  const url = new URL(req.url);
  if (url.origin !== self.location.origin) return;
  if (req.mode === 'navigate') {
    // Network first so updates show at once; the cached page when offline.
    e.respondWith(fetch(req).catch(() =>
      caches.match(url.pathname).then((r) => r || caches.match('/'))));
    return;
  }
  e.respondWith(caches.match(req).then((r) => r || fetch(req)));
});
