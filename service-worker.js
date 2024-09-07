const CACHE_NAME = 'myapp-cache-v1';
const urlsToCache = [
  '/static/styles.css',
  '/static/app.js',
  // Ajouter d'autres URLs si nécessaire
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(urlsToCache);
    }).catch((error) => {
      console.error('Erreur lors de l\'ouverture du cache :', error);
    })
  );
});

self.addEventListener('activate', (event) => {
  const cacheWhitelist = [CACHE_NAME];
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheWhitelist.indexOf(cacheName) === -1) {
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
});

self.addEventListener('fetch', (event) => {
  event.respondWith(
    caches.match(event.request).then((response) => {
      if (response) {
        return response;
      }

      return fetch(event.request).then((networkResponse) => {
        // Vérifiez que la requête et la réponse sont valides avant de mettre en cache
        if (event.request.method === 'GET' && networkResponse && networkResponse.status === 200 && networkResponse.type === 'basic') {
          const responseClone = networkResponse.clone();
          caches.open(CACHE_NAME).then((cache) => {
            // Vérifiez que la requête a un schéma supporté
            if (event.request.url.startsWith('http')) {
              cache.put(event.request, responseClone);
            }
          }).catch((error) => {
            console.error('Erreur lors de la mise en cache :', error);
          });
        }
        return networkResponse;
      }).catch((error) => {
        console.error('Erreur lors de la récupération de la requête :', error);
      });
    })
  );
});
