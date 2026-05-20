// Kill-switch service worker — replaces any stale SW left from a prior app
// at this origin (localhost:3000). On activation it unregisters itself,
// clears all Cache Storage entries, and forces clients to reload so their
// next requests bypass the dead worker entirely.
self.addEventListener('install', () => {
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    (async () => {
      try {
        const keys = await caches.keys()
        await Promise.all(keys.map((k) => caches.delete(k)))
      } catch {}
      try {
        await self.registration.unregister()
      } catch {}
      const clients = await self.clients.matchAll({
        type: 'window',
      })
      for (const client of clients) {
        try {
          client.navigate(client.url)
        } catch {}
      }
    })()
  )
})

self.addEventListener('fetch', () => {})
