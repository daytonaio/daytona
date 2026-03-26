import asyncio
import sys

from daytona import AsyncDaytona, AsyncSandbox, PtySize


async def interactive_pty_session(sandbox: AsyncSandbox):
    print("=== First PTY Session: Interactive Command with Exit ===")

    pty_session_id = "interactive-pty-session"

    # Create PTY session with data handler
    def handle_pty_data(data: bytes):
        # Decode UTF-8 bytes to text and write directly to preserve terminal formatting
        text = data.decode("utf-8", errors="replace")
        _ = sys.stdout.write(text)
        _ = sys.stdout.flush()

    # Create PTY session with custom dimensions and data handler
    pty_handle = await sandbox.process.create_pty_session(
        id=pty_session_id, pty_size=PtySize(cols=120, rows=30), on_data=handle_pty_data
    )

    # Send interactive command
    await pty_handle.send_input('printf "Enter your name: " && read name && printf "Hello, %s\\n" "$name"\n')

    # Wait and respond
    await asyncio.sleep(1)
    await pty_handle.send_input("Bob\n")

    await asyncio.sleep(1)
    pty_session_info = await sandbox.process.resize_pty_session(pty_session_id, PtySize(cols=80, rows=25))
    print(f"\nPTY session resized to {pty_session_info.cols}x{pty_session_info.rows}")

    # Send another command
    await asyncio.sleep(1)
    await pty_handle.send_input("ls -la\n")

    # Send exit command
    await asyncio.sleep(1)
    await pty_handle.send_input("exit\n")

    # Wait for PTY to exit
    result = await pty_handle.wait()
    print(f"\nPTY session exited with code: {result.exit_code}")
    if result.error:
        print(f"Error: {result.error}")


async def kill_pty_session(sandbox: AsyncSandbox):
    print("\n=== Second PTY Session: Kill PTY Session ===")

    pty_session_id = "kill-pty-session"

    # Create PTY session with data handler
    def handle_pty_data(data: bytes):
        # Decode UTF-8 bytes to text and write directly to preserve terminal formatting
        text = data.decode("utf-8", errors="replace")
        _ = sys.stdout.write(text)
        _ = sys.stdout.flush()

    # Create PTY session
    pty_handle = await sandbox.process.create_pty_session(
        id=pty_session_id, pty_size=PtySize(cols=120, rows=30), on_data=handle_pty_data
    )

    # Send a long-running command
    print("\nSending long-running command (infinite loop)...")
    await pty_handle.send_input('while true; do echo "Running... $(date)"; sleep 1; done\n')

    # Let it run for a few seconds
    await asyncio.sleep(3)

    # Kill the PTY session
    await pty_handle.kill()

    # Wait for PTY to terminate
    result = await pty_handle.wait()
    print(f"\nPTY session terminated. Exit code: {result.exit_code}")
    if result.error:
        print(f"Error: {result.error}")


async def main():
    async with AsyncDaytona() as daytona:
        sandbox = await daytona.create()

        try:
            # Interactive PTY session with exit
            await interactive_pty_session(sandbox)
            # PTY session killed with .kill()
            await kill_pty_session(sandbox)
        except Exception as error:
            print(f"Error executing PTY commands: {error}")
        finally:
            print(f"\nDeleting sandbox: {sandbox.id}")
            await daytona.delete(sandbox)


if __name__ == "__main__":
    asyncio.run(main())
