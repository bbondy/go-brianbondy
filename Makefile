format:
	golangci-lint run --fix

lint:
	golangci-lint run

test:
	go test -v 

deploy:
	gcloud app deploy 

auth:
	gcloud auth login 

# Convert images to WebP format for better performance
webp:
	python3 scripts/convert_images_to_webp.py

# Force convert all images to WebP (even if they already exist)
webp-force:
	python3 scripts/convert_images_to_webp.py --force

# Process new blog post images (run after adding a new blog post)
blog-images:
	python3 scripts/process_new_blog_images.py 