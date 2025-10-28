#!/usr/bin/env python3
"""
WebSocket client that connects to the Go server and handles close code 4008
"""

import asyncio
import websockets
from websockets.exceptions import ConnectionClosed


async def test_websocket_client():
    uri = "ws://localhost:8088/ws"
    
    print(f"Connecting to {uri}...")
    
    try:
        async with websockets.connect(uri) as websocket:
            print("Connected successfully!")
            
            try:
                # Wait for messages or close frame
                while True:
                    message = await websocket.recv()
                    print(f"Received message: {message}")
                    
            except ConnectionClosed as e:
                print(f"\n🔴 Connection closed!")
                print(f"   Close code: {e.code}")
                print(f"   Close reason: {e.reason}")
                
                if e.code == 4008:
                    print(f"   ✅ Successfully received close code 4008!")
                else:
                    print(f"   ❌ Expected 4008, but got {e.code}")
                    
    except Exception as e:
        print(f"Error: {type(e).__name__}: {e}")


if __name__ == "__main__":
    print("=" * 60)
    print("WebSocket Client - Testing Close Code 4008")
    print("=" * 60)
    asyncio.run(test_websocket_client())
    print("=" * 60)

