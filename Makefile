.DEFAULT_GOAL := update

update: format lint test update-cache webp github-stats strava-stats

strava-stats:
	python3 scripts/fetch_strava_times.py

# Default target - runs all essential tasks
all: lint format test update-cache webp pictures-manifest

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
	@current_version=$$(grep -m1 -o 'cachebust=[0-9]*' templates/base.html | head -n1 | cut -d= -f2); \
	new_version=$$(expr $$current_version + 1); \
	sed -i '' "0,/cachebust=$$current_version/s//cachebust=$$new_version/" templates/base.html; \
	sed -i '' "0,/cachebust=$$current_version/s//cachebust=$$new_version/" templates/base.html; \
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

# Process new blog post images (run after adding a new blog post)
blog-images:
	python3 scripts/process_new_blog_images.py

# Fetch GitHub statistics for projects
github-stats:
	python3 scripts/fetch_github_stats.py

# Generate pictures manifest for running-tagged blog posts
pictures-manifest:
	python3 scripts/generate_pictures_manifest.py

# Build for production (everything except deploy)
build: all

# Quick development setup (just code quality)
dev: lint format test

# Show current cache busting version
cache-version:
	@grep -o 'cachebust=[0-9]*' templates/base.html

# Show help
help:
	@echo "Available targets:"
	@echo "  update       - Run all update scripts (format, lint, test, cache, webp, github-stats, strava-stats)"
	@echo "  all          - Run all essential tasks (lint, format, test, cache, webp)"
	@echo "  build        - Same as 'all' (production build)"
	@echo "  dev          - Quick development setup (lint, format, test)"
	@echo "  format       - Format code with golangci-lint"
	@echo "  lint         - Lint code with golangci-lint"
	@echo "  test         - Run tests"
	@echo "  update-cache - Update cache busting version"
	@echo "  webp         - Convert images to WebP"
	@echo "  webp-force   - Force convert all images to WebP"
	@echo "  blog-images  - Process new blog post images"
	@echo "  github-stats - Fetch GitHub statistics for projects"
	@echo "  strava-stats - Fetch Strava run statistics"
	@echo "  deploy       - Deploy to Google App Engine"
	@echo "  cache-version- Show current cache busting version" 