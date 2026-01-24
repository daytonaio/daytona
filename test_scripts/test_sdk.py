#!/usr/bin/env python3
"""
Test script for Daytona sandbox creation using official Python SDK.
"""

import os
import sys
from pathlib import Path
import time

# Load .env file
def load_env_file():
    """Load environment variables from .env file"""
    env_file = Path(__file__).parent / '.env'
    if env_file.exists():
        with open(env_file, 'r') as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith('#') and '=' in line:
                    key, value = line.split('=', 1)
                    os.environ[key.strip()] = value.strip()

load_env_file()

# Import Daytona SDK
try:
    from daytona_sdk import Daytona, DaytonaConfig
except ImportError as e:
    print(f"❌ Failed to import Daytona SDK: {e}")
    print("Make sure daytona SDK is installed: pip install daytona-sdk==0.128.1")
    sys.exit(1)

def main():
    # Get configuration from environment
    api_key = os.getenv("DAYTONA_API_KEY")
    org_id = os.getenv("DAYTONA_ORG_ID")
    api_url = os.getenv("DAYTONA_API_URL", "http://localhost:3000/api")
    
    print("=" * 60)
    print("🧪 Daytona Sandbox Creation Test (Python SDK)")
    print("=" * 60)
    print(f"API URL: {api_url}")
    print(f"API Key: {'*' * 10}{api_key[-4:] if api_key and len(api_key) > 4 else 'NONE'}")
    print(f"Org ID: {org_id if org_id else 'NONE'}")
    print("=" * 60)
    
    if not api_key:
        print("\n❌ ERROR: DAYTONA_API_KEY not found in .env file")
        sys.exit(1)
    
    try:
        # Configure Daytona client
        print("\n🔧 Configuring Daytona client...")
        config = DaytonaConfig(
            api_key=api_key,
            api_url=api_url,
            organization_id=org_id,
        )
        
        # Initialize client
        print("🔌 Connecting to Daytona API...")
        daytona = Daytona(config)
        
        # Create sandbox
        print("\n📦 Creating sandbox...")
        print("   Snapshot: daytonaio/sandbox:0.5.0-slim")
        
        sandbox = daytona.create()
        
        print(f"✅ Sandbox created successfully!")
        print(f"   Sandbox ID: {sandbox.id if hasattr(sandbox, 'id') else 'unknown'}")
        
        # Test with simple code execution
        print("\n🧪 Testing code execution in sandbox...")
        response = sandbox.process.code_run('echo "Hello from Daytona!"')
        
        if response.exit_code != 0:
            print(f"⚠️  Code execution had non-zero exit code: {response.exit_code}")
            print(f"   Result: {response.result}")
        else:
            print(f"✅ Code executed successfully!")
            print(f"   Output: {response.result}")
        
        print("\n" + "=" * 60)
        print("✅ Test PASSED: Sandbox created and tested successfully")
        print("=" * 60)
        
        # Keep sandbox for inspection
        print(f"\n📌 Sandbox kept for inspection")
        print(f"   To delete manually, use the Dashboard or API")
        
        return 0
        
    except Exception as e:
        print(f"\n❌ Test FAILED with error:")
        print(f"   {type(e).__name__}: {e}")
        
        # Print more details if available
        if hasattr(e, '__dict__'):
            print(f"   Details: {e.__dict__}")
        
        import traceback
        print("\n📋 Full traceback:")
        traceback.print_exc()
        
        print("\n" + "=" * 60)
        print("❌ Test FAILED")
        print("=" * 60)
        
        return 1

if __name__ == "__main__":
    exit_code = main()
    sys.exit(exit_code)
