import os
from dotenv import load_dotenv
from daytona_sdk import Daytona, DaytonaConfig, CreateSandboxFromImageParams

load_dotenv('scripts/.env')

api_url = os.getenv('DAYTONA_API_URL')
api_key = os.getenv('DAYTONA_API_KEY')
org_id = os.getenv('DAYTONA_ORG_ID')

print("Testing sandbox creation with explicit image (Docker Hub auth fix verification)")
print("=" * 80)

# Configure Daytona client
config = DaytonaConfig(
    api_key=api_key,
    api_url=api_url,
    organization_id=org_id
)

daytona = Daytona(config)

# Test with explicit images that will require manifest lookup from Docker Hub
test_images = [
    "nginx:1.25-alpine",
    "python:3.11-slim",
    "redis:7.0-alpine"
]

for test_image in test_images:
    print(f"\n{'='*80}")
    print(f"Creating sandbox from image: {test_image}")
    print(f"{'='*80}")
    
    try:
        params = CreateSandboxFromImageParams(
            image=test_image
        )
        
        sandbox = daytona.create(params=params)
        print(f"✓ Sandbox created successfully!")
        print(f"  ID: {sandbox.id}")
        print(f"  State: {sandbox.state}")
        
        # Clean up
        print(f"  Cleaning up sandbox {sandbox.id}...")
        daytona.delete(sandbox.id)
        print(f"  ✓ Cleanup complete")
        
    except Exception as e:
        print(f"✗ Error creating sandbox from {test_image}: {e}")
        import traceback
        traceback.print_exc()

print(f"\n{'='*80}")
print("Test complete!")
