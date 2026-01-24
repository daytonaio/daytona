
import os
import sys
import time
import requests
from pathlib import Path

# Load .env file
def load_env_file():
    env_file = Path(__file__).parent / '.env'
    if env_file.exists():
        with open(env_file, 'r') as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith('#') and '=' in line:
                    key, value = line.split('=', 1)
                    os.environ[key.strip()] = value.strip()

load_env_file()

try:
    from daytona_sdk import Daytona, DaytonaConfig
    from daytona_sdk.common.daytona import CreateSandboxFromSnapshotParams
except ImportError as e:
    print(f"❌ Failed to import Daytona SDK: {e}")
    sys.exit(1)

def ensure_snapshot(api_url, api_key, org_id, image_name):
    headers = {
        "Authorization": f"Bearer {api_key}",
        "X-Organization-Id": org_id,
        "Content-Type": "application/json"
    }
    
    # Check if exists
    print(f"Checking for snapshot: {image_name}")
    resp = requests.get(f"{api_url}/snapshots", headers=headers)
    if resp.status_code == 200:
        snapshots = resp.json()
        print(f"DEBUG: snapshots type: {type(snapshots)}")
        print(f"DEBUG: snapshots content: {snapshots}")
        if isinstance(snapshots, dict):
            items = snapshots.get('items', snapshots.get('value', []))
        elif isinstance(snapshots, list):
            items = snapshots
        else:
            items = []
            
        for snap in items:
            if snap.get('imageName') == image_name:
                print(f"  Found existing snapshot: {snap.get('id')}")
                return snap.get('id')
                
    # Create if not exists
    print(f"Creating snapshot for: {image_name}")
    payload = {
        "name": f"auto-{int(time.time())}",
        "imageName": image_name
    }
    resp = requests.post(f"{api_url}/snapshots", headers=headers, json=payload)
    if resp.status_code >= 400:
        raise Exception(f"Failed to create snapshot: {resp.text}")
        
    data = resp.json()
    print(f"  Created snapshot: {data.get('id')}")
    return data.get('id')

def main():
    api_key = os.getenv("DAYTONA_API_KEY")
    org_id = os.getenv("DAYTONA_ORG_ID")
    api_url = os.getenv("DAYTONA_API_URL")
    
    # Use a specific, likely uncached image
    target_image = "nginx:1.25-alpine"
    
    print(f"Credential Logging Verification")
    print(f"Target Image: {target_image}")
    
    # 1. Ensure snapshot exists using direct API (as SDK snapshot management might be complex)
    try:
        snapshot_id = ensure_snapshot(api_url, api_key, org_id, target_image)
    except Exception as e:
        print(f"Snapshot setup failed: {e}")
        sys.exit(1)
    
    # 2. Use SDK to create sandbox using that snapshot
    config = DaytonaConfig(
        api_key=api_key,
        api_url=api_url,
        organization_id=org_id,
    )
    
    daytona = Daytona(config)
    
    print(f"\nCreating sandbox from snapshot ID: {snapshot_id}")
    
    try:
        # Correctly using params object
        params = CreateSandboxFromSnapshotParams(
            name=f"cred-test-{int(time.time())}",
            snapshot=snapshot_id
        )
        sandbox = daytona.create(params=params)
        
        print(f"\nSandbox created successfully!")
        print(f"ID: {sandbox.id}")
        
    except Exception as e:
        print(f"\nCreation failed: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

if __name__ == "__main__":
    main()
