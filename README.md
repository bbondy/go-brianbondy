# go-brianbondy

A personal blog and portfolio site for Brian R. Bondy, built with Go. Features include blog posts (in markdown), running logs, project showcases, and more.

## Prerequisites

- Go 1.18+
- Python 3 (for scripts)
- golangci-lint (`brew install golangci-lint`)
- Google Cloud SDK (for deployment)

## Directory Structure

- `data/` - Blog post data, manifests, and markdown content
- `static/` - Static assets (CSS, images, fonts)
- `templates/` - HTML templates
- `scripts/` - Helper scripts (e.g., for books, Strava images)

## Adding a Blog Post

1. Create a new markdown file in `data/markdown/blog/`.
2. Add an entry to `data/blogPostManifest.json` with the new post's metadata.
3. (Optional) Add images to `static/img/blogpost_<id>/`.

## Development

```
go run .
```
The site will be available at [http://localhost:8080](http://localhost:8080).

## Format & Lint

To check for linting issues without fixing them:

```
make lint
```

To automatically format and fix linting issues:

```
make format
```

## Testing

To run all tests:

```
make test
```

## Deployment

Authenticate with Google Cloud (if you haven't already):

```
make auth
```

Then deploy:

```
make deploy
```

## Updating book list

Download an export from https://www.goodreads.com/review/import and save it to `data/goodreads_library_export.csv`.

Run `python3 scripts/generate_books.py`

## Where to publish blog posts

- Facebook page and related groups (Adjust visibility to Public)
- Twitter
- LinkedIn
- Strava (if about running)

## License

See [LICENSE](LICENSE).

