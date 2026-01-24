import requests
import os
import json
import time

# Configuration from .env
api_key = "dtn_3d0209c8e694c3d68ebb3338d0b24c29181f07755d3e7782b7960cd1ce5f328e"
org_id = "663452f8-6fcf-4d6b-bce5-194ddc0773d7"
api_url = "http://localhost:3000/api"

headers = {
    "Authorization": f"Bearer {api_key}",
    "X-Organization-Id": org_id,
    "Content-Type": "application/json",
    "Accept": "application/json"
}

def create_sandbox():
    payload = {
        "name": f"sdk-repro-{int(time.time())}",
        "snapshot": "daytonaio/sandbox:0.5.0-slim"
    }
    
    print(f"--- Creating Sandbox: {payload['name']} ---")
    response = requests.post(f"{api_url}/sandbox", headers=headers, json=payload)
    print(f"Status: {response.status_code}")
    if response.status_code >= 400:
        print(f"Error: {response.text}")
        return None
    
    data = response.json()
    print(f"Sandbox ID: {data['id']}")
    return data['id']

def wait_for_sandbox(sandbox_id):
    print(f"--- Waiting for Sandbox {sandbox_id} to be STARTED ---")
    for _ in range(30):
        response = requests.get(f"{api_url}/sandbox/{sandbox_id}", headers=headers)
        if response.status_code == 200:
            data = response.json()
            state = data.get('state')
            print(f"Current State: {state}")
            if state == 'started':
                return True
            if state == 'error':
                print(f"Fatal Error: {data.get('errorReason')}")
                return False
        else:
            print(f"Failed to poll: {response.status_code} {response.text}")
        
        time.sleep(5)
    return False

if __name__ == "__main__":
    sid = create_sandbox()
    if sid:
        success = wait_for_sandbox(sid)
        if success:
            print("\n✅ REPRO SUCCESS: Sandbox is started!")
        else:
            print("\n❌ REPRO FAILED: Sandbox did not reach started state.")
