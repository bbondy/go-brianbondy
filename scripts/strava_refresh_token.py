#!/usr/bin/env python3
"""
Mint a long-lived Strava refresh token for unattended use.

Run this locally, once. It opens a browser for the Strava consent screen, then
prints the refresh token plus the `gh secret set` commands needed to wire up
the daily stats workflow. Re-run it whenever Strava rotates the token and the
scheduled workflow starts failing to authorize.

Usage:
    python3 scripts/strava_refresh_token.py
"""

import json
import os
import threading
import time
import webbrowser
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import parse_qs, urlencode, urlparse

import requests

AUTH_URL = 'https://www.strava.com/oauth/authorize'
TOKEN_URL = 'https://www.strava.com/oauth/token'
REDIRECT_URI = 'http://localhost:8080/'
SCOPE = 'read,activity:read_all'
TOKEN_PATH = os.path.expanduser('~/.strava_token.json')
REPO = 'bbondy/go-brianbondy'


class OAuthHandler(BaseHTTPRequestHandler):
    code = None

    def do_GET(self):
        params = parse_qs(urlparse(self.path).query)
        if 'code' in params:
            OAuthHandler.code = params['code'][0]
            body = b'<h1>Authorization successful. You can close this window.</h1>'
            status = 200
        else:
            body = b'<h1>Authorization failed.</h1>'
            status = 400
        self.send_response(status)
        self.send_header('Content-type', 'text/html')
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *args):
        pass  # Keep the console output limited to our own messages.


def authorize(client_id: str, client_secret: str) -> dict:
    OAuthHandler.code = None
    params = {
        'client_id': client_id,
        'redirect_uri': REDIRECT_URI,
        'response_type': 'code',
        'scope': SCOPE,
        # Force the consent screen so the granted scopes are always refreshed.
        'approval_prompt': 'force',
    }
    url = f'{AUTH_URL}?{urlencode(params)}'
    print('Opening browser for Strava authorization...')
    print(f'If it does not open, visit:\n  {url}\n')

    server = HTTPServer(('localhost', 8080), OAuthHandler)
    threading.Thread(target=server.handle_request, daemon=True).start()
    webbrowser.open(url)

    while OAuthHandler.code is None:
        time.sleep(0.2)

    resp = requests.post(TOKEN_URL, data={
        'client_id': client_id,
        'client_secret': client_secret,
        'code': OAuthHandler.code,
        'grant_type': 'authorization_code',
    }, timeout=15)
    if resp.status_code != 200:
        raise SystemExit(f'Failed to exchange code for token: '
                         f'{resp.status_code} {resp.text}')
    return resp.json()


def main() -> None:
    client_id = os.environ.get('STRAVA_CLIENT_ID') or input('Strava client ID: ').strip()
    client_secret = (os.environ.get('STRAVA_CLIENT_SECRET')
                     or input('Strava client secret: ').strip())
    if not client_id or not client_secret:
        raise SystemExit('Both a client ID and a client secret are required. '
                         'Get them from https://www.strava.com/settings/api')

    token_data = authorize(client_id, client_secret)
    refresh_token = token_data.get('refresh_token')
    if not refresh_token:
        raise SystemExit(f'No refresh token in the Strava response: {token_data}')

    with open(TOKEN_PATH, 'w') as f:
        json.dump(token_data, f)
    print(f'Token saved to {TOKEN_PATH}')

    # Strava's consent screen has a checkbox per permission and it is easy to
    # leave the activity one unticked. Such a token refreshes fine but 401s on
    # every activity call, so catch it here rather than in a CI run tomorrow.
    granted = token_data.get('scope', '')
    print(f'Granted scopes: {granted or "(not reported)"}')
    if 'activity:read_all' not in granted:
        raise SystemExit(
            "\nThis token cannot read your activities, so the stats workflow "
            "will fail with it.\nRe-run this script and tick the box for "
            "viewing your activity data on the Strava consent screen."
        )
    print()

    print('Refresh token:')
    print(f'  {refresh_token}\n')
    print('Store the credentials as repository secrets:')
    print(f'  gh secret set STRAVA_CLIENT_ID --repo {REPO} --body {client_id!r}')
    print(f'  gh secret set STRAVA_CLIENT_SECRET --repo {REPO} --body <client secret>')
    print(f'  gh secret set STRAVA_REFRESH_TOKEN --repo {REPO} --body {refresh_token!r}')


if __name__ == '__main__':
    main()
