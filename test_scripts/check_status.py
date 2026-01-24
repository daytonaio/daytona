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

# Get sandboxes
print("=== Sandboxes ===")
r = requests.get(f'{api_url}/sandboxes', headers=headers)
sandboxes = r.json()
print(json.dumps(sandboxes, indent=2))

# Get snapshots
print("\n=== Snapshots ===")
r = requests.get(f'{api_url}/snapshots', headers=headers)
snapshots = r.json()
print(json.dumps(snapshots, indent=2))
