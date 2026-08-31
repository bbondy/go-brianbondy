# Image workflow

When adding or replacing raster images under `static/img/`, run:

```sh
make responsive-images
```

This creates the 640w, 960w, and 1200w WebP variants used by `srcset` while
the original remains available for the lightbox. Include the generated variants
in the same commit as their source image. `make deploy` runs this target through
the `all` pipeline, but run it before committing to review and test the assets.

# Writing style

Do not use em dashes in prose, UI text, documentation, code comments, commit
messages, or other written content. Rewrite with commas, parentheses, colons,
semicolons, or separate sentences instead.
