#!/usr/bin/env python3
import json
import os
import re
import ssl
import sys
from pathlib import Path
from urllib.request import urlopen, Request


REPO_HTML_URL = "https://github.com/bbondy/cheatsheets"
RAW_BASE_URL = "https://raw.githubusercontent.com/bbondy/cheatsheets/master/"
OUTPUT_DIR = Path("data/markdown/cheatsheets")
MANIFEST_PATH = Path("data/cheatsheetsManifest.json")

DESCRIPTION_OVERRIDES = {
    "gdb": "Debugger commands and workflows for GNU Debugger (GDB).",
    "claude-code": "Notes and snippets for the Claude Code CLI workflow.",
    "docker": "Common Docker commands, images, and container operations.",
    "go": "Go language commands, tooling, and syntax reference.",
    "ipfs": "IPFS commands and concepts for content-addressed data.",
    "kubernetes": "Kubernetes kubectl commands and cluster workflows.",
    "rust": "Rust language syntax, tooling, and cargo commands.",
    "tmux": "tmux key bindings and session management shortcuts.",
}

TITLE_OVERRIDES = {
    "kubernetes": "Kubernetes",
}


def fetch_url(url: str) -> str:
    context = ssl.create_default_context()
    context.check_hostname = False
    context.verify_mode = ssl.CERT_NONE
    req = Request(url, headers={"User-Agent": "cheatsheet-generator"})
    with urlopen(req, context=context) as response:
        return response.read().decode("utf-8")


def extract_filenames(html: str) -> list[str]:
    pattern = re.compile(r'href="/bbondy/cheatsheets/blob/master/([^"#?]+\.md)"')
    matches = pattern.findall(html)
    filenames = sorted(set(matches))
    return [name for name in filenames if name.lower() != "readme.md"]


def title_from_markdown(markdown: str, fallback: str) -> str:
    for line in markdown.splitlines():
        line = line.strip()
        if line.startswith("# "):
            return line[2:].strip()
    return fallback


def humanize_slug(slug: str) -> str:
    words = slug.replace("-", " ").replace("_", " ").split()
    return " ".join(word.capitalize() for word in words)


def description_for(slug: str, title: str) -> str:
    if slug in DESCRIPTION_OVERRIDES:
        return DESCRIPTION_OVERRIDES[slug]
    return f"Cheatsheet for {title} commands, tips, and workflows."


def main() -> int:
    try:
        html = fetch_url(REPO_HTML_URL)
    except Exception as exc:
        print(f"Failed to fetch repository HTML: {exc}", file=sys.stderr)
        return 1

    filenames = extract_filenames(html)
    if not filenames:
        print("No markdown files found in repository HTML.", file=sys.stderr)
        return 1

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    manifest = []
    for filename in filenames:
        raw_url = RAW_BASE_URL + filename
        try:
            markdown = fetch_url(raw_url)
        except Exception as exc:
            print(f"Failed to fetch {raw_url}: {exc}", file=sys.stderr)
            return 1

        base_name = os.path.splitext(os.path.basename(filename))[0]
        slug = base_name.lower()
        fallback_title = humanize_slug(slug)
        title = TITLE_OVERRIDES.get(slug, title_from_markdown(markdown, fallback_title))
        description = description_for(slug, title)

        output_path = OUTPUT_DIR / f"{slug}.md"
        output_path.write_text(markdown, encoding="utf-8")

        manifest.append(
            {
                "title": title,
                "slug": slug,
                "description": description,
            }
        )

    MANIFEST_PATH.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
    print(f"Wrote {len(manifest)} cheatsheets to {MANIFEST_PATH}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
