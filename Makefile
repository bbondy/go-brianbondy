.DEFAULT_GOAL := update

# Every target is a command, not a file. Without this, `make locales` is a no-op
# because the locales/ directory already exists and looks up to date.
.PHONY: update memorable-runs cheatsheets locales blog-translations all format \
	lint test update-cache deploy auth webp webp-force responsive-images \
	blog-images github-stats pictures-manifest run-km-manifest \
	strava-run-manifest build dev cache-version strava-clear-token help

update: format lint test update-cache webp github-stats memorable-runs

memorable-runs:
	python3 scripts/fetch_memorable_runs.py

cheatsheets:
	python3 scripts/generate_cheatsheets.py

locales:
	python3 scripts/update_locales.py

blog-translations:
	python3 scripts/translate_blog_posts.py

# Default target - runs all essential tasks
all: lint format test update-cache webp responsive-images pictures-manifest

# Code quality
format:
	golangci-lint run --fix

lint:
	golangci-lint run

test:
	go test -v

# Update cache busting version in templates
update-cache:
	@echo "Updating cache busting version..."
	@current_version=$$(grep -m1 -o 'main.css?v=[0-9]*' templates/base.html | cut -d= -f2); \
	if [ -z "$$current_version" ]; then \
		echo "Could not find the main.css cache version in templates/base.html" >&2; \
		exit 1; \
	fi; \
	new_version=$$(expr $$current_version + 1); \
	sed -i '' "s/main.css?v=$$current_version/main.css?v=$$new_version/g" templates/base.html; \
	echo "Cache busting updated to version $$new_version"

# Deployment
deploy: all
	gcloud app deploy

auth:
	gcloud auth login

# Image optimization
webp:
	python3 scripts/convert_images_to_webp.py

# Force convert all images to WebP (even if they already exist)
webp-force:
	python3 scripts/convert_images_to_webp.py --force

responsive-images:
	python3 scripts/generate_responsive_images.py

# Process new blog post images (run after adding a new blog post)
blog-images:
	python3 scripts/process_new_blog_images.py

# Fetch GitHub statistics for projects
github-stats:
	python3 scripts/fetch_github_stats.py

# Generate pictures manifest for running-tagged blog posts
pictures-manifest:
	python3 scripts/generate_pictures_manifest.py

# Generate run manifest with km and pace
run-km-manifest:
	python3 scripts/generate_run_km_manifest.py

# Generate Strava run manifest using Strava API
strava-run-manifest:
	python3 scripts/generate_strava_run_manifest.py

# Build for production (everything except deploy)
build: all

# Quick development setup (just code quality)
dev: lint format test

# Show current cache busting version
cache-version:
	@grep -o 'cachebust=[0-9]*' templates/base.html

# Clear saved Strava OAuth token
strava-clear-token:
	rm -f ~/.strava_token.json

# Show help
help:
	@echo "Available targets:"
	@echo "  update       - Run all update scripts (format, lint, test, cache, webp, github-stats, memorable-runs)"
	@echo "  all          - Run all essential tasks (lint, format, test, cache, webp)"
	@echo "  build        - Same as 'all' (production build)"
	@echo "  dev          - Quick development setup (lint, format, test)"
	@echo "  format       - Format code with golangci-lint"
	@echo "  lint         - Lint code with golangci-lint"
	@echo "  test         - Run tests"
	@echo "  update-cache - Update cache busting version"
	@echo "  webp         - Convert images to WebP"
	@echo "  webp-force   - Force convert all images to WebP"
	@echo "  responsive-images - Generate 640w, 960w, and 1200w WebP variants"
	@echo "  blog-images  - Process new blog post images"
	@echo "  blog-translations - Generate missing French blog sidecars with local Ollama"
	@echo "  github-stats - Fetch GitHub statistics for projects"
	@echo "  memorable-runs - Fetch memorable run statistics"
	@echo "  cheatsheets   - Generate cheatsheets manifest and markdown from GitHub"
	@echo "  locales       - Extract and deduplicate translatable UI strings"
	@echo "  strava-clear-token - Delete cached Strava OAuth token (~/.strava_token.json)"
	@echo "  deploy       - Deploy to Google App Engine"
	@echo "  cache-version- Show current cache busting version"
	@echo "  run-km-manifest - Generate run manifest with km and pace"
	@echo "  strava-run-manifest - Generate run manifest using Strava API (expects STRAVA_ACCESS_TOKEN or STRAVA_CLIENT_ID/STRAVA_CLIENT_SECRET/STRAVA_REFRESH_TOKEN in env, e.g. via .envrc)"
