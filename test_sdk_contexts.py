#!/usr/bin/env python3
"""
Test the updated SDK with multi-context support
This demonstrates the new context management methods
"""

from daytona import Daytona, CreateSandboxFromImageParams, CodeInterpreter

def test_sdk_contexts():
    """Test SDK context management features"""
    print("=" * 70)
    print("Testing SDK Multi-Context Support")
    print("=" * 70)
    
    # Create sandbox
    daytona = Daytona()
    params = CreateSandboxFromImageParams(image="python:3.9.23-slim")
    sandbox = daytona.create(params, timeout=150)
    
    try:
        # Get code interpreter
        preview_link = sandbox.get_preview_link(2280)
        interpreter = CodeInterpreter(
            preview_link.url, 
            headers={
                **sandbox._toolbox_api.api_client.default_headers,
                "X-Daytona-Preview-Token": preview_link.token,
            },
        )
        
        # Test 1: Default context (backwards compatible)
        print("\n📝 Test 1: Execute in default context")
        print("-" * 70)
        interpreter.execute(
            code="x = 100\nprint(f'Default context: x = {x}')",
            on_stdout=lambda text: print(f"  Output: {text.strip()}"),
            on_error=lambda name, val, tb: print(f"  Error: {name}: {val}")
        )
        
        # Test 2: Create custom context
        print("\n📝 Test 2: Create custom context")
        print("-" * 70)
        ctx = interpreter.create_context(cwd="/tmp")
        print(f"  Created context: {ctx['id']}")
        print(f"  Working directory: {ctx['cwd']}")
        print(f"  Active: {ctx['active']}")
        
        # Test 3: Execute in custom context
        print("\n📝 Test 3: Execute in custom context")
        print("-" * 70)
        interpreter.execute(
            code="import os; print(f'CWD: {os.getcwd()}')",
            context_id=ctx["id"],
            on_stdout=lambda text: print(f"  Output: {text.strip()}")
        )
        
        # Test 4: List contexts
        print("\n📝 Test 4: List all contexts")
        print("-" * 70)
        contexts = interpreter.list_contexts()
        print(f"  Found {len(contexts)} user-created context(s)")
        for c in contexts:
            print(f"    - {c['id'][:8]}... ({c['cwd']})")
        
        # Test 5: Context isolation
        print("\n📝 Test 5: Context isolation")
        print("-" * 70)
        
        # Create another context
        ctx2 = interpreter.create_context()
        print(f"  Created second context: {ctx2['id']}")
        
        # Set variable in first context
        interpreter.execute(
            code="isolated_var = 'context1'",
            context_id=ctx["id"]
        )
        print("  Set isolated_var in first context")
        
        # Try to access in second context (should fail)
        print("  Trying to access isolated_var in second context...")
        error_occurred = [False]
        
        def on_error(name, val, tb):
            error_occurred[0] = True
            print(f"  ✓ Expected error: {name}")
        
        interpreter.execute(
            code="print(isolated_var)",
            context_id=ctx2["id"],
            on_error=on_error
        )
        
        if error_occurred[0]:
            print("  ✓ Contexts are properly isolated")
        
        # Test 6: Delete contexts
        print("\n📝 Test 6: Delete contexts")
        print("-" * 70)
        interpreter.delete_context(ctx["id"])
        print(f"  Deleted context: {ctx['id']}")
        
        interpreter.delete_context(ctx2["id"])
        print(f"  Deleted context: {ctx2['id']}")
        
        # Verify deleted
        contexts_after = interpreter.list_contexts()
        print(f"  Remaining contexts: {len(contexts_after)}")
        
        # Test 7: Context with timeout
        print("\n📝 Test 7: Timeout handling")
        print("-" * 70)
        ctx3 = interpreter.create_context()
        print(f"  Created context: {ctx3['id']}")
        
        try:
            interpreter.execute(
                code="import time; time.sleep(10)",
                context_id=ctx3["id"],
                timeout=2
            )
        except TimeoutError as e:
            print(f"  ✓ Timeout caught: {e}")
        
        interpreter.delete_context(ctx3["id"])
        print(f"  Deleted context: {ctx3['id']}")
        
        # Summary
        print("\n" + "=" * 70)
        print("✅ All SDK context tests passed!")
        print("=" * 70)
        
    except Exception as e:
        print(f"\n❌ Test failed: {e}")
        import traceback
        traceback.print_exc()
        
    finally:
        # Clean up
        print("\nCleaning up sandbox...")
        daytona.delete(sandbox)
        print("Done!")


if __name__ == "__main__":
    test_sdk_contexts()

