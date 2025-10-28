#!/usr/bin/env python3
"""
Test script for multi-context interpreter support
Run this after deploying the server changes
"""

import asyncio
import json
import requests
import websockets
import time

BASE_URL = "http://localhost:3987"
WS_URL = "ws://localhost:3987"


def test_create_context():
    """Test 1: Create a new context"""
    print("\n" + "="*60)
    print("TEST 1: Create Context")
    print("="*60)
    
    response = requests.post(f"{BASE_URL}/process/interpreter/context", json={
        "cwd": "/tmp/test-context",
        "language": "python"
    })
    
    print(f"Status Code: {response.status_code}")
    data = response.json()
    print(f"Response: {json.dumps(data, indent=2)}")
    
    assert response.status_code == 200, "Failed to create context"
    assert "id" in data, "Context ID not in response"
    assert data["language"] == "python", "Language mismatch"
    assert data["active"] == True, "Context not active"
    
    print("✅ PASSED: Context created successfully")
    return data["id"]


def test_list_contexts():
    """Test 2: List user-created contexts (excludes default)"""
    print("\n" + "="*60)
    print("TEST 2: List Contexts")
    print("="*60)
    
    response = requests.get(f"{BASE_URL}/process/interpreter/context")
    
    print(f"Status Code: {response.status_code}")
    data = response.json()
    print(f"Response: {json.dumps(data, indent=2)}")
    
    assert response.status_code == 200, "Failed to list contexts"
    assert "contexts" in data, "Contexts not in response"
    
    # Default context should NOT be in the list
    has_default = any(ctx["id"] == "default" for ctx in data["contexts"])
    assert not has_default, "Default context should not be in user context list"
    
    print("✅ PASSED: Contexts listed successfully (default excluded)")
    return data["contexts"]


async def test_execute_default_context():
    """Test 3: Execute in default context"""
    print("\n" + "="*60)
    print("TEST 3: Execute in Default Context")
    print("="*60)
    
    uri = f"{WS_URL}/process/interpreter/execute"
    
    async with websockets.connect(uri) as ws:
        # Send execution request (no contextId = use default)
        request = {
            "code": "x = 100\nprint(f'x = {x}')"
        }
        await ws.send(json.dumps(request))
        
        outputs = []
        try:
            while True:
                msg = await asyncio.wait_for(ws.recv(), timeout=5)
                data = json.loads(msg)
                print(f"Received: {data}")
                outputs.append(data)
                
                if data.get("type") == "control" and data.get("text") in ["completed", "error_completed"]:
                    break
        except (websockets.exceptions.ConnectionClosed, asyncio.TimeoutError):
            pass
    
    # Verify output
    stdout_chunks = [o for o in outputs if o.get("type") == "stdout"]
    assert len(stdout_chunks) > 0, "No stdout received"
    assert "x = 100" in stdout_chunks[0].get("text", ""), "Wrong output"
    
    print("✅ PASSED: Executed in default context")


async def test_execute_specific_context(context_id):
    """Test 4: Execute in specific context"""
    print("\n" + "="*60)
    print(f"TEST 4: Execute in Specific Context ({context_id})")
    print("="*60)
    
    uri = f"{WS_URL}/process/interpreter/execute"
    
    async with websockets.connect(uri) as ws:
        # Send execution request with specific contextId
        request = {
            "code": "y = 200\nprint(f'y = {y}')",
            "contextId": context_id
        }
        await ws.send(json.dumps(request))
        
        outputs = []
        try:
            while True:
                msg = await asyncio.wait_for(ws.recv(), timeout=5)
                data = json.loads(msg)
                print(f"Received: {data}")
                outputs.append(data)
                
                if data.get("type") == "control" and data.get("text") in ["completed", "error_completed"]:
                    break
        except (websockets.exceptions.ConnectionClosed, asyncio.TimeoutError):
            pass
    
    # Verify output
    stdout_chunks = [o for o in outputs if o.get("type") == "stdout"]
    assert len(stdout_chunks) > 0, "No stdout received"
    assert "y = 200" in stdout_chunks[0].get("text", ""), "Wrong output"
    
    print("✅ PASSED: Executed in specific context")


async def test_context_isolation():
    """Test 5: Verify context isolation"""
    print("\n" + "="*60)
    print("TEST 5: Context Isolation")
    print("="*60)
    
    # Create two contexts
    ctx1_response = requests.post(f"{BASE_URL}/process/interpreter/context", json={"language": "python"})
    ctx1_id = ctx1_response.json()["id"]
    print(f"Context 1 ID: {ctx1_id}")
    
    ctx2_response = requests.post(f"{BASE_URL}/process/interpreter/context", json={"language": "python"})
    ctx2_id = ctx2_response.json()["id"]
    print(f"Context 2 ID: {ctx2_id}")
    
    uri = f"{WS_URL}/process/interpreter/execute"
    
    # Set variable in context 1
    async with websockets.connect(uri) as ws:
        request = {"code": "isolated_var = 'context1'", "contextId": ctx1_id}
        await ws.send(json.dumps(request))
        async for msg in ws:
            data = json.loads(msg)
            if data.get("type") == "control" and data.get("text") in ["completed", "error_completed"]:
                break
    
    print("Set isolated_var = 'context1' in context 1")
    
    # Try to access variable in context 2 (should fail)
    async with websockets.connect(uri) as ws:
        request = {"code": "print(isolated_var)", "contextId": ctx2_id}
        await ws.send(json.dumps(request))
        
        has_error = False
        async for msg in ws:
            data = json.loads(msg)
            print(f"Received: {data}")
            if data.get("type") == "error" and "NameError" in data.get("name", ""):
                has_error = True
                break
            if data.get("type") == "control":
                break
    
    assert has_error, "Variable should not be accessible in different context"
    
    print("✅ PASSED: Contexts are properly isolated")


async def test_context_restart_after_exit(context_id):
    """Test 6: Context restarts after exit()"""
    print("\n" + "="*60)
    print("TEST 6: Context Restart After Exit")
    print("="*60)
    
    uri = f"{WS_URL}/process/interpreter/execute"
    
    # Call exit() in the context
    async with websockets.connect(uri) as ws:
        request = {"code": "exit()", "contextId": context_id}
        await ws.send(json.dumps(request))
        
        try:
            async for msg in ws:
                data = json.loads(msg)
                print(f"Received: {data}")
                if data.get("type") == "control" and data.get("text") == "exit":
                    break
        except websockets.exceptions.ConnectionClosed:
            pass
    
    print("Called exit() in context")
    time.sleep(1)  # Give it a moment
    
    # Try to execute again (should auto-restart)
    async with websockets.connect(uri) as ws:
        request = {"code": "print('Restarted!')", "contextId": context_id}
        await ws.send(json.dumps(request))
        
        restarted = False
        async for msg in ws:
            data = json.loads(msg)
            print(f"Received: {data}")
            if data.get("type") == "stdout" and "Restarted!" in data.get("text", ""):
                restarted = True
            if data.get("type") == "control" and data.get("text") in ["completed", "error_completed"]:
                break
    
    assert restarted, "Context did not restart after exit"
    
    print("✅ PASSED: Context restarted successfully")


async def test_nonexistent_context():
    """Test 7: Error when using non-existent context"""
    print("\n" + "="*60)
    print("TEST 7: Non-existent Context Error")
    print("="*60)
    
    uri = f"{WS_URL}/process/interpreter/execute"
    
    async with websockets.connect(uri) as ws:
        request = {"code": "print('test')", "contextId": "nonexistent-context-id"}
        await ws.send(json.dumps(request))
        
        has_error = False
        try:
            msg = await asyncio.wait_for(ws.recv(), timeout=2)
            data = json.loads(msg)
            print(f"Received: {data}")
            
            if data.get("type") == "error" and "ContextError" in data.get("name", ""):
                has_error = True
        except (websockets.exceptions.ConnectionClosed, asyncio.TimeoutError):
            pass
    
    assert has_error, "Should receive ContextError for non-existent context"
    
    print("✅ PASSED: Proper error for non-existent context")


def test_delete_context(context_id):
    """Test 8: Delete a context"""
    print("\n" + "="*60)
    print("TEST 8: Delete Context")
    print("="*60)
    
    response = requests.delete(f"{BASE_URL}/process/interpreter/context/{context_id}")
    
    print(f"Status Code: {response.status_code}")
    data = response.json()
    print(f"Response: {json.dumps(data, indent=2)}")
    
    assert response.status_code == 200, "Failed to delete context"
    assert "message" in data, "Message not in response"
    
    print("✅ PASSED: Context deleted successfully")


def test_cannot_delete_default():
    """Test 9: Cannot delete default context"""
    print("\n" + "="*60)
    print("TEST 9: Cannot Delete Default Context")
    print("="*60)
    
    response = requests.delete(f"{BASE_URL}/process/interpreter/context/default")
    
    print(f"Status Code: {response.status_code}")
    data = response.json()
    print(f"Response: {json.dumps(data, indent=2)}")
    
    assert response.status_code == 400, "Should not be able to delete default context"
    assert "error" in data, "Error message not in response"
    
    print("✅ PASSED: Default context protected from deletion")


def test_invalid_language():
    """Test 10: Error on invalid language"""
    print("\n" + "="*60)
    print("TEST 10: Invalid Language Error")
    print("="*60)
    
    response = requests.post(f"{BASE_URL}/process/interpreter/context", json={
        "language": "javascript"
    })
    
    print(f"Status Code: {response.status_code}")
    data = response.json()
    print(f"Response: {json.dumps(data, indent=2)}")
    
    assert response.status_code == 400, "Should reject invalid language"
    assert "error" in data, "Error message not in response"
    
    print("✅ PASSED: Invalid language properly rejected")


async def main():
    print("\n" + "="*70)
    print("    MULTI-CONTEXT INTERPRETER TEST SUITE")
    print("="*70)
    print("\nMake sure the daemon is running on port 3987")
    print("Start with: cd apps/daemon && go run cmd/daemon/main.go\n")
    
    try:
        # Test 1: Create context
        context_id = test_create_context()
        
        # Test 2: List contexts
        test_list_contexts()
        
        # Test 3: Execute in default context
        await test_execute_default_context()
        
        # Test 4: Execute in specific context
        await test_execute_specific_context(context_id)
        
        # Test 5: Context isolation
        await test_context_isolation()
        
        # Test 6: Context restart after exit
        await test_context_restart_after_exit(context_id)
        
        # Test 7: Non-existent context error
        await test_nonexistent_context()
        
        # Test 8: Delete context
        test_delete_context(context_id)
        
        # Test 9: Cannot delete default
        test_cannot_delete_default()
        
        # Test 10: Invalid language
        test_invalid_language()
        
        # Summary
        print("\n" + "="*70)
        print("    TEST SUMMARY")
        print("="*70)
        print("✅ All 10 tests PASSED!")
        print("\nMulti-context support is working correctly!")
        print("="*70)
        
        return 0
        
    except Exception as e:
        print(f"\n❌ TEST FAILED: {e}")
        import traceback
        traceback.print_exc()
        return 1


if __name__ == "__main__":
    exit_code = asyncio.run(main())
    exit(exit_code)

