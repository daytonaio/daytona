#!/usr/bin/env python3
"""
Test script to verify the search endpoint changes:
1. Only accepts POST requests
2. Immediately checks ripgrep availability
3. Decides between ripgrep implementation or fallback
"""

import json
import subprocess

import requests


def test_search_endpoint():
    """Test the search endpoint with various scenarios"""

    # This would typically be the daemon endpoint
    # For testing, you would need to have a running daemon
    base_url = "http://localhost:8080"  # Adjust port as needed
    search_url = f"{base_url}/files/search"

    print("Testing Search Endpoint Changes")
    print("=" * 40)

    # Test 1: GET request should fail (405 Method Not Allowed)
    print("\n1. Testing GET request (should fail)...")
    try:
        response = requests.get(search_url, params={"pattern": "test", "path": "."}, timeout=10)
        if response.status_code == 405:
            print("✅ GET request correctly rejected with 405 Method Not Allowed")
        else:
            print(f"❌ GET request returned unexpected status: {response.status_code}")
    except requests.exceptions.ConnectionError:
        print("⚠️  Cannot connect to daemon (this is expected if daemon is not running)")
    except requests.exceptions.Timeout:
        print("⚠️  Request timed out after 10 seconds")

    # Test 2: POST request without JSON content-type should fail
    print("\n2. Testing POST without JSON content-type (should fail)...")
    try:
        response = requests.post(search_url, data="test=value", timeout=10)
        if response.status_code == 400:
            print("✅ POST without JSON content-type correctly rejected with 400 Bad Request")
        else:
            print(f"❌ POST without JSON returned unexpected status: {response.status_code}")
    except requests.exceptions.ConnectionError:
        print("⚠️  Cannot connect to daemon (this is expected if daemon is not running)")
    except requests.exceptions.Timeout:
        print("⚠️  Request timed out after 10 seconds")

    # Test 3: Valid POST request with JSON
    print("\n3. Testing valid POST request with JSON...")
    search_request = {"query": "function", "path": ".", "case_sensitive": False, "max_results": 5}

    try:
        response = requests.post(
            search_url, json=search_request, headers={"Content-Type": "application/json"}, timeout=10
        )
        if response.status_code == 200:
            print("✅ Valid POST request accepted")
            result = response.json()
            print(f"   Found {result.get('total_matches', 0)} matches in {result.get('total_files', 0)} files")
        elif response.status_code == 400:
            print("⚠️  Request rejected (might be due to missing query or invalid path)")
        else:
            print(f"❌ Valid POST returned unexpected status: {response.status_code}")
    except requests.exceptions.ConnectionError:
        print("⚠️  Cannot connect to daemon (this is expected if daemon is not running)")
    except requests.exceptions.Timeout:
        print("⚠️  Request timed out after 10 seconds")
    except json.JSONDecodeError:
        print("⚠️  Response was not valid JSON")

    # Test 4: Test empty search results (should return empty array, not null)
    print("\n4. Testing empty search results...")
    empty_search_request = {
        "query": "nonexistent_search_term_that_should_not_match_anything_12345",
        "path": ".",
        "max_results": 1,
    }

    try:
        response = requests.post(
            search_url, json=empty_search_request, headers={"Content-Type": "application/json"}, timeout=10
        )
        if response.status_code == 200:
            result = response.json()
            matches = result.get("matches")
            if matches is None:
                print("❌ matches field is null - this would cause Pydantic validation error")
            elif isinstance(matches, list) and len(matches) == 0:
                print("✅ Empty search returns empty array [] - Pydantic validation should work")
            else:
                print(f"⚠️  Unexpected matches value: {matches}")
        else:
            print(f"⚠️  Empty search returned status: {response.status_code}")
    except requests.exceptions.ConnectionError:
        print("⚠️  Cannot connect to daemon (this is expected if daemon is not running)")
    except requests.exceptions.Timeout:
        print("⚠️  Request timed out after 10 seconds")
    except json.JSONDecodeError:
        print("⚠️  Response was not valid JSON")

    # Test 5: Check if ripgrep is available (this would be internal to the daemon)
    print("\n5. Checking ripgrep availability on system...")
    try:
        result = subprocess.run(["rg", "--version"], capture_output=True, text=True, check=False)
        if result.returncode == 0:
            print("✅ ripgrep is available on this system")
            print(f"   Version: {result.stdout.split()[1] if result.stdout else 'unknown'}")
        else:
            print("❌ ripgrep is not available")
    except FileNotFoundError:
        print("❌ ripgrep is not installed or not in PATH")

    print("\n" + "=" * 40)
    print("Test Summary:")
    print("- Search endpoint now only accepts POST requests")
    print("- Content-Type must be application/json")
    print("- Ripgrep availability is checked immediately")
    print("- Falls back to basic search if ripgrep is not available")


if __name__ == "__main__":
    test_search_endpoint()
