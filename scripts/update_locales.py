#!/usr/bin/env python3
"""Discover translatable UI strings and merge them into locale JSON files."""

import json
import re
from pathlib import Path
from typing import Set, Tuple


ROOT = Path(__file__).resolve().parent.parent
LOCALES = ROOT / "locales"
SOURCE_FILES = sorted((ROOT / "templates").glob("*.html")) + sorted(
    (ROOT / "static" / "js").glob("*.js")
)
TEMPLATE_PATTERN = re.compile(r'\{\{t\s+"((?:\\.|[^"\\])*)"')
CLIENT_PATTERN = re.compile(r"window\.site(?:T|Format)\('((?:\\.|[^'\\])*)'")

# Page copy also lives in data rather than in templates, so the resume, career,
# and projects pages need their prose pulled straight from those sources.
CAREER_SOURCE = ROOT / "data" / "career.go"
PROJECT_SOURCE = ROOT / "data" / "projectManifest.json"

# Only prose is translatable. Proper nouns (a person, company, school, or
# product name), URLs, and counts deliberately stay out of the catalog. Date
# ranges are included because they spell out month names, which have to read in
# the reader's language on the resume and career pages.
CAREER_TEXT_FIELDS = {
    ("CareerProfile", "Headline"),
    ("CareerProfile", "ExecutiveSummary"),
    ("CareerScaleProof", "Label"),
    ("CareerScaleProof", "SourceLabel"),
    ("CareerRole", "Title"),
    ("CareerRole", "Location"),
    ("CareerRole", "Dates"),
    ("CareerRole", "Summary"),
    ("CareerHighlight", "Text"),
    ("CareerHighlight", "SourceLabel"),
    ("CareerSkillGroup", "Name"),
    ("CareerSkillGroup", "Items"),
    ("CareerTimelineEntry", "Era"),
    ("CareerTimelineEntry", "Title"),
    ("CareerTimelineEntry", "Description"),
    ("CareerTimelineEntry", "Technologies"),
    ("CareerProject", "Description"),
    ("CareerEducation", "Degree"),
    ("CareerEducation", "Location"),
    ("CareerEducation", "Dates"),
    ("CareerEducation", "Coursework"),
    ("CareerEducation", "Development"),
}
CAREER_TEXT_SLICES = ("Interests", "ArchiveBackground")
PROJECT_TEXT_KEYS = ("title", "description")

COMPOSITE_TYPE_PATTERN = re.compile(r"(?:\[\])?(Career[A-Za-z]*)\s*$")
GO_FIELD_PATTERN = re.compile(r'([A-Z][A-Za-z]*):\s*"((?:\\.|[^"\\])*)"')
GO_SLICE_PATTERN = r'{}:\s*\[\]string\{{(.*?)\}}'
GO_SLICE_ITEM_PATTERN = re.compile(r'"((?:\\.|[^"\\])*)"')


def composite_owners(source: str):
    """Map each offset in career.go to the struct type whose literal encloses it.

    Elements of a slice literal are written bare, as `{Field: "value"}`, so they
    inherit the element type named on the enclosing `[]CareerX{` line.
    """
    owners = [None] * (len(source) + 1)
    stack = []
    in_string = False
    escaped = False
    for index, char in enumerate(source):
        if in_string:
            if escaped:
                escaped = False
            elif char == "\\":
                escaped = True
            elif char == '"':
                in_string = False
        elif char == '"':
            in_string = True
        elif char == "{":
            match = COMPOSITE_TYPE_PATTERN.search(source[:index])
            named = match.group(1) if match else (stack[-1] if stack else None)
            stack.append(named)
        elif char == "}" and stack:
            stack.pop()
        owners[index] = stack[-1] if stack else None
    return owners


def career_strings() -> Set[str]:
    if not CAREER_SOURCE.exists():
        return set()
    contents = CAREER_SOURCE.read_text(encoding="utf-8")
    owners = composite_owners(contents)
    strings = {
        decoded(match.group(2))
        for match in GO_FIELD_PATTERN.finditer(contents)
        if (owners[match.start()], match.group(1)) in CAREER_TEXT_FIELDS
    }
    for field in CAREER_TEXT_SLICES:
        block = re.search(GO_SLICE_PATTERN.format(field), contents, re.DOTALL)
        if block:
            strings.update(
                decoded_source(match)
                for match in GO_SLICE_ITEM_PATTERN.finditer(block.group(1))
            )
    return strings


def project_strings() -> Set[str]:
    if not PROJECT_SOURCE.exists():
        return set()
    projects = json.loads(PROJECT_SOURCE.read_text(encoding="utf-8"))
    return {
        value
        for project in projects
        for key in PROJECT_TEXT_KEYS
        if (value := str(project.get(key, "")).strip())
    }


def decoded(value: str) -> str:
    return value.replace(r'\"', '"').replace(r"\'", "'").replace(r"\\", "\\")


def decoded_source(match) -> str:
    return decoded(match.group(1))


def discover_strings() -> Tuple[Set[str], Set[str]]:
    strings: Set[str] = set()
    client_strings: Set[str] = set()
    for source_file in SOURCE_FILES:
        contents = source_file.read_text(encoding="utf-8")
        strings.update(decoded_source(match) for match in TEMPLATE_PATTERN.finditer(contents))
        discovered_client_strings = {
            decoded_source(match) for match in CLIENT_PATTERN.finditer(contents)
        }
        strings.update(discovered_client_strings)
        client_strings.update(discovered_client_strings)
    strings.update(career_strings())
    strings.update(project_strings())
    return strings, client_strings


def load_catalog(language: str):
    path = LOCALES / f"{language}.json"
    return json.loads(path.read_text(encoding="utf-8"))


def write_catalog(language: str, catalog) -> None:
    path = LOCALES / f"{language}.json"
    path.write_text(
        json.dumps(catalog, ensure_ascii=False, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )


def main() -> None:
    discovered, client_strings = discover_strings()
    english = load_catalog("en")
    french = load_catalog("fr")

    for source in discovered:
        english.setdefault(source, source)
        french.setdefault(source, "")

    write_catalog("en", english)
    write_catalog("fr", french)
    (LOCALES / "client.json").write_text(
        json.dumps(sorted(client_strings), ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )
    print(f"Updated {len(discovered)} deduplicated UI strings.")


if __name__ == "__main__":
    main()
