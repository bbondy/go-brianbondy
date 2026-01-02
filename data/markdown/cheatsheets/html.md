# html5

## Document

Doctype and root:  
```html
<!doctype html>
<html lang="en">
```

Head basics:  
```html
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Title</title>
```

Base URL for relative links:  
`<base href="https://example.com/">`

Body:  
```html
</head>
<body>
  ...
</body>
</html>
```

## Semantics

Layout landmarks:  
`<header>` `<nav>` `<main>` `<section>` `<article>` `<aside>` `<footer>`

Headings:  
`<h1>` ... `<h6>`

Text grouping:  
`<p>` `<blockquote>` `<pre>` `<hr>`

Inline text:  
`<strong>` `<em>` `<code>` `<span>` `<mark>` `<small>` `<time>`  
`<abbr>` `<cite>` `<kbd>` `<samp>` `<sub>` `<sup>` `<bdi>` `<bdo>` `<data>`

Lists:  
`<ul>` `<ol>` `<li>` `<dl>` `<dt>` `<dd>`

## Links and navigation

Link:  
`<a href="https://example.com">Text</a>`

Open in new tab safely:  
`<a href="..." target="_blank" rel="noopener">`

Page anchor:  
`<a href="#section-id">` with `<section id="section-id">`

## Images and media

Image with alt:  
`<img src="image.jpg" alt="Description">`

Responsive image:  
```html
<img
  src="image-800.jpg"
  srcset="image-400.jpg 400w, image-800.jpg 800w"
  sizes="(max-width: 600px) 100vw, 600px"
  alt="Description">
```

Figure with caption:  
```html
<figure>
  <img src="..." alt="...">
  <figcaption>Caption</figcaption>
</figure>
```

Audio/video:  
```html
<audio controls src="song.mp3"></audio>
<video controls width="640" src="movie.mp4"></video>
```

Picture art direction:  
```html
<picture>
  <source media="(min-width: 800px)" srcset="wide.jpg">
  <img src="fallback.jpg" alt="Description">
</picture>
```

Media text tracks:  
```html
<video controls src="movie.mp4">
  <track kind="subtitles" srclang="en" src="subs.vtt" default>
</video>
```

## Tables

Basic table:  
```html
<table>
  <caption>Caption</caption>
  <thead>
    <tr><th scope="col">A</th><th scope="col">B</th></tr>
  </thead>
  <tbody>
    <tr><td>1</td><td>2</td></tr>
  </tbody>
</table>
```

Table columns:  
```html
<colgroup>
  <col span="2">
  <col class="highlight">
</colgroup>
```

## Forms

Form structure:  
```html
<form action="/submit" method="post">
  <label for="email">Email</label>
  <input id="email" name="email" type="email" required>
  <button type="submit">Send</button>
</form>
```

Common inputs:  
`text` `email` `password` `number` `date` `time` `datetime-local` `month` `week`  
`checkbox` `radio` `range` `color` `file` `tel` `url` `search` `hidden`

Input helpers:  
`placeholder` `required` `min` `max` `step` `minlength` `maxlength`  
`pattern` `autocomplete` `autofocus` `disabled` `readonly`  
`inputmode` `enterkeyhint` `autocapitalize` `spellcheck`

Textarea and select:  
```html
<textarea name="message" rows="4"></textarea>
<select name="size">
  <option value="s">Small</option>
</select>
```

Grouped controls:  
```html
<fieldset>
  <legend>Shipping</legend>
  ...
</fieldset>
```

Form validation:  
`required` `pattern` `type="email"` `type="url"`

## Forms (advanced)

Datalist:  
```html
<input list="cities" name="city">
<datalist id="cities">
  <option value="Paris">
</datalist>
```

Progress and meter:  
`<progress value="30" max="100"></progress>`  
`<meter value="0.6" min="0" max="1"></meter>`

Output:  
`<output name="total">42</output>`

Button overrides:  
`<button formaction="/save" formmethod="post">Save</button>`  
`<button formnovalidate>Skip validation</button>`

Multiple files:  
`<input type="file" multiple>`

Capture from device:  
`<input type="file" accept="image/*" capture="environment">`

## Interactive

Details/summary:  
```html
<details>
  <summary>More</summary>
  <p>Hidden content</p>
</details>
```

Dialog:  
```html
<dialog open>Dialog content</dialog>
```

Editable content:  
`<div contenteditable="true"></div>`

## Metadata and SEO

Description and icons:  
```html
<meta name="description" content="...">
<link rel="icon" href="/favicon.ico">
```

Canonical and robots:  
```html
<link rel="canonical" href="https://example.com/page">
<meta name="robots" content="index,follow">
```

Social previews:  
```html
<meta property="og:title" content="...">
<meta property="og:image" content="...">
```

## Scripts and modules

Defer/async:  
`<script src="app.js" defer></script>`  
`<script src="analytics.js" async></script>`

Modules:  
`<script type="module" src="app.mjs"></script>`

Import maps:  
```html
<script type="importmap">
{
  "imports": {
    "lit": "https://cdn.example.com/lit.js"
  }
}
</script>
```

Preload module:  
`<link rel="modulepreload" href="/app.mjs">`

Noscript:  
`<noscript>Enable JavaScript</noscript>`

## Templates

Template element:  
```html
<template id="card-template">
  <article class="card">...</article>
</template>
```

Slots (web components):  
```html
<my-card>
  <span slot="title">Title</span>
</my-card>
```

## Accessibility

Landmarks and labels:  
`<main>` `<nav>` `<header>` `<footer>` `<label for="id">`

Alt text:  
`<img alt="Meaningful description">`

ARIA basics:  
`role="dialog"` `aria-label="Close"` `aria-expanded="false"`

## Embeds

Iframes:  
`<iframe src="..." title="Description"></iframe>`

Inline SVG:  
```html
<svg viewBox="0 0 24 24" aria-hidden="true">
  <path d="..."></path>
</svg>
```

## Storage and offline

Local/session storage:  
```html
<script>
  localStorage.setItem("theme", "dark");
  const theme = localStorage.getItem("theme");
  sessionStorage.setItem("draft", "...");
</script>
```

Offline hint:  
`<meta name="theme-color" content="#111111">`

Storage limits (browser-dependent):  
`navigator.storage.estimate().then(({ quota, usage }) => {})`

IndexedDB (open DB):  
```html
<script>
  const request = indexedDB.open("app-db", 1);
  request.onupgradeneeded = () => {
    const db = request.result;
    db.createObjectStore("items", { keyPath: "id" });
  };
</script>
```

Cache API (in SW):  
```html
<script>
// sw.js
self.addEventListener("fetch", (event) => {
  event.respondWith(
    caches.match(event.request).then((cached) => cached || fetch(event.request))
  );
});
</script>
```

## Service workers

Register service worker:  
```html
<script>
  if ("serviceWorker" in navigator) {
    navigator.serviceWorker.register("/sw.js");
  }
</script>
```

Basic cache in sw:  
```html
<script>
// sw.js
self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open("v1").then((cache) => cache.addAll(["/"]))
  );
});
</script>
```

## PWA manifest

Manifest link:  
`<link rel="manifest" href="/manifest.json">`

Basic manifest:  
```html
<script>
// manifest.json
{
  "name": "App",
  "short_name": "App",
  "start_url": "/",
  "display": "standalone",
  "icons": [{ "src": "/icon-192.png", "sizes": "192x192", "type": "image/png" }]
}
</script>
```

## History and navigation

History API:  
```html
<script>
  history.pushState({ page: 2 }, "", "?page=2");
  window.addEventListener("popstate", (e) => {});
</script>
```

Location helpers:  
`location.href` `location.search` `location.hash`

## Fetch and streams

Fetch with abort:  
```html
<script>
  const controller = new AbortController();
  fetch("/api", { signal: controller.signal });
  controller.abort();
</script>
```

Readable stream:  
```html
<script>
  const stream = new ReadableStream({
    start(controller) { controller.enqueue(new TextEncoder().encode("hi")); }
  });
</script>
```

## Drag and drop

Draggable element:  
`<div draggable="true"></div>`

Drop target:  
```html
<div ondragover="event.preventDefault()" ondrop="onDrop(event)"></div>
```

## Canvas

Canvas element:  
`<canvas id="chart" width="300" height="150"></canvas>`

Basic drawing:  
```html
<script>
  const ctx = document.getElementById("chart").getContext("2d");
  ctx.fillStyle = "#0ea5e9";
  ctx.fillRect(10, 10, 100, 50);
</script>
```

## WebSocket

Connect and send:  
```html
<script>
  const ws = new WebSocket("wss://example.com/socket");
  ws.addEventListener("open", () => ws.send("hello"));
  ws.addEventListener("message", (e) => {});
</script>
```

## Workers

Web worker:  
```html
<script>
  const worker = new Worker("worker.js");
  worker.postMessage({ task: "work" });
</script>
```

Shared worker:  
```html
<script>
  const worker = new SharedWorker("shared.js");
  worker.port.start();
</script>
```

## Geolocation

Get current position:  
```html
<script>
  navigator.geolocation.getCurrentPosition(
    (pos) => {},
    (err) => {},
    { enableHighAccuracy: true }
  );
</script>
```

## Permissions

Query permission state:  
```html
<script>
  navigator.permissions.query({ name: "geolocation" })
    .then((status) => {
      console.log(status.state);
    });
</script>
```

Prompt on use:  
`navigator.geolocation.getCurrentPosition(...)`  
`navigator.mediaDevices.getUserMedia(...)`

## Events

Pointer and touch:  
`pointerdown` `pointermove` `touchstart`

Keyboard:  
`keydown` `keyup` `input`

Passive listener:  
```html
<script>
  window.addEventListener("scroll", () => {}, { passive: true });
</script>
```

## WebRTC

Get media stream:  
```html
<script>
  navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    .then((stream) => {
      const video = document.querySelector("video");
      video.srcObject = stream;
    });
</script>
```

Peer connection (minimal):  
```html
<script>
  const pc = new RTCPeerConnection();
  pc.onicecandidate = (e) => {};
</script>
```

## Web Audio

Audio context and oscillator:  
```html
<script>
  const ctx = new AudioContext();
  const osc = ctx.createOscillator();
  osc.connect(ctx.destination);
  osc.start();
  osc.stop(ctx.currentTime + 0.5);
</script>
```

## Clipboard and share

Clipboard read/write:  
```html
<script>
  navigator.clipboard.writeText("hello");
  navigator.clipboard.readText().then((text) => {});
</script>
```

Web Share API:  
```html
<script>
  navigator.share({ title: "Title", url: "https://example.com" });
</script>
```

## Notifications and badges

Notifications:  
```html
<script>
  Notification.requestPermission().then((perm) => {
    if (perm === "granted") new Notification("Hello");
  });
</script>
```

App badge:  
`navigator.setAppBadge(3)`  
`navigator.clearAppBadge()`

## File System Access

Pick file and read:  
```html
<script>
  const [handle] = await window.showOpenFilePicker();
  const file = await handle.getFile();
</script>
```

## Security and performance

CSP meta:  
`<meta http-equiv="Content-Security-Policy" content="default-src 'self'">`

SRI for scripts:  
```html
<script src="app.js" integrity="sha384-..." crossorigin="anonymous"></script>
```

Preconnect and DNS prefetch:  
`<link rel="preconnect" href="https://cdn.example.com">`  
`<link rel="dns-prefetch" href="//cdn.example.com">`

Referrer policy:  
`<meta name="referrer" content="no-referrer">`

Fetch priority hints:  
`<img src="hero.jpg" fetchpriority="high">`
