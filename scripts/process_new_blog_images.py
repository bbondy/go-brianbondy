#!/usr/bin/env python3
"""
Script to process new blog post images.
This script is designed to be run after adding a new blog post to automatically
convert any new images to WebP format and optimize them for the website.

Usage:
    python3 scripts/process_new_blog_images.py [blog_post_id]
    
If blog_post_id is provided, it will only process images in that specific blog post directory.
If no blog_post_id is provided, it will process all blog post directories.
"""

import os
import sys
import subprocess
import argparse
from pathlib import Path

def run_webp_conversion(directory=None, force=False):
    """Run the WebP conversion script"""
    cmd = ['python3', 'scripts/convert_images_to_webp.py']
    
    if directory:
        cmd.extend(['--directory', directory])
    
    if force:
        cmd.append('--force')
    
    print(f"Running: {' '.join(cmd)}")
    result = subprocess.run(cmd, cwd=os.getcwd())
    return result.returncode == 0

def find_blog_post_directories():
    """Find all blog post directories in static/img"""
    img_dir = Path('static/img')
    blog_dirs = []
    
    if not img_dir.exists():
        print("Error: static/img directory not found")
        return []
    
    # Look for directories that match blogpost_* pattern
    for item in img_dir.iterdir():
        if item.is_dir() and item.name.startswith('blogpost_'):
            blog_dirs.append(item)
    
    return sorted(blog_dirs)

def process_specific_blog_post(blog_post_id):
    """Process images for a specific blog post"""
    blog_dir = Path(f'static/img/blogpost_{blog_post_id}')
    
    if not blog_dir.exists():
        print(f"Error: Blog post directory not found: {blog_dir}")
        return False
    
    print(f"Processing blog post {blog_post_id}...")
    return run_webp_conversion(str(blog_dir))

def process_all_blog_posts():
    """Process images for all blog posts"""
    blog_dirs = find_blog_post_directories()
    
    if not blog_dirs:
        print("No blog post directories found in static/img")
        return True
    
    print(f"Found {len(blog_dirs)} blog post directories:")
    for blog_dir in blog_dirs:
        print(f"  - {blog_dir.name}")
    
    print("\nProcessing all blog post images...")
    return run_webp_conversion()

def main():
    parser = argparse.ArgumentParser(description='Process new blog post images')
    parser.add_argument('blog_post_id', nargs='?', type=int,
                       help='Specific blog post ID to process (e.g., 190)')
    parser.add_argument('--force', action='store_true',
                       help='Force convert all images even if .webp already exists')
    
    args = parser.parse_args()
    
    print("🖼️  Blog Post Image Processor")
    print("=" * 40)
    
    if args.blog_post_id:
        success = process_specific_blog_post(args.blog_post_id)
    else:
        success = process_all_blog_posts()
    
    if success:
        print("\n✅ Image processing completed successfully!")
        print("\nNext steps:")
        print("1. Test your website locally: go run .")
        print("2. Run tests: make test")
        print("3. Deploy: make deploy")
    else:
        print("\n❌ Image processing failed!")
        sys.exit(1)

if __name__ == '__main__':
    main() 