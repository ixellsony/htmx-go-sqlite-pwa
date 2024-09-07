self.addEventListener('install', function(event) {
    event.waitUntil(
      caches.open('myapp-cache').then(function(cache) {
        return cache.addAll([
          '/static/styles.css',
          '/static/app.js',
        ]);
      })
    );
  });
  
  self.addEventListener('fetch', function(event) {
    event.respondWith(
      caches.match(event.request).then(function(response) {
        return response || fetch(event.request);
      })
    );
  });
  