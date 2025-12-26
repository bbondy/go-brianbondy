# Repository Guidelines

## Project Structure & Module Organization
- `main.go`, `routes.go`, `handlers.go`, `template.go`, `utils.go` are the core Go entrypoints, routing, handlers, and helpers.
- `templates/` holds HTML templates; `static/` contains CSS, images, and other assets.
- `data/` stores markdown content, manifests, and data files used by the site.
- `scripts/` contains Python utilities for content/image processing and data automation.
- Tests live alongside source in `*_test.go` files (e.g., `handlers_test.go`).

## Build, Test, and Development Commands
- `go run .` runs the site locally at `http://localhost:8080`.
- `make test` runs `go test -v` for all packages.
- `make lint` runs `golangci-lint` checks; `make format` auto-fixes lint issues.
- `make webp` converts images to WebP; `make blog-images` processes images for a new blog post.
- `make deploy` runs the full `all` pipeline and deploys to Google App Engine (`app.yaml`).

## Coding Style & Naming Conventions
- Go formatting follows `gofmt` conventions via `golangci-lint run --fix`.
- Use Go naming: `CamelCase` for exported identifiers, `camelCase` for unexported.
- Keep files grouped by feature; tests mirror file names (e.g., `routes.go` → `routes_test.go`).

## Testing Guidelines
- Primary framework is Go’s built-in `testing` package.
- Name tests `TestXxx` in `*_test.go`.
- Run the full suite with `make test` before submitting changes.

## Commit & Pull Request Guidelines
- Commit messages are short, imperative, and descriptive (e.g., “Update stats data”, “Fix border on light mode”).
- PRs should include a clear description, relevant links to issues, and screenshots for UI/template changes.
- Note any data or content updates (e.g., `data/` or `static/` additions) in the PR description.

## Configuration & Automation Notes
- Deployment requires Google Cloud SDK auth (`make auth`) and `gcloud` setup.
- Strava-related scripts may use `STRAVA_ACCESS_TOKEN`; don’t commit secrets or tokens.
- Python scripts assume Python 3 and may need extra packages (see script headers or README).
