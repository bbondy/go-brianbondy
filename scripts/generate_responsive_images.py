#!/usr/bin/env python3
"""Generate responsive WebP variants next to the site's source images.

The original image remains the lightbox source. Generated variants are used by
srcset for inline images, so browsers download only the needed size.
"""

import argparse
import re
import shutil
import subprocess
import sys
from pathlib import Path

WIDTHS = (640, 960, 1200)
IMAGE_SUFFIXES = {".jpg", ".jpeg", ".png", ".webp"}
VARIANT_SUFFIX = re.compile(r"-(640|960|1200)\.webp$", re.IGNORECASE)


def image_width(path: Path) -> int:
    result = subprocess.run(
        ["sips", "-g", "pixelWidth", str(path)],
        capture_output=True,
        check=True,
        text=True,
    )
    match = re.search(r"pixelWidth:\s*(\d+)", result.stdout)
    if not match:
        raise ValueError(f"could not read width for {path}")
    return int(match.group(1))


def variant_path(path: Path, width: int) -> Path:
    return path.with_name(f"{path.stem}-{width}.webp")


def source_images(root: Path):
    for path in root.rglob("*"):
        if not path.is_file() or path.suffix.lower() not in IMAGE_SUFFIXES:
            continue
        if VARIANT_SUFFIX.search(path.name):
            continue
        yield path


def generate_variant(source: Path, destination: Path, width: int, dry_run: bool) -> bool:
    if destination.exists() and destination.stat().st_mtime >= source.stat().st_mtime:
        return False
    command = ["cwebp", "-quiet", "-q", "82", "-resize", str(width), "0", str(source), "-o", str(destination)]
    if dry_run:
        print(" ".join(command))
        return True
    subprocess.run(command, check=True)
    return True


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("directory", nargs="?", default="static/img")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    if shutil.which("cwebp") is None:
        print("cwebp is required to generate responsive images", file=sys.stderr)
        return 1

    root = Path(args.directory)
    if not root.is_dir():
        print(f"image directory does not exist: {root}", file=sys.stderr)
        return 1

    generated = 0
    for source in source_images(root):
        try:
            width = image_width(source)
        except (subprocess.CalledProcessError, ValueError) as error:
            print(f"skipping {source}: {error}", file=sys.stderr)
            continue
        for variant_width in WIDTHS:
            if variant_width < width:
                generated += generate_variant(source, variant_path(source, variant_width), variant_width, args.dry_run)

    print(f"Generated {generated} responsive image variants")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
