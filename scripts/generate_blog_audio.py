#!/usr/bin/env python3
"""Generate English blog narration and sentence timings with local Ollama Phonon.

Examples:
  python3 scripts/generate_blog_audio.py --id 186
  python3 scripts/generate_blog_audio.py --all
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
import urllib.error
import urllib.request
import wave


ROOT = Path(__file__).resolve().parents[1]
MODEL = "phonon"
PIPELINE_VERSION = 1
SAMPLE_RATE = 24000
SILENCE_SECONDS = 0.18
MAX_CHARS = 1100
OLLAMA_MIN_VERSION = (0, 40, 9)
checked_ollama_hosts = set()


def require_command(name, install_hint):
    if shutil.which(name) is None:
        raise RuntimeError(f"{name} is required. {install_hint}")


def ensure_ollama(host):
    if host in checked_ollama_hosts:
        return
    require_command(
        "ollama", "On macOS, run: brew install ollama\n"
        "Then start it with: brew services start ollama\n"
        "See https://ollama.com/download for other installs.",
    )
    try:
        with urllib.request.urlopen(host.rstrip("/") + "/api/version", timeout=5) as response:
            version = json.load(response)["version"]
        with urllib.request.urlopen(host.rstrip("/") + "/api/tags", timeout=5) as response:
            models = json.load(response).get("models", [])
    except (OSError, ValueError, KeyError) as error:
        raise RuntimeError(
            f"Ollama is installed but its API is not available at {host}. "
            "Start Ollama (or run: brew services start ollama), then retry."
        ) from error
    match = re.match(r"^(\d+)\.(\d+)\.(\d+)", version)
    if not match or tuple(map(int, match.groups())) < OLLAMA_MIN_VERSION:
        raise RuntimeError(
            f"Ollama {version} is too old for Phonon. Upgrade to 0.40.9 or later "
            "(on macOS: brew upgrade ollama)."
        )
    if not any(model.get("name") in ("phonon", "phonon:latest") for model in models):
        print("Phonon is not installed; downloading it with ollama pull phonon...", flush=True)
        environment = os.environ.copy()
        environment["OLLAMA_HOST"] = host
        subprocess.run(["ollama", "pull", "phonon"], env=environment, check=True)
    checked_ollama_hosts.add(host)


def split_sentences(block):
    # Split at sentence punctuation while keeping quotation marks with the text.
    pieces = re.split(r'(?<=[.!?])\s+(?=["“‘A-Z])', block)
    for piece in pieces:
        piece = piece.strip()
        while len(piece) > MAX_CHARS:
            boundary = piece.rfind(" ", 0, MAX_CHARS + 1)
            if boundary < MAX_CHARS // 2:
                boundary = MAX_CHARS
            yield piece[:boundary].strip()
            piece = piece[boundary:].strip()
        if piece:
            yield piece


def speech(text, voice, host):
    body = json.dumps({
        "model": MODEL, "voice": voice, "input": text,
        "response_format": "pcm",
    }).encode()
    request = urllib.request.Request(
        host.rstrip("/") + "/v1/audio/speech", data=body,
        headers={"Content-Type": "application/json"}, method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=180) as response:
            audio = response.read()
    except urllib.error.HTTPError as error:
        raise RuntimeError(f"Ollama rejected speech request: {error.read().decode()}") from error
    if not audio or len(audio) % 2:
        raise RuntimeError("Ollama returned empty or invalid PCM audio")
    return audio


def source_blocks(source_binary, post_id):
    return json.loads(subprocess.check_output(
        [str(source_binary), str(post_id)], cwd=ROOT, text=True,
    ))


def cached_speech(text, voice, host, cache, force):
    key = hashlib.sha256(json.dumps([MODEL, voice, text], ensure_ascii=False).encode()).hexdigest()
    path = cache / f"{key}.pcm"
    if not force and path.exists():
        audio = path.read_bytes()
        if audio and len(audio) % 2 == 0:
            return audio
    print(f"  Generating: {text[:80]}", flush=True)
    ensure_ollama(host)
    audio = speech(text, voice, host)
    temporary = cache / f"{key}.tmp"
    temporary.write_bytes(audio)
    temporary.replace(path)
    return audio


def generate_post(post_id, source_binary, args, cache, output):
    blocks = source_blocks(source_binary, post_id)
    if not blocks:
        print(f"  Skipping post {post_id}: no prose to narrate", flush=True)
        return

    source_hash = hashlib.sha256(json.dumps(
        [PIPELINE_VERSION, MODEL, args.voice, SAMPLE_RATE, SILENCE_SECONDS, MAX_CHARS, blocks],
        ensure_ascii=False,
    ).encode()).hexdigest()
    mp3 = output / f"{post_id}.mp3"
    timings = output / f"{post_id}.json"
    if not args.force and mp3.is_file() and timings.is_file():
        try:
            current = json.loads(timings.read_text())
            audio_hash = hashlib.sha256(mp3.read_bytes()).hexdigest()
            if (current.get("sourceHash") == source_hash
                    and current.get("audioSHA256") == audio_hash
                    and current.get("sentences")):
                print(f"  Skipping post {post_id}: audio is current", flush=True)
                return
        except (OSError, ValueError):
            pass

    silence = b"\0\0" * round(SILENCE_SECONDS * SAMPLE_RATE)
    cues = []
    frames = 0
    with tempfile.TemporaryDirectory(prefix=f"post-{post_id}-", dir=cache) as temp_dir:
        wav_path = Path(temp_dir) / "narration.wav"
        mp3_path = Path(temp_dir) / "narration.mp3"
        with wave.open(str(wav_path), "wb") as wav:
            wav.setnchannels(1)
            wav.setsampwidth(2)
            wav.setframerate(SAMPLE_RATE)
            for block_index, block in enumerate(blocks):
                for text in split_sentences(block):
                    audio = cached_speech(text, args.voice, args.host, cache, args.force)
                    start = frames / SAMPLE_RATE
                    wav.writeframes(audio)
                    frames += len(audio) // 2
                    end = frames / SAMPLE_RATE
                    cues.append({
                        "block": block_index, "text": text,
                        "start": round(start, 3), "end": round(end, 3),
                    })
                    wav.writeframes(silence)
                    frames += len(silence) // 2

        subprocess.run([
            "ffmpeg", "-hide_banner", "-loglevel", "error", "-y",
            "-i", str(wav_path), "-codec:a", "libmp3lame", "-qscale:a", "3", str(mp3_path),
        ], check=True)
        mp3_path.replace(mp3)

    temporary_timings = output / f"{post_id}.json.tmp"
    temporary_timings.write_text(json.dumps({
        "post": post_id, "voice": args.voice, "model": MODEL,
        "sourceHash": source_hash,
        "audioSHA256": hashlib.sha256(mp3.read_bytes()).hexdigest(),
        "sentences": cues,
    }, indent=2) + "\n")
    temporary_timings.replace(timings)
    print(f"  Wrote {mp3.name} and {timings.name} ({frames / SAMPLE_RATE:.1f}s)", flush=True)


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    posts = ap.add_mutually_exclusive_group(required=True)
    posts.add_argument("--all", action="store_true", help="Generate all posts in the manifest")
    posts.add_argument("--id", type=int, help="Generate one post by ID")
    ap.add_argument("--voice", default="Easton", help="Phonon voice; see ollama show phonon")
    ap.add_argument("--host", default="http://localhost:11434", help="Ollama API URL")
    ap.add_argument("--force", action="store_true", help="Regenerate cached speech and output")
    args = ap.parse_args()

    require_command("go", "Install Go 1.19 or later: https://go.dev/dl/")
    require_command("ffmpeg", "On macOS, run: brew install ffmpeg")

    manifest = json.loads((ROOT / "data" / "blogPostManifest.json").read_text())
    known_ids = [post["id"] for post in manifest]
    if args.id is not None and args.id not in known_ids:
        ap.error(f"post {args.id} is not in data/blogPostManifest.json")
    selected_ids = known_ids if args.all else [args.id]

    cache = ROOT / ".cache" / "blog-audio"
    cache.mkdir(parents=True, exist_ok=True)
    output = ROOT / "static" / "audio" / "blog"
    output.mkdir(parents=True, exist_ok=True)
    source_binary = cache / "blog_audio_source"
    subprocess.run([
        "go", "build", "-o", str(source_binary), "scripts/blog_audio_source.go",
    ], cwd=ROOT, check=True)

    for number, post_id in enumerate(selected_ids, 1):
        print(f"[{number}/{len(selected_ids)}] Post {post_id}", flush=True)
        generate_post(post_id, source_binary, args, cache, output)


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, subprocess.CalledProcessError, urllib.error.URLError, OSError) as error:
        sys.exit(str(error))
