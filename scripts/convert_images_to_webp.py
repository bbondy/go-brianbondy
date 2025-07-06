#!/usr/bin/env python3
"""
Script to convert new images to WebP format for blog posts.
This script will find all .jpg, .jpeg, and .png files in static/img and convert them to .webp
if the .webp version doesn't already exist.

Usage:
    python3 scripts/convert_images_to_webp.py [--force] [--directory path]

Options:
    --force: Convert all images even if .webp already exists
    --directory: Specify a specific directory to process (default: static/img)
"""

import os
import sys
import subprocess
import argparse
from pathlib import Path

def check_webp_tool():
    """Check if cwebp tool is available"""
    try:
        subprocess.run(['cwebp', '-version'], capture_output=True, check=True)
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False

def convert_image_to_webp(input_path, output_path, quality=80):
    """Convert a single image to WebP format"""
    try:
        cmd = [
            'cwebp',
            '-q', str(quality),
            '-mt',  # Use multi-threading
            '-af',  # Auto-filter
            '-f', '40',  # Filter strength
            '-sharpness', '0',  # Sharpness
            input_path,
            '-o', output_path
        ]
        
        result = subprocess.run(cmd, capture_output=True, text=True)
        
        if result.returncode == 0:
            # Get file sizes for comparison
            original_size = os.path.getsize(input_path)
            webp_size = os.path.getsize(output_path)
            savings = ((original_size - webp_size) / original_size) * 100
            
            print(f"✓ Converted: {input_path} → {output_path}")
            print(f"  Size: {original_size:,} bytes → {webp_size:,} bytes ({savings:.1f}% smaller)")
            return True
        else:
            print(f"✗ Failed to convert {input_path}: {result.stderr}")
            return False
            
    except Exception as e:
        print(f"✗ Error converting {input_path}: {e}")
        return False

def should_convert_image(input_path, output_path, force=False):
    """Determine if an image should be converted"""
    if not os.path.exists(input_path):
        return False
    
    # If force is True, always convert
    if force:
        return True
    
    # If output doesn't exist, convert
    if not os.path.exists(output_path):
        return True
    
    # If input is newer than output, convert
    input_mtime = os.path.getmtime(input_path)
    output_mtime = os.path.getmtime(output_path)
    
    return input_mtime > output_mtime

def process_directory(directory_path, force=False, quality=80):
    """Process all images in a directory and its subdirectories"""
    directory = Path(directory_path)
    
    if not directory.exists():
        print(f"Error: Directory {directory_path} does not exist")
        return False
    
    # Supported image formats
    supported_formats = {'.jpg', '.jpeg', '.png'}
    
    # Find all supported image files
    image_files = []
    for ext in supported_formats:
        image_files.extend(directory.rglob(f'*{ext}'))
        image_files.extend(directory.rglob(f'*{ext.upper()}'))
    
    if not image_files:
        print(f"No supported image files found in {directory_path}")
        return True
    
    print(f"Found {len(image_files)} image files to process...")
    
    converted_count = 0
    skipped_count = 0
    failed_count = 0
    
    for image_path in image_files:
        # Generate WebP path
        webp_path = image_path.with_suffix('.webp')
        
        if should_convert_image(str(image_path), str(webp_path), force):
            if convert_image_to_webp(str(image_path), str(webp_path), quality):
                converted_count += 1
            else:
                failed_count += 1
        else:
            print(f"⏭ Skipped: {image_path} (WebP already exists and up-to-date)")
            skipped_count += 1
    
    print(f"\nConversion Summary:")
    print(f"  Converted: {converted_count}")
    print(f"  Skipped: {skipped_count}")
    print(f"  Failed: {failed_count}")
    print(f"  Total: {len(image_files)}")
    
    return failed_count == 0

def main():
    parser = argparse.ArgumentParser(description='Convert images to WebP format')
    parser.add_argument('--force', action='store_true', 
                       help='Convert all images even if .webp already exists')
    parser.add_argument('--directory', default='static/img',
                       help='Directory to process (default: static/img)')
    parser.add_argument('--quality', type=int, default=80,
                       help='WebP quality (0-100, default: 80)')
    
    args = parser.parse_args()
    
    # Check if cwebp tool is available
    if not check_webp_tool():
        print("Error: cwebp tool not found. Please install it first:")
        print("  macOS: brew install webp")
        print("  Ubuntu/Debian: sudo apt-get install webp")
        print("  CentOS/RHEL: sudo yum install libwebp-tools")
        sys.exit(1)
    
    print(f"Processing directory: {args.directory}")
    print(f"Force conversion: {args.force}")
    print(f"WebP quality: {args.quality}")
    print("-" * 50)
    
    success = process_directory(args.directory, args.force, args.quality)
    
    if success:
        print("\n✅ All conversions completed successfully!")
        sys.exit(0)
    else:
        print("\n❌ Some conversions failed!")
        sys.exit(1)

if __name__ == '__main__':
    main() 