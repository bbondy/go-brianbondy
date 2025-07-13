#!/usr/bin/env python3
"""
Script to fetch GitHub statistics for projects and update the project manifest.
This script scrapes GitHub pages to get commit counts and language breakdowns.
"""

import re
import requests
import time
from urllib.parse import urlparse, quote
from typing import Dict, Any, Tuple

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

def get_commit_count(owner: str, repo: str, author: str, keywords: list | None = None) -> int:
    if keywords:
        # Search with keywords using OR logic
        keyword_queries = []
        for keyword in keywords:
            keyword_queries.append(f'"{keyword}"')
        keyword_query = "+OR+".join(keyword_queries)
        url = f"https://github.com/search?q=repo%3A{owner}%2F{repo}+author%3A{author}+({keyword_query})&type=commits"
    else:
        url = f"https://github.com/search?q=repo%3A{owner}%2F{repo}+author%3A{author}&type=commits"
    
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
    }
    try:
        response = make_request_with_retry(url, headers)
        response.raise_for_status()
        content = response.text
        
        # Check if we got a rate limit page
        if "Too many requests" in content:
            print(f"  Rate limited for {owner}/{repo} commits")
            return 0
            
        # Try the robust h2 pattern first, then fallback patterns
        match = re.search(r'<h2 id="search-results-count"[^>]*>([\d.,a-zA-Z]+) results', content)
        if not match:
            match = re.search(r'([\d.,a-zA-Z]+) commits', content)
        if not match:
            match = re.search(r'([\d.,a-zA-Z]+) results', content)
        if match:
            count_str = match.group(1)
            result = parse_abbreviated_number(count_str)
            time.sleep(2)  # Be respectful to GitHub's servers
            return result
        else:
            print(f"Could not find commit count on search page for {owner}/{repo}")
            time.sleep(2)  # Be respectful to GitHub's servers
            return 0
    except Exception as e:
        print(f"Error getting commit count from search for {owner}/{repo}: {e}")
        time.sleep(2)  # Be respectful to GitHub's servers
        return 0

def get_pr_count(owner: str, repo: str, author: str, keywords: list | None = None) -> int:
    if keywords:
        # Search with keywords using OR logic
        keyword_queries = []
        for keyword in keywords:
            keyword_queries.append(f'"{keyword}"')
        keyword_query = "+OR+".join(keyword_queries)
        url = f"https://github.com/search?q=repo%3A{owner}%2F{repo}+author%3A{author}+({keyword_query})&type=pullrequests"
    else:
        url = f"https://github.com/search?q=repo%3A{owner}%2F{repo}+author%3A{author}&type=pullrequests"
    
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
    }
    try:
        response = make_request_with_retry(url, headers)
        response.raise_for_status()
        content = response.text
        
        # Check if we got a rate limit page
        if "Too many requests" in content:
            print(f"  Rate limited for {owner}/{repo} PRs")
            return 0
            
        # Try the robust h2 pattern first, then fallback patterns
        match = re.search(r'<h2 id="search-results-count"[^>]*>([\d.,a-zA-Z]+) results', content)
        if not match:
            match = re.search(r'([\d.,a-zA-Z]+) pull requests', content)
        if not match:
            match = re.search(r'([\d.,a-zA-Z]+) results', content)
        if match:
            count_str = match.group(1)
            result = parse_abbreviated_number(count_str)
            time.sleep(2)  # Be respectful to GitHub's servers
            return result
        else:
            print(f"Could not find PR count on search page for {owner}/{repo}")
            time.sleep(2)  # Be respectful to GitHub's servers
            return 0
    except Exception as e:
        print(f"Error getting PR count from search for {owner}/{repo}: {e}")
        time.sleep(2)  # Be respectful to GitHub's servers
        return 0

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
    for project in projects:
        if project.get("github"):
            print(f"\nProcessing: {project['title']}")
            print(f"GitHub URL: {project['github']}")
            
            # Force re-parse if current commitCount is 70 (likely a default/fallback value)
            force_update = False
            if project.get("githubStats", {}).get("commitCount") == 70:
                print(f"  Force re-parsing (current commitCount is 70)")
                force_update = True
            
            stats = fetch_github_stats(project["github"], "bbondy", project)
            if stats["commitCount"] > 0 or stats["prCount"] > 0:
                # Only include githubStats if we have real data
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
            json.dump(projects, f, indent=2)
        print(f"\nUpdated {manifest_path} with GitHub statistics")
    else:
        print("\nNo updates made")

if __name__ == "__main__":
    update_project_manifest() 