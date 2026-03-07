#!/usr/bin/env python3
"""
Script to fetch time, distance, and elevation information from Strava activities and update the run manifest.
This script will:
1. Read the current memorableRuns.json
2. For each activity with Strava URLs, fetch the time, distance, and elevation information
3. Add 'time', 'distance', and 'elevation' fields to each activity
4. Update the manifest file
"""

import os
import json
import re
import requests
import time
from urllib.parse import urlparse, parse_qs
from typing import Dict, List, Optional, Tuple
import webbrowser
import threading
from http.server import HTTPServer, BaseHTTPRequestHandler
# Note: We removed unconditional BeautifulSoup imports because the script now relies solely
# on the Strava REST API. The legacy HTML-scraping helper, kept for possible fallback use,
# will perform its own lazy import if it is ever called.

# We no longer scrape HTML, so BeautifulSoup is not required.
# Instead, we will use the official Strava v3 REST API, which requires
# an access token (scope "activity:read_all") to read activity details.

# Read the access token once at startup. You can create a personal token
# quickly in Strava settings or reuse the OAuth flow used by other scripts
# in this repository. Place it in the environment as STRAVA_ACCESS_TOKEN.
STRAVA_ACCESS_TOKEN = os.environ.get("STRAVA_ACCESS_TOKEN")
STRAVA_CLIENT_ID = os.environ.get("STRAVA_CLIENT_ID")
STRAVA_CLIENT_SECRET = os.environ.get("STRAVA_CLIENT_SECRET")
STRAVA_REFRESH_TOKEN = os.environ.get("STRAVA_REFRESH_TOKEN")

# OAuth/Token constants (reused from generate_strava_run_manifest.py)
TOKEN_PATH = os.path.expanduser('~/.strava_token.json')
REDIRECT_URI = 'http://localhost:8080/'
AUTH_URL = 'https://www.strava.com/oauth/authorize'
TOKEN_URL = 'https://www.strava.com/oauth/token'

# Local HTTP server handler to capture OAuth redirect
class _OAuthHandler(BaseHTTPRequestHandler):
    code: Optional[str] = None
    def do_GET(self):
        parsed = urlparse(self.path)
        params = parse_qs(parsed.query)
        if 'code' in params:
            _OAuthHandler.code = params['code'][0]
            self.send_response(200)
            self.send_header('Content-type', 'text/html')
            self.end_headers()
            self.wfile.write(b'<h1>Authorization successful! You can close this window.</h1>')
        else:
            self.send_response(400)
            self.end_headers()
            self.wfile.write(b'<h1>Authorization failed.</h1>')


def _start_oauth_server():
    server = HTTPServer(('localhost', 8080), _OAuthHandler)
    thread = threading.Thread(target=server.handle_request)
    thread.start()
    return server


def _get_access_token_via_oauth(client_id: str, client_secret: str) -> str:
    """Launch browser OAuth flow and return fresh access token."""
    params = {
        'client_id': client_id,
        'redirect_uri': REDIRECT_URI,
        'response_type': 'code',
        'scope': 'activity:read_all',
        'approval_prompt': 'auto',
    }
    auth_url = AUTH_URL + '?' + '&'.join(f'{k}={v}' for k, v in params.items())
    print('Opening browser for Strava login...')
    webbrowser.open(auth_url)
    _start_oauth_server()
    while _OAuthHandler.code is None:
        time.sleep(0.2)
    code = _OAuthHandler.code
    data = {
        'client_id': client_id,
        'client_secret': client_secret,
        'code': code,
        'grant_type': 'authorization_code',
    }
    resp = requests.post(TOKEN_URL, data=data, timeout=10)
    resp.raise_for_status()
    token_data = resp.json()
    with open(TOKEN_PATH, 'w') as f:
        json.dump(token_data, f)
    print('Access token saved to', TOKEN_PATH)
    return token_data['access_token']


def _refresh_access_token(refresh_token: Optional[str] = None) -> Optional[str]:
    """Use refresh-token grant to obtain a new Strava access token."""
    global STRAVA_ACCESS_TOKEN, STRAVA_REFRESH_TOKEN
    refresh_token = refresh_token or STRAVA_REFRESH_TOKEN
    if not (STRAVA_CLIENT_ID and STRAVA_CLIENT_SECRET and refresh_token):
        return None

    data = {
        'client_id': STRAVA_CLIENT_ID,
        'client_secret': STRAVA_CLIENT_SECRET,
        'grant_type': 'refresh_token',
        'refresh_token': refresh_token,
    }
    try:
        resp = requests.post(TOKEN_URL, data=data, timeout=10)
        if resp.status_code != 200:
            print(f"Failed to refresh access token: {resp.status_code} – {resp.text[:200]}")
            return None
        token_data = resp.json()
        STRAVA_ACCESS_TOKEN = token_data.get('access_token')
        STRAVA_REFRESH_TOKEN = token_data.get('refresh_token', STRAVA_REFRESH_TOKEN)
        with open(TOKEN_PATH, 'w') as f:
            json.dump(token_data, f)
        print('Access token refreshed and saved to', TOKEN_PATH)
        return STRAVA_ACCESS_TOKEN
    except Exception as e:
        print(f"Failed to refresh access token: {e}")
        return None


def _load_saved_token() -> Optional[str]:
    if os.path.exists(TOKEN_PATH):
        with open(TOKEN_PATH, 'r') as f:
            data = json.load(f)
            return data.get('access_token')
    return None


def ensure_access_token() -> Optional[str]:
    """Return a valid Strava access token, obtaining one via OAuth if necessary."""
    global STRAVA_ACCESS_TOKEN
    if STRAVA_ACCESS_TOKEN:
        return STRAVA_ACCESS_TOKEN
    # Try saved token
    saved = _load_saved_token()
    if saved:
        STRAVA_ACCESS_TOKEN = saved
        return STRAVA_ACCESS_TOKEN
    refreshed = _refresh_access_token()
    if refreshed:
        return refreshed
    # Fallback to OAuth flow
    client_id = STRAVA_CLIENT_ID or input('Enter your Strava client_id: ')
    client_secret = STRAVA_CLIENT_SECRET or input('Enter your Strava client_secret: ')
    STRAVA_ACCESS_TOKEN = _get_access_token_via_oauth(client_id, client_secret)
    return STRAVA_ACCESS_TOKEN

# Base URL for Strava activity details
STRAVA_API_ACTIVITY_URL = "https://www.strava.com/api/v3/activities/"

# Helper: seconds → "Xh Ym" / "Xm"
def seconds_to_time_str(seconds: int) -> str:
    minutes = seconds // 60
    hours = minutes // 60
    minutes = minutes % 60
    if hours > 0 and minutes > 0:
        return f"{hours}h {minutes}m"
    elif hours > 0:
        return f"{hours}h"
    else:
        return f"{minutes}m"

# Fetch metrics for a single activity ID via the Strava API
def fetch_strava_activity_metrics_from_api(activity_id: str, access_token: str) -> Optional[Dict[str, str]]:
    """Return {'time', 'distance', 'elevation'} strings or None on failure."""
    url = STRAVA_API_ACTIVITY_URL + activity_id
    headers = {"Authorization": f"Bearer {access_token}"}
    try:
        resp = requests.get(url, headers=headers, timeout=10)
        if resp.status_code == 401:
            refreshed = _refresh_access_token()
            if refreshed:
                headers = {"Authorization": f"Bearer {refreshed}"}
                resp = requests.get(url, headers=headers, timeout=10)
        if resp.status_code != 200:
            print(f"Failed to fetch {activity_id} via API: {resp.status_code} – {resp.text[:200]}")
            return None
        data = resp.json()

        metrics: Dict[str, str] = {}

        # Time (moving_time in seconds)
        moving_time = data.get("moving_time", 0)
        if moving_time:
            metrics["time"] = seconds_to_time_str(moving_time)

        # Distance (meters → miles)
        distance_m = data.get("distance", 0.0)
        if distance_m:
            miles = distance_m / 1609.34
            metrics["distance"] = miles_to_distance_string(miles)

        # Elevation gain (meters → feet)
        elevation_m = data.get("total_elevation_gain", 0.0)
        if elevation_m:
            feet = int(round(elevation_m * 3.28084))
            metrics["elevation"] = feet_to_elevation_string(feet)

        return metrics if metrics else None
    except Exception as e:
        print(f"Exception fetching activity {activity_id}: {e}")
    return None

def extract_activity_id_from_url(url: str) -> Optional[str]:
    """Extract activity ID from Strava URL."""
    if 'strava.com/activities/' in url:
        match = re.search(r'/activities/(\d+)', url)
        if match:
            return match.group(1)
    return None

def fetch_strava_activity_metrics_from_web(url: str) -> Optional[Dict[str, str]]:
    """Legacy fallback that scrapes the Strava web page (unused in normal operation).

    This function now performs a lazy import of BeautifulSoup so that the main
    script has no hard dependency on `bs4` when only the API path is used.
    """
    # Function body removed – legacy scraper no longer supported.

def time_string_to_minutes(time_str: str) -> int:
    """Convert time string like '12h 34m' or '34m' to total minutes."""
    if not time_str:
        return 0

    total_minutes = 0
    # Match hours and minutes
    h_match = re.search(r'(\d+)h', time_str)
    m_match = re.search(r'(\d+)m', time_str)

    if h_match:
        total_minutes += int(h_match.group(1)) * 60
    if m_match:
        total_minutes += int(m_match.group(1))

    return total_minutes

def minutes_to_time_string(minutes: int) -> str:
    """Convert total minutes to 'Xh Ym' format."""
    if minutes == 0:
        return "0m"

    hours = minutes // 60
    remaining_minutes = minutes % 60

    if hours > 0 and remaining_minutes > 0:
        return f"{hours}h {remaining_minutes}m"
    elif hours > 0:
        return f"{hours}h"
    else:
        return f"{remaining_minutes}m"

def distance_string_to_miles(distance_str: str) -> float:
    """Convert distance string to miles."""
    if not distance_str:
        return 0.0

    # Match distance with units
    match = re.search(r'([\d.]+)\s*(mi|km)', distance_str, re.IGNORECASE)
    if match:
        value = float(match.group(1))
        unit = match.group(2).lower()
        if unit == 'km':
            return value * 0.621371  # Convert km to miles
        else:
            return value
    return 0.0

def miles_to_distance_string(miles: float) -> str:
    """Convert miles to distance string."""
    if miles == 0:
        return "0mi"

    if miles >= 1:
        return f"{miles:.1f}mi"
    else:
        return f"{miles:.2f}mi"

def elevation_string_to_feet(elevation_str: str) -> int:
    """Convert elevation string to feet."""
    if not elevation_str:
        return 0

    # Match elevation with units
    match = re.search(r'([\d,]+)\s*(ft|m)', elevation_str, re.IGNORECASE)
    if match:
        value = int(match.group(1).replace(',', ''))
        unit = match.group(2).lower()
        if unit == 'm':
            return int(value * 3.28084)  # Convert meters to feet
        else:
            return value
    return 0

def feet_to_elevation_string(feet: int) -> str:
    """Convert feet to elevation string."""
    if feet == 0:
        return "0ft"

    if feet >= 1000:
        return f"{feet:,}ft"
    else:
        return f"{feet}ft"

def fetch_total_metrics_from_multiple_strava_urls(urls: List[str]) -> Optional[Dict[str, str]]:
    """Aggregate metrics for a list of Strava activity URLs using the v3 API."""

    access_token = ensure_access_token()
    if not access_token:
        print("Unable to obtain Strava access token – skipping API metrics fetch.")
        return None

    total_minutes = 0
    total_miles = 0.0
    total_feet = 0
    found_metrics = 0

    for url in urls:
        if 'strava.com/activities/' not in url:
            continue

        activity_id = extract_activity_id_from_url(url)
        if not activity_id:
            continue

        metrics = fetch_strava_activity_metrics_from_api(activity_id, access_token)
        access_token = STRAVA_ACCESS_TOKEN or access_token
        if not metrics:
            continue

        found_metrics += 1

        # Add up time
        if 'time' in metrics:
            minutes = time_string_to_minutes(metrics['time'])
            total_minutes += minutes
            print(f"  - {url}: {metrics.get('time', 'N/A')} ({minutes} minutes)")

        # Add up distance
        if 'distance' in metrics:
            miles = distance_string_to_miles(metrics['distance'])
            total_miles += miles
            print(f"    Distance: {metrics.get('distance', 'N/A')} ({miles:.1f} miles)")

        # Add up elevation
        if 'elevation' in metrics:
            feet = elevation_string_to_feet(metrics['elevation'])
            total_feet += feet
            print(f"    Elevation: {metrics.get('elevation', 'N/A')} ({feet} feet)")

        # Be nice to Strava API – small pause to avoid hitting rate limits
        time.sleep(0.2)

    if found_metrics == 0:
        return None

    total_metrics: Dict[str, str] = {}
    if total_minutes > 0:
        total_metrics['time'] = minutes_to_time_string(total_minutes)
    if total_miles > 0:
        total_metrics['distance'] = miles_to_distance_string(total_miles)
    if total_feet > 0:
        total_metrics['elevation'] = feet_to_elevation_string(total_feet)

    print(f"  Total: {total_metrics.get('time', 'N/A')}, {total_metrics.get('distance', 'N/A')}, {total_metrics.get('elevation', 'N/A')} (from {found_metrics} activities)")
    return total_metrics

def parse_time_from_description(description: str) -> Optional[str]:
    """Extract time information from existing descriptions."""
    if not description:
        return None

    # Common time patterns in descriptions
    time_patterns = [
        r'(\d+)\s*h\s*(\d+)\s*m',  # 80h 37m
        r'(\d+)\s*hours?\s*(\d+)\s*minutes?',  # 80 hours 37 minutes
        r'(\d+)\s*hours?',  # 80 hours
        r'(\d+)\s*minutes?',  # 37 minutes
        r'~(\d+)\s*hours?',  # ~91 hours
        r'(\d+)\s*h,\s*(\d+)\s*m',  # 25h, 25m
        r'(\d+)\s*hours?,\s*(\d+)\s*minutes?',  # 25 hours, 25 minutes
        r'(\d+)\s*h\s*and\s*(\d+)\s*loops?',  # 30h and 30 loops
    ]

    for pattern in time_patterns:
        match = re.search(pattern, description, re.IGNORECASE)
        if match:
            if len(match.groups()) == 2:
                hours = int(match.group(1))
                minutes = int(match.group(2))
                return f"{hours}h {minutes}m"
            elif len(match.groups()) == 1:
                hours = int(match.group(1))
                return f"{hours}h"

    return None

def clean_description(description: str, time_from_desc: Optional[str]) -> Optional[str]:
    """
    Remove time information from the description. If only time is left, remove the description entirely.
    """
    if not description or not time_from_desc:
        return description

    # Remove the time phrase and any leading/trailing conjunctions or punctuation
    # Remove patterns like '30h and 30 loops', 'Finished in 25 hours, 25 minutes', etc.
    cleanup_patterns = [
        r'(?i)\bFinished in\s*\d+\s*hours?,?\s*\d*\s*minutes?\b',
        r'(?i)\bFinished in\s*\d+\s*hours?\b',
        r'(?i)~?\d+\s*hours?\b',
        r'(?i)\d+\s*h\s*\d+\s*m',
        r'(?i)\d+\s*hours?\s*\d+\s*minutes?',
        r'(?i)\d+\s*hours?,\s*\d+\s*minutes?',
        r'(?i)\d+\s*h,\s*\d+\s*m',
        r'(?i)\d+\s*h\s*and\s*\d+\s*loops?',
        r'(?i)\d+\s*h',
        r'(?i)\d+\s*minutes?',
    ]
    cleaned = description
    for pattern in cleanup_patterns:
        cleaned = re.sub(pattern, '', cleaned, flags=re.IGNORECASE)
    # Remove extra spaces, commas, 'and', etc.
    cleaned = re.sub(r'^[,\s]*(and)?[,\s]*', '', cleaned)
    cleaned = re.sub(r'[,\s]*(and)?[,\s]*$', '', cleaned)
    cleaned = cleaned.strip(' ,.-')
    # If nothing is left, return None
    if not cleaned:
        return None
    return cleaned

def update_run_manifest():
    """Update the run manifest with time, distance, and elevation information."""

    # Read current manifest (if it exists)
    try:
        with open('data/memorableRuns.json', 'r') as f:
            runs = json.load(f)
    except FileNotFoundError:
        print("memorableRuns.json not found – starting a new manifest from scratch.")
        runs = []

    updated_runs = []

    for run in runs:
        # Check if time is already in description
        time_from_desc = parse_time_from_description(run.get('description', ''))

        if time_from_desc:
            # Add time field and clean up description
            run['time'] = time_from_desc
            desc = run.get('description', '')
            cleaned_desc = clean_description(desc, time_from_desc)
            if cleaned_desc:
                run['description'] = cleaned_desc
            else:
                run.pop('description', None)
        else:
            # Try to fetch from Strava if we have URLs and time is missing
            strava_urls = run.get('strava_urls', [])
            if strava_urls and (not run.get('time')):
                print(f"Processing '{run['title']}' with {len(strava_urls)} Strava URLs:")
                fetched_metrics = fetch_total_metrics_from_multiple_strava_urls(strava_urls)
                if fetched_metrics:
                    run.update(fetched_metrics)
                    print(f"Fetched metrics for '{run['title']}': {fetched_metrics}")
                else:
                    print(f"Activity '{run['title']}' needs manual review")
            # Also reprocess activities with multiple Strava URLs to ensure metrics are added up
            elif len(strava_urls) > 1:
                print(f"Reprocessing '{run['title']}' with {len(strava_urls)} Strava URLs to add up metrics:")
                fetched_metrics = fetch_total_metrics_from_multiple_strava_urls(strava_urls)
                if fetched_metrics:
                    run.update(fetched_metrics)
                    print(f"Updated metrics for '{run['title']}': {fetched_metrics}")
            # Also fetch metrics for single Strava URL activities that don't have distance/elevation
            elif strava_urls and (not run.get('distance') or not run.get('elevation')):
                print(f"Fetching missing metrics for '{run['title']}' with {len(strava_urls)} Strava URLs:")
                fetched_metrics = fetch_total_metrics_from_multiple_strava_urls(strava_urls)
                if fetched_metrics:
                    # Only update missing metrics, don't overwrite existing time
                    if not run.get('distance') and 'distance' in fetched_metrics:
                        run['distance'] = fetched_metrics['distance']
                    if not run.get('elevation') and 'elevation' in fetched_metrics:
                        run['elevation'] = fetched_metrics['elevation']
                    print(f"Updated missing metrics for '{run['title']}': {fetched_metrics}")

        updated_runs.append(run)

    # Write updated manifest
    with open('data/memorableRuns.json', 'w') as f:
        json.dump(updated_runs, f, indent=2)

    print(f"Updated {len(updated_runs)} activities in memorableRuns.json")

if __name__ == "__main__":
    update_run_manifest()
