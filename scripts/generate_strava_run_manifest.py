import os
import requests
import json
import webbrowser
import threading
import time
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import urlparse, parse_qs

OUTPUT_PATH = 'data/stravaRunManifest.json'
TOKEN_PATH = os.path.expanduser('~/.strava_token.json')
STRAVA_API_URL = 'https://www.strava.com/api/v3/athlete/activities'
PER_PAGE = 200
REDIRECT_URI = 'http://localhost:8080/'
AUTH_URL = 'https://www.strava.com/oauth/authorize'
TOKEN_URL = 'https://www.strava.com/oauth/token'
# Strava may hand back a new refresh token on any refresh. When that happens in
# CI the stored secret is now stale, so drop the replacement here for the
# workflow to pick up.
ROTATED_TOKEN_PATH = 'strava_rotated_refresh_token.txt'


def non_interactive() -> bool:
    """True when there is no human around to complete the browser OAuth flow."""
    return bool(os.environ.get('CI') or os.environ.get('STRAVA_NON_INTERACTIVE'))


def require_interactive(action: str) -> None:
    """Abort instead of hanging on a browser prompt that nobody can answer."""
    if non_interactive():
        raise SystemExit(
            f"Cannot {action} without a browser. Set STRAVA_CLIENT_ID, "
            "STRAVA_CLIENT_SECRET and a valid STRAVA_REFRESH_TOKEN. "
            "Run scripts/strava_refresh_token.py locally to mint a new refresh token."
        )


# Helper to convert seconds to "Xh Ym" or "Xm"
def seconds_to_time_str(seconds):
    minutes = seconds // 60
    hours = minutes // 60
    minutes = minutes % 60
    if hours > 0:
        return f"{hours}h {minutes}m"
    else:
        return f"{minutes}m"

# Helper to calculate pace (min/km)
def calc_pace_min_per_km(moving_time_sec, distance_km):
    if moving_time_sec == 0 or distance_km == 0:
        return None
    pace = (moving_time_sec / 60) / distance_km
    mins = int(pace)
    secs = int(round((pace - mins) * 60))
    return f"{mins}:{secs:02d} min/km"

# Helper to format elevation
def format_elevation(total_elevation_gain):
    if total_elevation_gain is None or total_elevation_gain == 0:
        return ""
    # Strava returns elevation in meters
    return f"{int(total_elevation_gain)}m"

# HTTP handler for OAuth redirect
class OAuthHandler(BaseHTTPRequestHandler):
    code = None
    def do_GET(self):
        parsed = urlparse(self.path)
        params = parse_qs(parsed.query)
        if 'code' in params:
            OAuthHandler.code = params['code'][0]
            self.send_response(200)
            self.send_header('Content-type', 'text/html')
            self.end_headers()
            self.wfile.write(b'<h1>Authorization successful! You can close this window.</h1>')
        else:
            self.send_response(400)
            self.end_headers()
            self.wfile.write(b'<h1>Authorization failed.</h1>')

# Start local server to receive OAuth code
def get_oauth_code():
    server = HTTPServer(('localhost', 8080), OAuthHandler)
    thread = threading.Thread(target=server.handle_request)
    thread.start()
    return server

def get_access_token(client_id, client_secret):
    require_interactive('authorize with Strava')
    OAuthHandler.code = None
    # Step 1: Open browser for user login
    params = {
        'client_id': client_id,
        'redirect_uri': REDIRECT_URI,
        'response_type': 'code',
        'scope': 'read,activity:read_all',
        'approval_prompt': 'force',
    }
    url = AUTH_URL + '?' + '&'.join(f'{k}={v}' for k, v in params.items())
    print(f'Opening browser for Strava login...')
    webbrowser.open(url)
    # Step 2: Wait for redirect with code
    server = get_oauth_code()
    while OAuthHandler.code is None:
        time.sleep(0.2)
    code = OAuthHandler.code
    # Step 3: Exchange code for access token
    data = {
        'client_id': client_id,
        'client_secret': client_secret,
        'code': code,
        'grant_type': 'authorization_code',
    }
    resp = requests.post(TOKEN_URL, data=data)
    if resp.status_code != 200:
        raise Exception(f"Failed to get access token: {resp.status_code} {resp.text}")
    token_data = resp.json()
    # Save token for future use
    with open(TOKEN_PATH, 'w') as f:
        json.dump(token_data, f)
    print('Access token saved to', TOKEN_PATH)
    return token_data['access_token']

def refresh_access_token(refresh_token=None):
    client_id = os.environ.get('STRAVA_CLIENT_ID')
    client_secret = os.environ.get('STRAVA_CLIENT_SECRET')
    if refresh_token is None and os.path.exists(TOKEN_PATH):
        with open(TOKEN_PATH, 'r') as f:
            token_data = json.load(f)
        refresh_token = token_data.get('refresh_token')
    refresh_token = refresh_token or os.environ.get('STRAVA_REFRESH_TOKEN')
    if not (client_id and client_secret and refresh_token):
        return None

    data = {
        'client_id': client_id,
        'client_secret': client_secret,
        'grant_type': 'refresh_token',
        'refresh_token': refresh_token,
    }
    resp = requests.post(TOKEN_URL, data=data, timeout=10)
    if resp.status_code != 200:
        print(f"Failed to refresh access token: {resp.status_code} {resp.text}")
        return None

    token_data = resp.json()
    with open(TOKEN_PATH, 'w') as f:
        json.dump(token_data, f)
    print('Access token refreshed and saved to', TOKEN_PATH)

    new_refresh_token = token_data.get('refresh_token')
    if new_refresh_token and new_refresh_token != refresh_token:
        print('WARNING: Strava rotated the refresh token. '
              'The stored STRAVA_REFRESH_TOKEN is now stale and must be replaced.')
        with open(ROTATED_TOKEN_PATH, 'w') as f:
            f.write(new_refresh_token)

    return token_data.get('access_token')

def load_saved_token():
    if os.path.exists(TOKEN_PATH):
        with open(TOKEN_PATH, 'r') as f:
            token_data = json.load(f)
        # Optionally, check for expiration and refresh if needed
        return token_data.get('access_token')
    return None

def fetch_all_activities(access_token):
    page = 1
    activities = []
    headers = {'Authorization': f'Bearer {access_token}'}
    while True:
        params = {'per_page': PER_PAGE, 'page': page}
        resp = requests.get(STRAVA_API_URL, headers=headers, params=params)
        if resp.status_code == 401:
            # Token is invalid or expired
            raise AuthorizationError(f"Strava API authorization error: {resp.status_code} {resp.text}")
        elif resp.status_code != 200:
            raise Exception(f"Strava API error: {resp.status_code} {resp.text}")
        acts = resp.json()
        if not acts:
            break
        for act in acts:
            act_type = act.get('type', 'Unknown')
            date = act.get('start_date_local') or act.get('start_date')
            if date:
                date = date[:10]
            title = act.get('name')
            moving_time = act.get('moving_time', 0)
            time_str = seconds_to_time_str(moving_time)
            # Use actual distance from Strava for all activity types
            distance_km = round((act.get('distance', 0) / 1000), 2)
            if act_type == 'Run':
                pace = calc_pace_min_per_km(moving_time, distance_km)
            else:
                pace = ""
            activities.append({
                'date': date,
                'title': title,
                'distance_km': distance_km,
                'time': time_str,
                'pace': pace,
                'activity_id': str(act.get('id', '')),
                'type': act_type,
                'elevation': format_elevation(act.get('total_elevation_gain'))
            })
        page += 1
    return activities

# Custom exception for authorization errors
class AuthorizationError(Exception):
    pass

def main():
    client_id = os.environ.get('STRAVA_CLIENT_ID')
    client_secret = os.environ.get('STRAVA_CLIENT_SECRET')
    access_token = load_saved_token() or os.environ.get('STRAVA_ACCESS_TOKEN')
    if not access_token:
        access_token = refresh_access_token()
    if not access_token:
        access_token = get_access_token(client_id, client_secret)

    try:
        runs = fetch_all_activities(access_token)
    except AuthorizationError as e:
        print(f"Authorization failed: {e}")
        refreshed_token = refresh_access_token()
        if refreshed_token:
            print("Retrying with refreshed access token...")
            try:
                runs = fetch_all_activities(refreshed_token)
            except AuthorizationError as refreshed_error:
                print(f"Authorization still failing after refresh: {refreshed_error}")
                if os.path.exists(TOKEN_PATH):
                    print(f"Deleting stale token file: {TOKEN_PATH}")
                    os.remove(TOKEN_PATH)
                print("Re-authorizing with Strava to refresh scopes...")
                access_token = get_access_token(client_id, client_secret)
                runs = fetch_all_activities(access_token)
        else:
            if os.path.exists(TOKEN_PATH):
                print(f"Deleting expired token file: {TOKEN_PATH}")
                os.remove(TOKEN_PATH)
            print("Re-authorizing with Strava...")
            access_token = get_access_token(client_id, client_secret)
            runs = fetch_all_activities(access_token)

    with open(OUTPUT_PATH, 'w') as f:
        json.dump(runs, f, indent=2)
    print(f"Wrote {len(runs)} runs to {OUTPUT_PATH}")

if __name__ == '__main__':
    main()
