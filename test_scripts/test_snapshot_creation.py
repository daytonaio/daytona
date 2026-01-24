import os
from dotenv import load_dotenv
import requests
import json

load_dotenv('scripts/.env')

api_url = os.getenv('DAYTONA_API_URL')
api_key = os.getenv('DAYTONA_API_KEY')
org_id = os.getenv('DAYTONA_ORG_ID')

headers = {
    'Authorization': f'Bearer {api_key}',
    'X-Organization-Id': org_id
}

# Test snapshot creation with a fresh image
test_image = "python:3.11-slim"

print(f"Testing snapshot creation with: {test_image}")
print("=" * 60)

# Create snapshot
snapshot_data = {
    "imageName": test_image
}

r = requests.post(f'{api_url}/snapshots', headers=headers, json=snapshot_data)
print(f"Status Code: {r.status_code}")
print(f"Response: {json.dumps(r.json(), indent=2)}")
