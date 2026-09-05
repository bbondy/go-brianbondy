#!/usr/bin/env python3
"""
Script to fetch GitHub statistics for projects and update the project manifest.

When a GitHub token is available (GITHUB_TOKEN or GH_TOKEN) the counts come from
the GitHub REST search API, which is what CI uses. Without a token the script
falls back to scraping the public search pages, which works from a residential
IP but is unreliable from shared/CI addresses.
"""

import os
import re
import requests
import time
from urllib.parse import quote, urlparse
from typing import Dict, Any, Optional, Tuple

GITHUB_API_URL = "https://api.github.com"
GITHUB_TOKEN = os.environ.get("GITHUB_TOKEN") or os.environ.get("GH_TOKEN")

# The authenticated search API allows 30 requests per minute.
SEARCH_API_DELAY_SECONDS = 2.5

def parse_abbreviated_number(s: str) -> int:
    s = s.replace(',', '').strip()
    match = re.match(r'([\d.]+)\s*([kKmMbB]?)', s)
    if not match:
        return 0
    num = float(match.group(1))
    suffix = match.group(2).lower()
    if suffix == 'k':
        num *= 1_000
    elif suffix == 'm':
        num *= 1_000_000
    elif suffix == 'b':
        num *= 1_000_000_000
    return int(num)

def make_request_with_retry(url: str, headers: dict, max_retries: int = 3) -> requests.Response:
    """Make a request with exponential backoff retry logic for rate limiting."""
    for attempt in range(max_retries):
        try:
            response = requests.get(url, headers=headers, timeout=10)
            if response.status_code == 429 and attempt < max_retries - 1:
                # Rate limited, wait with exponential backoff
                wait_time = 2 ** attempt  # 1, 2, 4 seconds
                print(f"  Rate limited, waiting {wait_time} seconds before retry...")
                time.sleep(wait_time)
                continue
            return response
        except Exception as e:
            if attempt < max_retries - 1:
                wait_time = 2 ** attempt
                print(f"  Request failed, waiting {wait_time} seconds before retry: {e}")
                time.sleep(wait_time)
                continue
            raise
    return response

def extract_repo_info(github_url: str) -> Tuple[str, str]:
    parsed = urlparse(github_url)
    if parsed.netloc != "github.com":
        raise ValueError(f"Invalid GitHub URL: {github_url}")
    path_parts = parsed.path.strip("/").split("/")
    if len(path_parts) < 2:
        raise ValueError(f"Invalid GitHub URL: {github_url}")
    return path_parts[0], path_parts[1]

def build_search_query(owner: str, repo: str, author: str,
                       keywords: list | None = None, extra: str | None = None) -> str:
    """Build a GitHub search query string shared by the API and scraping paths."""
    parts = [f"repo:{owner}/{repo}", f"author:{author}"]
    if extra:
        parts.append(extra)
    if keywords:
        parts.append("(" + " OR ".join(f'"{keyword}"' for keyword in keywords) + ")")
    return " ".join(parts)

def search_api_count(endpoint: str, query: str, accept: str,
                     max_retries: int = 3) -> Optional[int]:
    """Return total_count from a search API endpoint, or None if it could not be read."""
    if not GITHUB_TOKEN:
        return None

    url = f"{GITHUB_API_URL}{endpoint}"
    headers = {
        "Accept": accept,
        "Authorization": f"Bearer {GITHUB_TOKEN}",
        "X-GitHub-Api-Version": "2022-11-28",
        "User-Agent": "go-brianbondy-stats",
    }
    params = {"q": query, "per_page": 1}

    for attempt in range(max_retries):
        try:
            response = requests.get(url, headers=headers, params=params, timeout=15)
        except Exception as e:
            if attempt < max_retries - 1:
                wait_time = 2 ** attempt
                print(f"  Search API request failed, retrying in {wait_time}s: {e}")
                time.sleep(wait_time)
                continue
            print(f"  Search API request failed: {e}")
            return None

        if response.status_code in (403, 429):
            # Primary or secondary rate limit. Honour Retry-After when present.
            wait_time = int(response.headers.get("Retry-After", 2 ** (attempt + 3)))
            if attempt < max_retries - 1:
                print(f"  Search API rate limited, waiting {wait_time}s before retry...")
                time.sleep(wait_time)
                continue
            print(f"  Search API rate limited: {response.text[:200]}")
            return None

        if response.status_code != 200:
            print(f"  Search API returned {response.status_code}: {response.text[:200]}")
            return None

        time.sleep(SEARCH_API_DELAY_SECONDS)
        return response.json().get("total_count")

    return None

def scrape_search_count(search_url: str, fallback_noun: str, label: str) -> int:
    """Scrape the public search results page for a result count."""
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
    }
    try:
        response = make_request_with_retry(search_url, headers)
        response.raise_for_status()
        content = response.text

        # Check if we got a rate limit page
        if "Too many requests" in content:
            print(f"  Rate limited for {label}")
            return 0

        # Try the robust h2 pattern first, then fallback patterns
        match = re.search(r'<h2 id="search-results-count"[^>]*>([\d.,a-zA-Z]+) results', content)
        if not match:
            match = re.search(rf'([\d.,a-zA-Z]+) {fallback_noun}', content)
        if not match:
            match = re.search(r'([\d.,a-zA-Z]+) results', content)
        if match:
            return parse_abbreviated_number(match.group(1))
        print(f"Could not find count on search page for {label}")
        return 0
    except Exception as e:
        print(f"Error getting count from search page for {label}: {e}")
        return 0
    finally:
        time.sleep(2)  # Be respectful to GitHub's servers

def scrape_url_for(query: str, search_type: str) -> str:
    return f"https://github.com/search?q={quote(query)}&type={search_type}"

def get_commit_count(owner: str, repo: str, author: str, keywords: list | None = None) -> int:
    query = build_search_query(owner, repo, author, keywords)
    count = search_api_count(
        "/search/commits", query, "application/vnd.github.cloak-preview+json")
    if count is not None:
        return count
    return scrape_search_count(
        scrape_url_for(query, "commits"), "commits", f"{owner}/{repo} commits")

def get_pr_count(owner: str, repo: str, author: str, keywords: list | None = None) -> int:
    query = build_search_query(owner, repo, author, keywords, extra="type:pr")
    count = search_api_count(
        "/search/issues", query, "application/vnd.github+json")
    if count is not None:
        return count
    return scrape_search_count(
        scrape_url_for(build_search_query(owner, repo, author, keywords), "pullrequests"),
        "pull requests", f"{owner}/{repo} PRs")

def get_project_keywords(project: dict) -> list | None:
    """Get keywords for specific projects to avoid double counting."""
    return project.get("searchKeywords")

def fetch_github_stats(github_url: str, author: str, project: dict | None = None) -> Dict[str, Any]:
    try:
        owner, repo = extract_repo_info(github_url)
        print(f"Fetching stats for {owner}/{repo}...")
        
        keywords = get_project_keywords(project) if project else None
        if keywords:
            print(f"  Using keywords: {', '.join(keywords)}")
        
        commit_count = get_commit_count(owner, repo, author, keywords)
        pr_count = get_pr_count(owner, repo, author, keywords)
        time.sleep(1)
        return {
            "commitCount": commit_count,
            "prCount": pr_count
        }
    except Exception as e:
        print(f"Error fetching GitHub stats for {github_url}: {e}")
        return {
            "commitCount": 0,
            "prCount": 0
        }

def update_project_manifest():
    import json
    manifest_path = "data/projectManifest.json"
    with open(manifest_path, 'r') as f:
        projects = json.load(f)
    updated = False
    with_github = 0
    with_stats = 0
    for project in projects:
        if project.get("github"):
            with_github += 1
            print(f"\nProcessing: {project['title']}")
            print(f"GitHub URL: {project['github']}")

            # Force re-parse if current commitCount is 70 (likely a default/fallback value)
            force_update = False
            if project.get("githubStats", {}).get("commitCount") == 70:
                print("  Force re-parsing (current commitCount is 70)")
                force_update = True

            stats = fetch_github_stats(project["github"], "bbondy", project)
            if stats["commitCount"] > 0 or stats["prCount"] > 0:
                with_stats += 1
                # Only include githubStats if we have real data
                if project.get("githubStats") != stats:
                    project["githubStats"] = stats
                    updated = True
                print(f"  - Commits: {stats['commitCount']}")
                print(f"  - PRs: {stats['prCount']}")
            elif force_update:
                # Remove githubStats if we forced update but got no data
                if "githubStats" in project:
                    del project["githubStats"]
                    updated = True
                print("  - Removed githubStats (no data available)")
            else:
                print("  - No stats found")
    if updated:
        with open(manifest_path, 'w') as f:
            json.dump(projects, f, indent=2, ensure_ascii=False)
            f.write("\n")
        print(f"\nUpdated {manifest_path} with GitHub statistics")
    else:
        print("\nNo updates made")

    # A total wipeout means the search backend rejected us rather than the
    # projects genuinely having no activity. Fail loudly so CI surfaces it.
    if with_github and not with_stats:
        raise SystemExit(
            f"No stats could be fetched for any of the {with_github} GitHub projects")

if __name__ == "__main__":
    update_project_manifest()
