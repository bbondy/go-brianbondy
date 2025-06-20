#!/usr/bin/env python3
"""
Setup script for Strava API credentials
"""

import os
import json

def setup_strava_credentials():
    """Interactive setup for Strava API credentials"""
    print("Strava API Setup")
    print("================")
    print()
    print("To use the Strava image downloader, you need to create a Strava API application:")
    print()
    print("1. Go to https://www.strava.com/settings/api")
    print("2. Click 'Create Your Own Application'")
    print("3. Fill in the form:")
    print("   - Application Name: Brian's Running Site")
    print("   - Category: Website")
    print("   - Website: https://brianbondy.com")
    print("   - Authorization Callback Domain: localhost")
    print("4. Click 'Create'")
    print("5. Note down your Client ID and Client Secret")
    print()
    
    client_id = input("Enter your Strava Client ID: ").strip()
    client_secret = input("Enter your Strava Client Secret: ").strip()
    
    if not client_id or not client_secret:
        print("Error: Both Client ID and Client Secret are required")
        return False
    
    # Create .env file
    env_content = f"""# Strava API Credentials
STRAVA_CLIENT_ID={client_id}
STRAVA_CLIENT_SECRET={client_secret}
"""
    
    with open('.env', 'w') as f:
        f.write(env_content)
    
    print()
    print("Credentials saved to .env file")
    print("You can now run: python3 scripts/download_strava_images.py")
    print()
    print("Note: The .env file contains sensitive information. Make sure it's in your .gitignore")
    
    return True

if __name__ == "__main__":
    setup_strava_credentials() 