#!/usr/bin/env python3
"""Publish generated blog MP3s to the public Cloud Storage audio bucket.

The small timing JSON files and this script's URL manifest stay in the repo.
"""

import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import subprocess
import sys
import time
import urllib.error
import urllib.request


ROOT = Path(__file__).resolve().parents[1]
OUTPUT = ROOT / "static" / "audio" / "blog"
MANIFEST = ROOT / "data" / "blogAudioManifest.json"
BUCKET_NAME = re.compile(r"^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$")


def sha256(path):
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def verify_public_audio(url, size):
    for attempt in range(5):
        try:
            request = urllib.request.Request(url, method="HEAD")
            with urllib.request.urlopen(request, timeout=20) as response:
                if response.status == 200 and int(response.headers.get("Content-Length", "0")) == size:
                    return
                raise RuntimeError(f"unexpected public response for {url}: {response.status}")
        except (urllib.error.HTTPError, urllib.error.URLError) as error:
            if attempt == 4:
                raise RuntimeError(f"uploaded audio is not publicly readable: {url}: {error}") from error
            time.sleep(2)
    raise RuntimeError(f"uploaded audio could not be verified: {url}")


def save_manifest(entries):
    temporary = MANIFEST.with_suffix(".json.tmp")
    ordered = {key: entries[key] for key in sorted(entries, key=int)}
    temporary.write_text(json.dumps(ordered, indent=2) + "\n")
    temporary.replace(MANIFEST)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    posts = parser.add_mutually_exclusive_group(required=True)
    posts.add_argument("--all", action="store_true", help="Publish all locally generated posts")
    posts.add_argument("--id", type=int, help="Publish one post")
    parser.add_argument("--bucket", required=True, help="Public Cloud Storage bucket name")
    parser.add_argument("--dry-run", action="store_true", help="Show planned uploads")
    args = parser.parse_args()
    if not BUCKET_NAME.fullmatch(args.bucket):
        parser.error("invalid Cloud Storage bucket name")
    if not args.dry_run and shutil.which("gcloud") is None:
        parser.error("gcloud is required to publish audio. Install the Google Cloud CLI: https://cloud.google.com/sdk/docs/install")

    post_manifest = json.loads((ROOT / "data" / "blogPostManifest.json").read_text())
    known_ids = [post["id"] for post in post_manifest]
    if args.id is not None and args.id not in known_ids:
        parser.error(f"post {args.id} is not in the blog manifest")
    selected = known_ids if args.all else [args.id]
    published = json.loads(MANIFEST.read_text())
    found = 0

    for post_id in selected:
        mp3 = OUTPUT / f"{post_id}.mp3"
        timing = OUTPUT / f"{post_id}.json"
        if not mp3.is_file() or not timing.is_file():
            if args.id is not None:
                parser.error(f"post {post_id} needs both a generated MP3 and timing JSON")
            continue
        found += 1
        details = json.loads(timing.read_text())
        audio_hash = sha256(mp3)
        if (details.get("post") != post_id or not details.get("sentences")
                or details.get("audioSHA256") != audio_hash):
            raise RuntimeError(f"post {post_id} has mismatched or incomplete audio and timings; regenerate it")
        timing_hash = sha256(timing)
        object_name = f"blog-audio/{post_id}-{audio_hash[:16]}.mp3"
        url = f"https://storage.googleapis.com/{args.bucket}/{object_name}"
        entry = {"audioURL": url, "timingSHA256": timing_hash}
        if published.get(str(post_id)) == entry:
            print(f"Post {post_id}: already published", flush=True)
            continue
        if args.dry_run:
            print(f"Post {post_id}: would upload {mp3} to gs://{args.bucket}/{object_name}")
            continue

        print(f"Post {post_id}: uploading {mp3.name}", flush=True)
        subprocess.run([
            "gcloud", "storage", "cp", "--no-clobber",
            "--cache-control=public, max-age=31536000, immutable",
            "--content-type=audio/mpeg", str(mp3),
            f"gs://{args.bucket}/{object_name}",
        ], check=True)
        verify_public_audio(url, mp3.stat().st_size)
        published[str(post_id)] = entry
        save_manifest(published)
        print(f"Post {post_id}: published {url}", flush=True)

    if found == 0:
        raise RuntimeError("no generated blog audio found; run make blog-audio first")


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError, RuntimeError, subprocess.CalledProcessError) as error:
        sys.exit(str(error))
