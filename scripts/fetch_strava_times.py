#!/usr/bin/env python3
"""
Script to fetch time, distance, and elevation information from Strava activities and update the run manifest.
This script will:
1. Read the current runManifest.json
2. For each activity with Strava URLs, fetch the time, distance, and elevation information
3. Add 'time', 'distance', and 'elevation' fields to each activity
4. Update the manifest file
"""

import json
import re
import requests
import time
from urllib.parse import urlparse
from typing import Dict, List, Optional, Tuple
from bs4 import BeautifulSoup
from bs4.element import Tag

def extract_activity_id_from_url(url: str) -> Optional[str]:
    """Extract activity ID from Strava URL."""
    if 'strava.com/activities/' in url:
        match = re.search(r'/activities/(\d+)', url)
        if match:
            return match.group(1)
    return None

def fetch_strava_activity_metrics_from_web(url: str) -> Optional[Dict[str, str]]:
    """
    Fetch time, distance, and elevation from the Strava activity web page.
    Returns a dict with 'time', 'distance', and 'elevation' keys, or None if not found.
    """
    headers = {
        'User-Agent': 'Mozilla/5.0 (compatible; StravaTimeFetcher/1.0)'
    }
    try:
        resp = requests.get(url, headers=headers, timeout=10)
        if resp.status_code != 200:
            print(f"Failed to fetch {url}: {resp.status_code}")
            return None
        soup = BeautifulSoup(resp.text, 'html.parser')
        
        metrics = {}
        
        # Look for time metrics
        for label in ['Elapsed Time', 'Time']:
            el = soup.find(string=label)
            if el and hasattr(el, 'parent'):
                parent = el.parent
                value = None
                # Try next sibling if it's a Tag
                if parent and parent.next_sibling and hasattr(parent.next_sibling, 'get_text'):
                    value = parent.next_sibling.get_text(strip=True)
                if not value or value == label:
                    # Try parent.parent (table row)
                    if parent and parent.parent:
                        tds = parent.parent.find_all('td')
                        if len(tds) == 2:
                            value = tds[1].get_text(strip=True)
                if value and value != label:
                    # Convert value to 'Xh Ym' format
                    h_m = re.match(r'(?:(\d+)h)?\s*(\d+)?m', value)
                    if h_m:
                        hours = h_m.group(1)
                        minutes = h_m.group(2)
                        if hours and minutes:
                            metrics['time'] = f"{int(hours)}h {int(minutes)}m"
                        elif hours:
                            metrics['time'] = f"{int(hours)}h"
                        elif minutes:
                            metrics['time'] = f"{int(minutes)}m"
                    else:
                        metrics['time'] = value
                    break
        
        # Look for distance metrics
        for label in ['Distance', 'Miles', 'Kilometers']:
            el = soup.find(string=label)
            if el and hasattr(el, 'parent'):
                parent = el.parent
                value = None
                if parent and parent.next_sibling and hasattr(parent.next_sibling, 'get_text'):
                    value = parent.next_sibling.get_text(strip=True)
                if not value or value == label:
                    if parent and parent.parent:
                        tds = parent.parent.find_all('td')
                        if len(tds) == 2:
                            value = tds[1].get_text(strip=True)
                if value and value != label:
                    # Clean up distance value
                    distance_match = re.search(r'([\d.]+)\s*(mi|km|miles?|kilometers?)', value, re.IGNORECASE)
                    if distance_match:
                        distance = distance_match.group(1)
                        unit = distance_match.group(2).lower()
                        if unit in ['km', 'kilometers', 'kilometer']:
                            metrics['distance'] = f"{distance}km"
                        else:
                            metrics['distance'] = f"{distance}mi"
                    else:
                        metrics['distance'] = value
                    break
        
        # Look for elevation metrics
        for label in ['Elevation Gain', 'Elevation', 'Total Elevation']:
            el = soup.find(string=label)
            if el and hasattr(el, 'parent'):
                parent = el.parent
                value = None
                if parent and parent.next_sibling and hasattr(parent.next_sibling, 'get_text'):
                    value = parent.next_sibling.get_text(strip=True)
                if not value or value == label:
                    if parent and parent.parent:
                        tds = parent.parent.find_all('td')
                        if len(tds) == 2:
                            value = tds[1].get_text(strip=True)
                if value and value != label:
                    # Clean up elevation value
                    elevation_match = re.search(r'([\d,]+)\s*(ft|m|feet|meters?)', value, re.IGNORECASE)
                    if elevation_match:
                        elevation = elevation_match.group(1).replace(',', '')
                        unit = elevation_match.group(2).lower()
                        if unit in ['m', 'meters', 'meter']:
                            metrics['elevation'] = f"{elevation}m"
                        else:
                            metrics['elevation'] = f"{elevation}ft"
                    else:
                        metrics['elevation'] = value
                    break
        
        # Try to find metrics in meta tags (for mobile layout)
        meta_content = ''
        meta = soup.find('meta', {'name': 'description'})
        if isinstance(meta, Tag) and 'content' in meta.attrs:
            meta_content = meta.attrs['content']
        if isinstance(meta_content, str):
            # Extract time from meta
            if 'Elapsed Time' in meta_content and 'time' not in metrics:
                m = re.search(r'Elapsed Time: ([0-9h m]+)', meta_content)
                if m:
                    metrics['time'] = m.group(1).strip()
            
            # Extract distance from meta
            if 'Distance' in meta_content and 'distance' not in metrics:
                m = re.search(r'Distance: ([0-9.]+)\s*(mi|km)', meta_content, re.IGNORECASE)
                if m:
                    distance = m.group(1)
                    unit = m.group(2).lower()
                    if unit == 'km':
                        metrics['distance'] = f"{distance}km"
                    else:
                        metrics['distance'] = f"{distance}mi"
            
            # Extract elevation from meta
            if 'Elevation' in meta_content and 'elevation' not in metrics:
                m = re.search(r'Elevation: ([0-9,]+)\s*(ft|m)', meta_content, re.IGNORECASE)
                if m:
                    elevation = m.group(1).replace(',', '')
                    unit = m.group(2).lower()
                    if unit == 'm':
                        metrics['elevation'] = f"{elevation}m"
                    else:
                        metrics['elevation'] = f"{elevation}ft"
        
        # Fallback: search for metrics in the whole page
        text = soup.get_text()
        if isinstance(text, str):
            # Time fallback
            if 'time' not in metrics:
                m = re.search(r'Elapsed Time:?\s*([0-9]+h)?\s*([0-9]+m)', text)
                if m:
                    hours = m.group(1).strip() if m.group(1) else ''
                    minutes = m.group(2).strip() if m.group(2) else ''
                    metrics['time'] = f"{hours} {minutes}".strip()
            
            # Distance fallback
            if 'distance' not in metrics:
                m = re.search(r'Distance:?\s*([0-9.]+)\s*(mi|km)', text, re.IGNORECASE)
                if m:
                    distance = m.group(1)
                    unit = m.group(2).lower()
                    if unit == 'km':
                        metrics['distance'] = f"{distance}km"
                    else:
                        metrics['distance'] = f"{distance}mi"
            
            # Elevation fallback
            if 'elevation' not in metrics:
                m = re.search(r'Elevation:?\s*([0-9,]+)\s*(ft|m)', text, re.IGNORECASE)
                if m:
                    elevation = m.group(1).replace(',', '')
                    unit = m.group(2).lower()
                    if unit == 'm':
                        metrics['elevation'] = f"{elevation}m"
                    else:
                        metrics['elevation'] = f"{elevation}ft"
        
        return metrics if metrics else None
        
    except Exception as e:
        print(f"Error fetching/parsing {url}: {e}")
    return None

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
    """
    Fetch metrics from multiple Strava URLs and return the totals.
    Returns a dict with 'time', 'distance', and 'elevation' keys, or None if no metrics found.
    """
    total_minutes = 0
    total_miles = 0.0
    total_feet = 0
    found_metrics = 0
    
    for url in urls:
        if 'strava.com/activities/' in url:
            metrics = fetch_strava_activity_metrics_from_web(url)
            if metrics:
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
    
    if found_metrics > 0:
        total_metrics = {}
        if total_minutes > 0:
            total_metrics['time'] = minutes_to_time_string(total_minutes)
        if total_miles > 0:
            total_metrics['distance'] = miles_to_distance_string(total_miles)
        if total_feet > 0:
            total_metrics['elevation'] = feet_to_elevation_string(total_feet)
        
        print(f"  Total: {total_metrics.get('time', 'N/A')}, {total_metrics.get('distance', 'N/A')}, {total_metrics.get('elevation', 'N/A')} (from {found_metrics} activities)")
        return total_metrics
    
    return None

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
    
    # Read current manifest
    with open('data/runManifest.json', 'r') as f:
        runs = json.load(f)
    
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
    with open('data/runManifest.json', 'w') as f:
        json.dump(updated_runs, f, indent=2)
    
    print(f"Updated {len(updated_runs)} activities in runManifest.json")

if __name__ == "__main__":
    update_run_manifest() 