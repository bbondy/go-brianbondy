#!/usr/bin/env python3
"""Create French blog sidecars with a locally installed Ollama model.

The generated files are a starting point for editorial review. Existing files are
never replaced unless --force is supplied, so hand-edited translations are safe.
"""

import argparse
import json
import re
import sys
import urllib.error
import urllib.request
from pathlib import Path
from typing import Dict, Iterable, List, Tuple


ROOT = Path(__file__).resolve().parent.parent
BLOG_DIR = ROOT / "data" / "markdown" / "blog"
MANIFEST = ROOT / "data" / "blogPostManifest.json"
METADATA = ROOT / "data" / "blogPostTranslations.fr.json"
OLLAMA_URL = "http://127.0.0.1:11434/api/generate"
MAX_CHUNK_CHARS = 3200

PROTECTED_PATTERN = re.compile(
    r"<[^>]+>|`[^`\n]+`|https?://[^\s)\]>'\"]+|(?<=\()/(?:static|blog)/[^\s)]+"
)
REFERENCE_BLOCK = re.compile(r"(?m)^(?:\s*\[[^\]]+\]:\s*\S+\s*)+$")
MEDIA_BLOCK = re.compile(
    r"(?is)^\s*(?:(?:<!--.*?-->)|(?:<(?:img|iframe|video|source|p)\b[^>]*>(?:</(?:iframe|video|p)>)?))\s*$"
)


def ollama_generate(model: str, prompt: str, json_output: bool = False) -> str:
    request_body = {
        "model": model,
        "stream": False,
        "options": {"temperature": 0, "num_ctx": 8192},
        "prompt": prompt,
    }
    if json_output:
        request_body["format"] = "json"
    request = urllib.request.Request(
        OLLAMA_URL,
        data=json.dumps(request_body).encode("utf-8"),
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(request, timeout=300) as response:
            result = json.load(response)
    except (urllib.error.URLError, TimeoutError) as error:
        raise RuntimeError(
            "Ollama is unavailable. Start it and install the requested model first."
        ) from error
    if result.get("done_reason") == "length":
        raise RuntimeError("translation response was truncated")
    return result.get("response", "").strip()


def protect_markup(source: str) -> Tuple[str, Dict[str, str]]:
    protected: Dict[str, str] = {}

    def replace(match) -> str:
        token = "__KEEP_{:04d}__".format(len(protected))
        protected[token] = match.group(0)
        return token

    return PROTECTED_PATTERN.sub(replace, source), protected


def restore_markup(translated: str, protected: Dict[str, str]) -> str:
    for token, original in protected.items():
        if translated.count(token) != 1:
            raise RuntimeError("translation changed protected token " + token)
        translated = translated.replace(token, original)
    return translated


def should_copy_block(block: str) -> bool:
    stripped = block.strip()
    if not stripped:
        return True
    if REFERENCE_BLOCK.fullmatch(stripped) or MEDIA_BLOCK.fullmatch(stripped):
        return True
    lines = [line for line in block.splitlines() if line.strip()]
    return bool(lines) and all(line.startswith("    ") or line.startswith("\t") for line in lines)


def source_blocks(source: str) -> List[Tuple[str, str]]:
    """Return (content, following whitespace) pairs without losing layout."""
    parts = re.split(r"(\n\s*\n)", source)
    blocks: List[Tuple[str, str]] = []
    for index in range(0, len(parts), 2):
        content = parts[index]
        separator = parts[index + 1] if index + 1 < len(parts) else ""
        blocks.append((content, separator))
    return blocks


def grouped_blocks(blocks: Iterable[Tuple[str, str]]) -> Iterable[List[Tuple[str, str]]]:
    group: List[Tuple[str, str]] = []
    size = 0
    for block in blocks:
        block_size = len(block[0]) + len(block[1])
        if group and size + block_size > MAX_CHUNK_CHARS:
            yield group
            group = []
            size = 0
        group.append(block)
        size += block_size
    if group:
        yield group


def translate_chunk(model: str, source: str) -> str:
    protected_source, protected = protect_markup(source)
    placeholder_requirement = ""
    if protected:
        placeholder_requirement = "- Preserve every placeholder token exactly once and unchanged.\n"
    prompt = """You are a meticulous Canadian French translator.
Translate every human-readable English phrase in the Markdown below into natural French suitable for a personal technical blog. Return only the translated Markdown: no preface, notes, or surrounding code fence.

Requirements:
- Preserve Markdown structure, paragraph order, and blank lines.
{}\
- Preserve commands, code, identifiers, units, filenames, product/company names, and people/place names.
- Do not summarize, omit, add, or rewrite facts.
- Use proper French typography and idiomatic Canadian French.

MARKDOWN:
""".format(placeholder_requirement) + protected_source
    for attempt in range(3):
        try:
            translated = ollama_generate(model, prompt)
            if translated.startswith("MARKDOWN:"):
                translated = translated[len("MARKDOWN:") :].lstrip()
            if translated.lower().startswith("here is the translated markdown"):
                translated = translated.split("\n", 1)[1].lstrip()
            if translated.startswith("```markdown") and translated.endswith("```"):
                translated = translated[len("```markdown") : -len("```")].strip()
            translated = restore_markup(translated, protected)
            if re.search(r"__KEEP_", translated):
                raise RuntimeError("translation leaked an internal placeholder")
            if re.search(r"(?im)^note: i(?:'|’)ve (?:kept|preserved).*$", translated):
                raise RuntimeError("translation added an explanatory note")
            if len(translated) < max(1, len(source) // 3):
                raise RuntimeError("translation appears to have omitted content")
            return translated
        except RuntimeError:
            if attempt == 2:
                raise
    raise RuntimeError("unreachable")


def translate_markdown(model: str, source: str) -> str:
    output: List[str] = []
    for group in grouped_blocks(source_blocks(source)):
        translatable = [block for block in group if not should_copy_block(block[0])]
        if not translatable:
            output.extend(content + separator for content, separator in group)
            continue

        # Translate each grouped region with explicit boundary tokens, then put
        # back the source whitespace. Boundaries prevent paragraphs from merging.
        boundary = "\n__BLOG_PARAGRAPH_BOUNDARY__\n"
        joined = boundary.join(block[0] for block in translatable)
        translated = translate_chunk(model, joined)
        pieces = translated.split(boundary)
        if len(pieces) != len(translatable):
            # A smaller retry is slower but protects the file structure.
            pieces = [translate_chunk(model, block[0]) for block in translatable]
        piece_index = 0
        for content, separator in group:
            if should_copy_block(content):
                output.append(content + separator)
            else:
                output.append(pieces[piece_index].strip() + separator)
                piece_index += 1
    return "".join(output).rstrip() + "\n"


def translate_metadata(model: str, posts: List[dict], existing: dict, force: bool) -> dict:
    pending = [post for post in posts if force or str(post["id"]) not in existing]
    for offset in range(0, len(pending), 12):
        batch = pending[offset : offset + 12]
        source = {
            str(post["id"]): {
                "title": post["title"],
                "description": post.get("fbDescription", ""),
            }
            for post in batch
        }
        prompt = """Translate the human-readable values in this JSON object from English to natural Canadian French. Keep every key and the exact object structure. Preserve names, brands, technical identifiers, and numbers. Empty strings must stay empty. Return JSON only.

""" + json.dumps(source, ensure_ascii=False)
        result = json.loads(ollama_generate(model, prompt, json_output=True))
        if set(result) != set(source):
            raise RuntimeError("metadata translation changed post IDs")
        for post_id, values in result.items():
            title = str(values.get("title", "")).strip()
            description = str(values.get("description", "")).strip()
            if not title:
                raise RuntimeError("metadata translation omitted title for post " + post_id)
            existing[post_id] = {"title": title}
            if description:
                existing[post_id]["description"] = description
        METADATA.write_text(
            json.dumps(existing, ensure_ascii=False, indent=2, sort_keys=True) + "\n",
            encoding="utf-8",
        )
        print("Translated metadata: {}/{}".format(min(offset + 12, len(pending)), len(pending)), flush=True)
    return existing


def selected_posts(posts: List[dict], post_ids: List[int]) -> List[dict]:
    if not post_ids:
        return posts
    wanted = set(post_ids)
    selected = [post for post in posts if post["id"] in wanted]
    missing = wanted - {post["id"] for post in selected}
    if missing:
        raise RuntimeError("unknown post IDs: " + ", ".join(map(str, sorted(missing))))
    return selected


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--post", type=int, action="append", default=[], help="translate one post ID; repeat as needed")
    parser.add_argument("--model", default="llama3:latest", help="local Ollama model name")
    parser.add_argument("--force", action="store_true", help="replace existing generated translations")
    parser.add_argument("--metadata-only", action="store_true")
    parser.add_argument("--body-only", action="store_true")
    args = parser.parse_args()
    if args.metadata_only and args.body_only:
        parser.error("--metadata-only and --body-only cannot be combined")

    posts = selected_posts(json.loads(MANIFEST.read_text(encoding="utf-8")), args.post)
    existing = json.loads(METADATA.read_text(encoding="utf-8")) if METADATA.exists() else {}

    try:
        for index, post in enumerate(posts, 1):
            # Work newest-first and finish each post before moving on. If a run
            # is interrupted, the most useful recent translations are complete.
            if not args.body_only:
                translate_metadata(args.model, [post], existing, args.force)
            if not args.metadata_only:
                source_path = BLOG_DIR / (str(post["id"]) + ".markdown")
                target_path = BLOG_DIR / (str(post["id"]) + ".fr.markdown")
                if target_path.exists() and not args.force:
                    continue
                translated = translate_markdown(args.model, source_path.read_text(encoding="utf-8"))
                target_path.write_text(translated, encoding="utf-8")
                print("Translated body {}/{}: {}".format(index, len(posts), target_path.name), flush=True)
    except (RuntimeError, ValueError, json.JSONDecodeError) as error:
        print("Translation failed: " + str(error), file=sys.stderr)
        raise SystemExit(1)


if __name__ == "__main__":
    main()
