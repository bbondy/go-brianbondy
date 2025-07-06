# Blog Post Workflow Guide

This guide explains the complete process for adding new blog posts to brianbondy.com, including image optimization for better PageSpeed performance.

## Prerequisites

Before adding blog posts, ensure you have the required tools installed:

```bash
# Install WebP conversion tool
brew install webp  # macOS
# or
sudo apt-get install webp  # Ubuntu/Debian
# or
sudo yum install libwebp-tools  # CentOS/RHEL
```

## Complete Blog Post Workflow

### 1. Create the Blog Post Content

1. **Create markdown file**: Add a new file in `data/markdown/blog/` with the next available ID
   ```bash
   # Example: for blog post 191
   touch data/markdown/blog/191.markdown
   ```

2. **Add metadata**: Update `data/blogPostManifest.json` with the new post's information
   ```json
   {
     "id": 191,
     "title": "Your Blog Post Title",
     "created": "2025-01-15",
     "tags": ["tag1", "tag2"],
     "fbImagePath": "static/img/blogpost_191/featured-image.jpg",
     "fbDescription": "Brief description for social media"
   }
   ```

### 2. Add Images

1. **Create image directory**: Create a directory for your blog post images
   ```bash
   mkdir -p static/img/blogpost_191
   ```

2. **Add your images**: Place your `.jpg`, `.jpeg`, or `.png` images in the directory
   ```bash
   # Example: adding images
   cp your-image1.jpg static/img/blogpost_191/
   cp your-image2.png static/img/blogpost_191/
   ```

### 3. Process Images for WebP Optimization

**Option A: Process all blog post images (recommended)**
```bash
make blog-images
```

**Option B: Process only your specific blog post**
```bash
python3 scripts/process_new_blog_images.py 191
```

**Option C: Process all images in the entire site**
```bash
make webp
```

### 4. Test and Deploy

1. **Test locally**:
   ```bash
   go run .
   ```
   Visit `http://localhost:8080` to verify your blog post appears correctly

2. **Run tests**:
   ```bash
   make test
   ```

3. **Deploy**:
   ```bash
   make deploy
   ```

## Image Processing Details

### What the Image Processing Does

The image processing scripts automatically:

1. **Convert to WebP**: Convert `.jpg`, `.jpeg`, and `.png` files to `.webp` format
2. **Optimize quality**: Use quality setting of 80 (good balance of size vs quality)
3. **Smart conversion**: Only convert images that don't already have WebP versions or are newer than existing WebP files
4. **Size reporting**: Show file size savings for each converted image

### Available Commands

| Command | Description |
|---------|-------------|
| `make webp` | Convert all images in `static/img` to WebP |
| `make webp-force` | Force convert all images (even if WebP exists) |
| `make blog-images` | Process all blog post images |
| `python3 scripts/process_new_blog_images.py [ID]` | Process specific blog post images |

### Script Options

**Main WebP conversion script** (`scripts/convert_images_to_webp.py`):
```bash
python3 scripts/convert_images_to_webp.py --help
# Options:
#   --force      Convert all images even if .webp already exists
#   --directory  Specify directory to process (default: static/img)
#   --quality    WebP quality 0-100 (default: 80)
```

**Blog post specific script** (`scripts/process_new_blog_images.py`):
```bash
python3 scripts/process_new_blog_images.py --help
# Options:
#   blog_post_id  Specific blog post ID to process (e.g., 191)
#   --force       Force convert all images even if .webp already exists
```

## How Images Are Optimized on the Website

The website automatically optimizes images through the `optimizeImages` template function:

1. **WebP format**: Uses `.webp` versions when available
2. **Lazy loading**: Adds `loading="lazy"` to images
3. **Async decoding**: Adds `decoding="async"` for better performance
4. **Responsive images**: Adds `srcset` for different screen densities
5. **Fallback support**: Falls back to original format if WebP not supported

### Example Output

When you run the image processing, you'll see output like:
```
✓ Converted: static/img/blogpost_191/image1.jpg → static/img/blogpost_191/image1.webp
  Size: 245,760 bytes → 98,304 bytes (60.0% smaller)
⏭ Skipped: static/img/blogpost_191/image2.jpg (WebP already exists and up-to-date)
```

## Troubleshooting

### Common Issues

1. **WebP tool not found**:
   ```bash
   Error: cwebp tool not found. Please install it first:
     macOS: brew install webp
     Ubuntu/Debian: sudo apt-get install webp
     CentOS/RHEL: sudo yum install libwebp-tools
   ```

2. **Permission errors**: Make sure you have write permissions to the image directories

3. **Images not showing**: Check that the image paths in your markdown match the actual file locations

### Performance Benefits

- **Smaller file sizes**: WebP typically reduces image size by 25-35%
- **Faster loading**: Smaller files load faster
- **Better PageSpeed scores**: Optimized images improve overall site performance
- **Modern format support**: WebP is supported by all modern browsers

## Best Practices

1. **Use descriptive filenames**: `race-finish-line.jpg` instead of `img1.jpg`
2. **Optimize source images**: Start with reasonably sized images (don't upload 10MB photos)
3. **Run image processing after adding images**: Always run `make blog-images` after adding new images
4. **Test locally first**: Always test your blog post locally before deploying
5. **Keep original images**: The scripts preserve your original images alongside the WebP versions

## File Structure Example

After processing, your blog post directory will look like:
```
static/img/blogpost_191/
├── featured-image.jpg      # Original image
├── featured-image.webp     # Optimized WebP version
├── race-photo.jpg          # Original image
├── race-photo.webp         # Optimized WebP version
└── finish-line.png         # Original image
    └── finish-line.webp    # Optimized WebP version
```

The website will automatically serve the WebP versions to supported browsers while falling back to the original formats for older browsers. 