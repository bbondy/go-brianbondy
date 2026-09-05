# Image workflow

When adding or replacing raster images under `static/img/`, run:

```sh
make responsive-images
```

This creates the 640w, 960w, and 1200w WebP variants used by `srcset` while
the original remains available for the lightbox. Include the generated variants
in the same commit as their source image. `make deploy` runs this target through
the `all` pipeline, but run it before committing to review and test the assets.

# Cache busting for CSS and JS

Every stylesheet and script in `templates/base.html` is versioned with a `?v=N`
query string and served with `cache-control: max-age=31536000, immutable`. A
browser or the App Engine edge that already has a URL will never refetch it, so
changing a file without changing its version ships nothing to returning
visitors: they keep the old asset and mix it with the new inline critical CSS in
`base.html`, which is how the mobile top bar ended up with the nav links and the
theme toggles drawn on top of each other.

So any time you edit a file under `static/css/` or `static/js/`, bump its
version in the same commit.

For `static/css/main.css`, run:

```sh
make update-cache
```

That target only rewrites the `main.css?v=` occurrences. Every other versioned
asset (`editorial.css`, `resume.css`, `about.css`, `contact.css`,
`projects.css`, `running.css`, `all-posts.css`, `filters.css`,
`collections.css`, `home.css`, `lightbox.css`, and the versioned scripts) has
its own counter, so increment those by hand in `templates/base.html`.

Editing the inline critical CSS inside `base.html` needs no version bump, since
that markup is generated per request. Bump `main.css` anyway when the inline
change mirrors a rule that also lives there, so the two cannot disagree.

Localhost will not reproduce a stale asset bug: it is normally hitting each URL
for the first time. Verify against production with a cache busting query string,
for example `curl -s --compressed "https://brianbondy.com/static/css/main.css?nocache=1"`,
and compare it to the version the deployed HTML actually requests.

# Writing style

Do not use em dashes in prose, UI text, documentation, code comments, commit
messages, or other written content. Rewrite with commas, parentheses, colons,
semicolons, or separate sentences instead.
