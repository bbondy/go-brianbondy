# Brian Bondy's Website

This is the source code for [brianbondy.com](https://brianbondy.com), a personal website built with Go.

## Development

### Prerequisites

- Go 1.19+
- Python 3.7+ (for image processing scripts)
- `cwebp` tool (for WebP image conversion)

### Installation

1. Clone the repository
2. Install the `cwebp` tool:
   ```bash
   # macOS
   brew install webp
   
   # Ubuntu/Debian
   sudo apt-get install webp
   
   # CentOS/RHEL
   sudo yum install libwebp-tools
   ```

### Running Locally

```bash
go run .
```

The site will be available at `http://localhost:8080`

### Testing

```bash
make test
```

### Formatting

```bash
make format
```

### Deployment

```bash
make deploy
```

## Blog Post Workflow

### Adding a New Blog Post

1. Create a new markdown file in `data/markdown/blog/` with the next available ID
2. Add the blog post metadata to `data/blogPostManifest.json`
3. Add images to `static/img/blogpost_[ID]/` directory
4. Process the images for WebP optimization:
   ```bash
   make blog-images [ID]
   ```
   Or process all blog post images:
   ```bash
   make blog-images
   ```
5. Test locally: `go run .`
6. Run tests: `make test`
7. Deploy: `make deploy`

### Image Processing

The website automatically optimizes images for better performance by:
- Converting images to WebP format
- Adding lazy loading
- Adding async decoding
- Providing responsive image support

#### Manual Image Processing

Convert all images to WebP:
```bash
make webp
```

Force convert all images (even if WebP already exists):
```bash
make webp-force
```

Process images for a specific blog post:
```bash
python3 scripts/process_new_blog_images.py [blog_post_id]
```

#### Image Processing Scripts

- `scripts/convert_images_to_webp.py` - Main WebP conversion script
- `scripts/process_new_blog_images.py` - Blog post specific image processing
- `scripts/download_strava_images.py` - Download images from Strava activities
- `scripts/generate_books.py` - Generate book data from Goodreads export

## Project Structure

- `data/` - Blog posts, projects, and other content
- `static/` - CSS, images, and other static assets
- `templates/` - HTML templates
- `scripts/` - Utility scripts for content management
- `handlers.go` - HTTP request handlers
- `routes.go` - URL routing
- `utils.go` - Utility functions including image optimization

## Prerequisites

- Go 1.18+
- Python 3 (for scripts)
- golangci-lint (`brew install golangci-lint`)
- Google Cloud SDK (for deployment)

## Adding a Blog Post

1. Create a new markdown file in `data/markdown/blog/`.
2. Add an entry to `data/blogPostManifest.json` with the new post's metadata.
3. (Optional) Add images to `static/img/blogpost_<id>/`.

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

## Running Data Automation

### Auto-Fetching Run Times from Strava

The script `scripts/fetch_strava_times.py` helps automate the process of adding elapsed time (hours and minutes) to each running activity in `data/runManifest.json`.

**Features:**
- Extracts time from the description if present (and cleans up duplicates)
- If time is missing, fetches the elapsed time from the public Strava activity web page (no API credentials required)
- Updates the manifest with a `"time"` field for each activity
- Cleans up the description to avoid duplicate time display

**Requirements:**
- Python 3
- `requests` and `beautifulsoup4` libraries (install with `pip install requests beautifulsoup4`)

**Usage:**

```bash
python3 scripts/fetch_strava_times.py
```

After running, your `data/runManifest.json` will be updated with time fields for each activity. Activities without a Strava activity URL or with non-standard pages will be flagged for manual review.


### Auto-Fetching GitHub Project Stats

The script `scripts/fetch_github_stats.py` automates fetching commit and pull request counts for your projects from GitHub and updates `data/projectManifest.json`.

**Features:**
- Scrapes GitHub search pages for commit and PR counts by author (bbondy)
- Supports keyword-based filtering for subprojects (see `searchKeywords` in the manifest)
- Handles abbreviated numbers (e.g., "2.3k" → 2300)
- Retries on rate limiting with exponential backoff
- Waits 2 seconds between all requests to avoid rate limits
- Only includes real fetched data (removes stats if not available)

**Requirements:**
- Python 3
- `requests` library (install with `pip install requests`)

**Usage:**

```bash
python3 scripts/fetch_github_stats.py
```

After running, your `data/projectManifest.json` will be updated with the latest commit and PR counts for each project. If a project's data can't be fetched (e.g., due to rate limiting), it will be omitted from the stats until a successful fetch.

