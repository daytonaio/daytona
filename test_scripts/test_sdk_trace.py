
import os
import sys
import json
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
    from daytona_api_client import ApiClient
except ImportError as e:
    print(f"❌ Failed to import Daytona SDK: {e}")
    sys.exit(1)

# Monkeypatch verify
original_call_api = ApiClient.call_api

def traced_call_api(self, resource_path, method, *args, **kwargs):
    print(f"\n[SDK TRACE] {method} {resource_path}")
    print(f"  Args: {args}")
    print(f"  Kwargs: {kwargs}")
        
    try:
        response = original_call_api(self, resource_path, method, *args, **kwargs)
        return response
    except Exception as e:
        print(f"  ❌ Exception: {e}")
        # Print status code if available
        if hasattr(e, 'status'):
            print(f"  Status: {e.status}")
        if hasattr(e, 'body'):
            print(f"  Body: {e.body}")
        raise e

ApiClient.call_api = traced_call_api

def main():
    api_key = os.getenv("DAYTONA_API_KEY")
    org_id = os.getenv("DAYTONA_ORG_ID")
    api_url = os.getenv("DAYTONA_API_URL")
    
    print(f"Config: URL={api_url}, Org={org_id}")
    
    config = DaytonaConfig(
        api_key=api_key,
        target=api_url,
        organization_id=org_id,
    )
    
    daytona = Daytona(config)
    
    print("calling create...")
    try:
        sandbox = daytona.create()
        print("Success!")
    except Exception as e:
        print(f"Caught error: {e}")

if __name__ == "__main__":
    main()
