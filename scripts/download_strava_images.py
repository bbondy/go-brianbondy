#!/usr/bin/env python3
"""
Script to download images from Strava activities and save them to static/img/runs
Requires Strava API credentials and OAuth authentication
"""

import json
import os
import requests
import time
import webbrowser
from urllib.parse import urlparse, parse_qs
from http.server import HTTPServer, BaseHTTPRequestHandler
import threading
import re

# Strava API configuration
STRAVA_CLIENT_ID = None
STRAVA_CLIENT_SECRET = None
STRAVA_REDIRECT_URI = "http://localhost:8081/callback"
STRAVA_AUTH_URL = "https://www.strava.com/oauth/authorize"
STRAVA_TOKEN_URL = "https://www.strava.com/oauth/token"

class OAuthCallbackHandler(BaseHTTPRequestHandler):
    auth_code = None  # Class variable to hold the auth code

    def do_GET(self):
        if self.path.startswith('/callback'):
            # Parse the authorization code from the callback
            query_components = parse_qs(urlparse(self.path).query)
            
            # Store auth code in the class variable
            OAuthCallbackHandler.auth_code = query_components.get('code', [None])[0]
            
            # Send response to browser
            self.send_response(200)
            self.send_header('Content-type', 'text/html')
            self.end_headers()
            self.wfile.write(b"<html><body><h1>Authorization successful!</h1><p>You can close this window.</p></body></html>")
        else:
            self.send_response(404)
            self.end_headers()

def get_strava_access_token():
    """Get Strava access token using OAuth flow"""
    if not STRAVA_CLIENT_ID or not STRAVA_CLIENT_SECRET:
        print("Error: Please set STRAVA_CLIENT_ID and STRAVA_CLIENT_SECRET in .envrc or the environment")
        return None

    # Reset class variable before starting
    OAuthCallbackHandler.auth_code = None
    
    # Start local server to handle OAuth callback
    server = HTTPServer(('localhost', 8081), OAuthCallbackHandler)
    server_thread = threading.Thread(target=server.serve_forever)
    server_thread.daemon = True
    server_thread.start()
    
    # Open browser for OAuth authorization
    auth_url = f"{STRAVA_AUTH_URL}?client_id={STRAVA_CLIENT_ID}&redirect_uri={STRAVA_REDIRECT_URI}&response_type=code&scope=read_all"
    print(f"Opening browser for Strava authorization: {auth_url}")
    webbrowser.open(auth_url)
    
    # Wait for authorization code
    print("Waiting for authorization...")
    while OAuthCallbackHandler.auth_code is None:
        time.sleep(1)
    
    auth_code = OAuthCallbackHandler.auth_code
    server.shutdown()
    
    if not auth_code:
        print("Failed to get authorization code")
        return None
    
    # Exchange authorization code for access token
    token_data = {
        'client_id': STRAVA_CLIENT_ID,
        'client_secret': STRAVA_CLIENT_SECRET,
        'code': auth_code,
        'grant_type': 'authorization_code'
    }
    
    response = requests.post(STRAVA_TOKEN_URL, data=token_data)
    if response.status_code == 200:
        token_info = response.json()
        return token_info.get('access_token')
    else:
        print(f"Failed to get access token: {response.text}")
        return None

def get_strava_activity_images(activity_id, access_token):
    """Get images from Strava activity using the API"""
    headers = {
        'Authorization': f'Bearer {access_token}'
    }
    
    # Get activity photos. It's good practice to include parameters.
    params = {
        'photo_sources': 'true'
    }
    url = f'https://www.strava.com/api/v3/activities/{activity_id}/photos'
    response = requests.get(url, headers=headers, params=params)

    # Handle rate limiting
    if response.status_code == 429:
        print("Rate limit exceeded. Waiting for 60 seconds before retrying...")
        time.sleep(60)
        response = requests.get(url, headers=headers, params=params)
    
    if response.status_code == 200:
        photos = response.json()
        image_urls = []
        for photo in photos:
            if 'urls' in photo and photo['urls']:
                # The 'urls' dictionary contains different sizes.
                # Prioritize larger images for better quality.
                urls = photo['urls']
                if '2048' in urls:
                    image_urls.append(urls['2048'])
                elif '1024' in urls:
                    image_urls.append(urls['1024'])
                elif '600' in urls:
                    image_urls.append(urls['600'])
                else:
                    # Fallback to the first available URL if specific sizes aren't found
                    image_urls.append(list(urls.values())[0])
        return image_urls
    else:
        # Provide more detailed error output
        error_message = f"Failed to get photos for activity {activity_id}: {response.status_code}"
        try:
            error_details = response.json()
            error_message += f" - {error_details}"
        except ValueError:
            error_message += f" - {response.text}"
        print(error_message)
        return []

def download_image(url, filepath):
    """Download image from URL and save to filepath"""
    try:
        response = requests.get(url, timeout=10)
        response.raise_for_status()
        
        os.makedirs(os.path.dirname(filepath), exist_ok=True)
        with open(filepath, 'wb') as f:
            f.write(response.content)
        print(f"Downloaded: {filepath}")
        return True
    except Exception as e:
        print(f"Failed to download {url}: {e}")
        return False

def extract_activity_id_from_url(url):
    """Extract activity ID from Strava URL"""
    match = re.search(r'/activities/(\d+)', url)
    return match.group(1) if match else None

def process_runs_manifest():
    """Process the memorableRuns.json and download images for each run"""
    manifest_path = "data/memorableRuns.json"
    
    if not os.path.exists(manifest_path):
        print(f"Manifest file not found: {manifest_path}")
        return
    
    # Get Strava access token
    print("Getting Strava access token...")
    access_token = get_strava_access_token()
    if not access_token:
        print("Failed to get access token. Exiting.")
        return
    
    print("Successfully authenticated with Strava!")
    
    with open(manifest_path, 'r') as f:
        runs = json.load(f)
    
    for i, run in enumerate(runs):
        if not run.get('strava_urls'):
            continue
            
        # Create a safe filename for this run
        safe_title = re.sub(r'[^a-zA-Z0-9\s-]', '', run['title'])
        safe_title = re.sub(r'\s+', '-', safe_title).lower()
        
        # For each Strava URL, try to get images
        for j, strava_url in enumerate(run['strava_urls']):
            activity_id = extract_activity_id_from_url(strava_url)
            if not activity_id:
                continue
                
            print(f"Processing activity {activity_id} for run: {run['title']}")
            
            # Create directory for this run
            run_dir = f"static/img/runs/{safe_title}"
            
            # Get images from Strava API
            image_urls = get_strava_activity_images(activity_id, access_token)
            
            if image_urls:
                print(f"Found {len(image_urls)} images for activity {activity_id}")
                for k, image_url in enumerate(image_urls):
                    # Determine file extension
                    parsed_url = urlparse(image_url)
                    ext = os.path.splitext(parsed_url.path)[1] or '.jpg'
                    
                    # Create filename
                    if len(run['strava_urls']) > 1:
                        filename = f"part{j+1}_{k+1}{ext}"
                    else:
                        filename = f"image_{k+1}{ext}"
                    
                    filepath = os.path.join(run_dir, filename)
                    
                    # Download the image
                    if download_image(image_url, filepath):
                        # Update the manifest with the image path
                        if len(run['strava_urls']) > 1:
                            # For multi-part runs, we might want to store multiple image paths
                            if 'image_paths' not in run:
                                run['image_paths'] = []
                            run['image_paths'].append(filepath)
                        else:
                            run['image_path'] = filepath
                    
                    # Be nice to Strava's servers
                    time.sleep(1)
            else:
                print(f"No images found for activity {activity_id}")

            # Be nice to Strava's servers. Pause before the next API call.
            time.sleep(1)
    
    # Save updated manifest
    with open(manifest_path, 'w') as f:
        json.dump(runs, f, indent=2)
    
    print("Finished processing runs manifest")

if __name__ == "__main__":
    STRAVA_CLIENT_ID = os.getenv('STRAVA_CLIENT_ID')
    STRAVA_CLIENT_SECRET = os.getenv('STRAVA_CLIENT_SECRET')

    # Check if credentials are set
    if not STRAVA_CLIENT_ID or not STRAVA_CLIENT_SECRET:
        print("Please set your Strava API credentials:")
        print("1. Go to https://www.strava.com/settings/api")
        print("2. Create a new application")
        print("3. Add these lines to .envrc:")
        print("export STRAVA_CLIENT_ID='your_client_id'")
        print("export STRAVA_CLIENT_SECRET='your_client_secret'")
        print("\nOr run the setup script:")
        print("python3 scripts/setup_strava.py")
        print("No credentials found. Exiting.")
        exit(1)
    
    process_runs_manifest()
