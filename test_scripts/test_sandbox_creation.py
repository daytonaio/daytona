#!/usr/bin/env python3
"""
Test script for Daytona sandbox creation.
This script creates a sandbox and monitors its creation process.

Usage:
    python test_sandbox_creation.py [--api-url URL] [--token TOKEN]
"""

import requests
import time
import argparse
import sys
import os
from pathlib import Path
from typing import Optional, Dict, Any

# Load .env file if exists
def load_env_file():
    """Load environment variables from .env file in the same directory"""
    env_file = Path(__file__).parent / '.env'
    if env_file.exists():
        with open(env_file, 'r') as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith('#') and '=' in line:
                    key, value = line.split('=', 1)
                    os.environ[key.strip()] = value.strip()

# Load .env before using environment variables
load_env_file()

# Default configuration (can be overridden by .env or command line)
DEFAULT_API_URL = os.getenv("DAYTONA_API_URL", "http://localhost:3000")
DEFAULT_API_TOKEN = os.getenv("DAYTONA_API_KEY", "")
DEFAULT_ORG_ID = os.getenv("DAYTONA_ORG_ID", "")
DEFAULT_SNAPSHOT = "daytonaio/sandbox:0.5.0-slim"


class DaytonaClient:
    def __init__(self, api_url: str, api_token: Optional[str] = None, org_id: Optional[str] = None):
        self.api_url = api_url.rstrip('/')
        self.access_token = api_token
        self.organization_id = org_id
        
    def create_sandbox(self, name: str, snapshot: str) -> Optional[Dict[str, Any]]:
        """Create a new sandbox"""
        print(f"\n📦 Creating sandbox '{name}' with snapshot '{snapshot}'...")
        
        headers = {
            "Content-Type": "application/json"
        }
        
        if self.access_token:
            headers["Authorization"] = f"Bearer {self.access_token}"
        
        if self.organization_id:
            headers["X-Organization-Id"] = self.organization_id
        
        payload = {
            "name": name,
            "snapshot": snapshot,
            "user": "daytona"
        }
        
        print(f"📤 Request:")
        print(f"   URL: {self.api_url}/sandbox")
        print(f"   Headers: {headers}")
        print(f"   Payload: {payload}")
        
        try:
            response = requests.post(
                f"{self.api_url}/sandbox",
                json=payload,
                headers=headers,
                timeout=300,  # 5 minutes for sandbox creation
                allow_redirects=False  # Don't follow redirects, catch them
            )
            
            print(f"📥 Response:")
            print(f"   Status: {response.status_code}")
            print(f"   Headers: {dict(response.headers)}")
            print(f"   Content-Type: {response.headers.get('Content-Type', 'unknown')}")
            
            if response.status_code in [200, 201]:
                sandbox = response.json()
                print(f"✅ Sandbox created: {sandbox.get('id', 'unknown')}")
                print(f"   State: {sandbox.get('state', 'unknown')}")
                return sandbox
            elif response.status_code in [301, 302, 303, 307, 308]:
                print(f"⚠️  Redirect detected:")
                print(f"   Location: {response.headers.get('Location', 'unknown')}")
                return None
            else:
                print(f"❌ Failed to create sandbox:")
                print(f"   Status: {response.status_code}")
                print(f"   Response: {response.text[:500]}")  # First 500 chars
                return None
                
        except requests.exceptions.JSONDecodeError as e:
            print(f"❌ JSON decode error: {e}")
            print(f"   Response status: {response.status_code}")
            print(f"   Response text: {response.text[:500]}")
            return None
        except requests.exceptions.RequestException as e:
            print(f"❌ Request failed: {e}")
            return None
    
    def get_sandbox(self, sandbox_id: str) -> Optional[Dict[str, Any]]:
        """Get sandbox details"""
        headers = {}
        if self.access_token:
            headers["Authorization"] = f"Bearer {self.access_token}"
        if self.organization_id:
            headers["X-Organization-Id"] = self.organization_id
        
        try:
            response = requests.get(
                f"{self.api_url}/sandbox/{sandbox_id}",
                headers=headers,
                timeout=10
            )
            
            if response.status_code == 200:
                return response.json()
            else:
                print(f"⚠️  Failed to get sandbox: {response.status_code}")
                return None
                
        except requests.exceptions.RequestException as e:
            print(f"❌ Request failed: {e}")
            return None
    
    def wait_for_sandbox_started(self, sandbox_id: str, timeout: int = 120) -> bool:
        """Wait for sandbox to be in STARTED state"""
        print(f"\n⏳ Waiting for sandbox to start (timeout: {timeout}s)...")
        
        start_time = time.time()
        last_state = None
        
        while time.time() - start_time < timeout:
            sandbox = self.get_sandbox(sandbox_id)
            
            if sandbox:
                state = sandbox.get('state')
                
                if state != last_state:
                    print(f"   State: {state}")
                    last_state = state
                
                if state == 'STARTED':
                    print(f"✅ Sandbox started successfully!")
                    return True
                elif state == 'ERROR':
                    print(f"❌ Sandbox entered ERROR state")
                    error_reason = sandbox.get('errorReason', 'Unknown error')
                    print(f"   Error: {error_reason}")
                    return False
            
            time.sleep(2)
        
        print(f"⏱️  Timeout waiting for sandbox to start")
        return False
    
    def delete_sandbox(self, sandbox_id: str) -> bool:
        """Delete a sandbox"""
        print(f"\n🗑️  Deleting sandbox {sandbox_id}...")
        
        headers = {}
        if self.access_token:
            headers["Authorization"] = f"Bearer {self.access_token}"
        if self.organization_id:
            headers["X-Organization-Id"] = self.organization_id
        
        try:
            response = requests.delete(
                f"{self.api_url}/sandbox/{sandbox_id}",
                headers=headers,
                timeout=60
            )
            
            if response.status_code in [200, 204]:
                print(f"✅ Sandbox deleted successfully")
                return True
            else:
                print(f"⚠️  Failed to delete sandbox: {response.status_code}")
                return False
                
        except requests.exceptions.RequestException as e:
            print(f"❌ Delete request failed: {e}")
            return False


def main():
    parser = argparse.ArgumentParser(description="Test Daytona sandbox creation")
    parser.add_argument("--api-url", default=DEFAULT_API_URL, help=f"Daytona API URL (default: {DEFAULT_API_URL})")
    parser.add_argument("--token", default=DEFAULT_API_TOKEN, help="API token (from .env or DAYTONA_API_KEY)")
    parser.add_argument("--org-id", default=DEFAULT_ORG_ID, help="Organization ID (from .env or DAYTONA_ORG_ID)")
    parser.add_argument("--snapshot", default=DEFAULT_SNAPSHOT, help="Snapshot image")
    parser.add_argument("--name", help="Sandbox name (auto-generated if not provided)")
    parser.add_argument("--keep", action="store_true", help="Don't delete sandbox after test")
    
    args = parser.parse_args()
    
    # Generate sandbox name if not provided
    sandbox_name = args.name or f"test-sandbox-{int(time.time())}"
    
    print("=" * 60)
    print("🧪 Daytona Sandbox Creation Test")
    print("=" * 60)
    print(f"API URL: {args.api_url}")
    print(f"Snapshot: {args.snapshot}")
    print(f"Sandbox Name: {sandbox_name}")
    if args.token:
        print(f"API Token: {'*' * 10}{args.token[-4:] if len(args.token) > 4 else '****'}")
    if args.org_id:
        print(f"Organization ID: {args.org_id}")
    print("=" * 60)
    
    # Create client
    client = DaytonaClient(args.api_url, args.token, args.org_id)
    
    if not args.token:
        print("\n⚠️  WARNING: No API token provided!")
        print("   Set DAYTONA_API_KEY in scripts/.env file or use --token argument")
        print("   Attempting to proceed without authentication (will likely fail)\n")
    
    # Create sandbox
    sandbox = client.create_sandbox(sandbox_name, args.snapshot)
    
    if not sandbox:
        print("\n❌ Test FAILED: Could not create sandbox")
        sys.exit(1)
    
    sandbox_id = sandbox.get('id')
    
    # Wait for sandbox to start
    if client.wait_for_sandbox_started(sandbox_id):
        print(f"\n✅ Test PASSED: Sandbox created and started successfully")
        success = True
    else:
        print(f"\n❌ Test FAILED: Sandbox did not start")
        success = False
    
    # Delete sandbox unless --keep flag is set
    if not args.keep:
        client.delete_sandbox(sandbox_id)
    else:
        print(f"\n📌 Sandbox kept: {sandbox_id}")
        print(f"   Delete manually or run: curl -X DELETE {args.api_url}/sandbox/{sandbox_id}")
    
    print("\n" + "=" * 60)
    print("🏁 Test Complete")
    print("=" * 60)
    
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
