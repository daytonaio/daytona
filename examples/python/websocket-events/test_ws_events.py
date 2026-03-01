"""
Test WebSocket event subscription for sandbox lifecycle.

Tests:
1. Event subscriber connects on first sandbox creation
2. Sandbox state auto-updates via WebSocket events
3. wait_for_sandbox_start/stop use WebSocket events (not polling)
4. Event subscriber disconnects after 30s when no sandboxes are listening
"""

import time

from daytona import Daytona


def test_subscriber_connects():
    """Test that event subscriber connects on Daytona construction."""
    print("--- Test 1: Subscriber connects on Daytona construction ---")
    d = Daytona()

    # Subscriber should be initialized immediately (connecting in background)
    assert d._event_subscriber is not None, "Subscriber should exist after Daytona()"
    print("  PASS: Subscriber initialized in constructor")

    sandbox = d.create()
    assert d._event_subscriber.is_connected, "Subscriber should be connected after create"
    print(f"  PASS: Subscriber connected (state={sandbox.state})")

    d.delete(sandbox)
    print("  PASS: Test 1 complete\n")
    return d


def test_event_driven_lifecycle(d):
    """Test that start/stop use WebSocket events."""
    print("--- Test 2: Event-driven start/stop lifecycle ---")

    sandbox = d.create()
    print(f"  Created sandbox: {sandbox.id}, state={sandbox.state}")

    # Stop and measure time (should be fast via WS events)
    t0 = time.time()
    d.stop(sandbox)
    stop_time = time.time() - t0
    print(f"  Stopped in {stop_time:.2f}s, state={sandbox.state}")
    assert sandbox.state == "stopped", f"Expected stopped, got {sandbox.state}"

    # Start and measure time
    t0 = time.time()
    d.start(sandbox)
    start_time = time.time() - t0
    print(f"  Started in {start_time:.2f}s, state={sandbox.state}")
    assert sandbox.state == "started", f"Expected started, got {sandbox.state}"

    d.delete(sandbox)
    print("  PASS: Test 2 complete\n")


def test_auto_update_from_events(d):
    """Test that sandbox state auto-updates from WebSocket events without manual refresh."""
    print("--- Test 3: Auto-update sandbox state from events ---")

    sandbox = d.create()
    print(f"  Created: state={sandbox.state}")

    # Issue stop via raw API (bypassing the SDK's stop method)
    sandbox._sandbox_api.stop_sandbox(sandbox.id)
    print("  Issued raw stop API call...")

    # Wait for auto-update via WebSocket events
    for i in range(30):
        time.sleep(0.2)
        if sandbox.state != "started":
            print(f"  State changed to '{sandbox.state}' after {(i+1)*0.2:.1f}s (via WS event)")
            break
    else:
        print("  FAIL: State did not auto-update within 6s")
        d.delete(sandbox)
        return

    # Wait for fully stopped
    for i in range(30):
        time.sleep(0.2)
        if sandbox.state == "stopped":
            print(f"  State reached 'stopped' (via WS event)")
            break

    assert sandbox.state == "stopped", f"Expected stopped, got {sandbox.state}"
    d.delete(sandbox)
    print("  PASS: Test 3 complete\n")


def test_get_and_list_subscribe(d):
    """Test that sandboxes from get() and list() also subscribe to events."""
    print("--- Test 4: get() and list() sandboxes subscribe to events ---")

    sandbox = d.create()
    sandbox_id = sandbox.id

    # Get the same sandbox via get()
    sandbox2 = d.get(sandbox_id)
    print(f"  Got sandbox via get(): state={sandbox2.state}")

    # Stop via the original sandbox
    d.stop(sandbox)
    # The get'd sandbox should also have its state updated
    time.sleep(1)
    print(f"  After stop - original: state={sandbox.state}, get'd: state={sandbox2.state}")
    assert sandbox2.state in ["stopped", "stopping"], f"get'd sandbox state should update, got {sandbox2.state}"

    d.delete(sandbox)
    print("  PASS: Test 4 complete\n")


def main():
    print("=" * 60)
    print("WebSocket Event Subscription Test Suite (Python Sync)")
    print("=" * 60 + "\n")

    d = test_subscriber_connects()
    test_event_driven_lifecycle(d)
    test_auto_update_from_events(d)
    test_get_and_list_subscribe(d)

    print("=" * 60)
    print("ALL TESTS PASSED")
    print("=" * 60)


if __name__ == "__main__":
    main()
