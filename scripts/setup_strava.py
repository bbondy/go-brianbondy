#!/usr/bin/env python3
"""
Setup script for Strava API credentials
"""

import os

ENVRC_PATH = '.envrc'


def upsert_envrc_var(lines, key, value):
    """Insert or replace a key in .envrc while preserving unrelated lines."""
    new_line = f"export {key}={value}\n"
    for i, line in enumerate(lines):
        stripped = line.strip()
        if stripped.startswith(f"{key}=") or stripped.startswith(f"export {key}="):
            lines[i] = new_line
            return lines

    if lines and lines[-1].strip():
        lines.append('\n')
    lines.append(new_line)
    return lines

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

    lines = []
    if os.path.exists(ENVRC_PATH):
        with open(ENVRC_PATH, 'r') as f:
            lines = f.readlines()

    lines = upsert_envrc_var(lines, "STRAVA_CLIENT_ID", client_id)
    lines = upsert_envrc_var(lines, "STRAVA_CLIENT_SECRET", client_secret)

    with open(ENVRC_PATH, 'w') as f:
        f.writelines(lines)

    print()
    print("Credentials saved to .envrc")
    print("You can now run: python3 scripts/download_strava_images.py")
    print()
    print("Note: .envrc contains sensitive information and should not be committed")

    return True

if __name__ == "__main__":
    setup_strava_credentials()
