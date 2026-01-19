import sys
import threading
import time

from daytona import Daytona, PtySize, Sandbox


def interactive_pty_session(sandbox: Sandbox):
    print("=== First PTY Session: Interactive Command with Exit ===")

    pty_session_id = "interactive-pty-session"

    # Create PTY session (returns PtyHandle in sync version)
    pty_handle = sandbox.process.create_pty_session(id=pty_session_id, pty_size=PtySize(cols=120, rows=30))

    def handle_pty_data(data: bytes):
        # Decode UTF-8 bytes to text and write directly to preserve terminal formatting
        text = data.decode("utf-8", errors="replace")
        _ = sys.stdout.write(text)
        _ = sys.stdout.flush()

    # Send interactive command
    print("\nSending interactive read command...")
    pty_handle.send_input('printf "Enter your name: " && read name && printf "Hello, %s\\n" "$name"\n')

    # Wait and respond
    time.sleep(1)
    pty_handle.send_input("Alice\n")

    _ = pty_handle.resize(PtySize(cols=80, rows=25))

    # Send another command
    time.sleep(1)
    pty_handle.send_input("ls -la\n")

    # Send exit command
    time.sleep(1)
    pty_handle.send_input("exit\n")

    # Using iterator to handle PTY data
    print("\n--- Using iterator approach to handle PTY output ---")
    for data in pty_handle:
        handle_pty_data(data)

    print(f"\nPTY session exited with code: {pty_handle.exit_code}")
    if pty_handle.error:
        print(f"Error: {pty_handle.error}")


def kill_pty_session(sandbox: Sandbox):
    print("\n=== Second PTY Session: Kill PTY Session ===")

    pty_session_id = "kill-pty-session"

    # Create PTY session
    pty_handle = sandbox.process.create_pty_session(id=pty_session_id, pty_size=PtySize(cols=120, rows=30))

    def handle_pty_data(data: bytes):
        # Decode UTF-8 bytes to text and write directly to preserve terminal formatting
        text = data.decode("utf-8", errors="replace")
        _ = sys.stdout.write(text)
        _ = sys.stdout.flush()

    # Send a long-running command
    print("\nSending long-running command (infinite loop)...")
    pty_handle.send_input('while true; do echo "Running... $(date)"; sleep 1; done\n')

    # Using thread and wait() method to handle PTY output
    thread = threading.Thread(target=pty_handle.wait, args=(handle_pty_data, 10))
    thread.start()

    # Let it run for a few seconds
    time.sleep(3)

    # Kill the PTY session
    pty_handle.kill()

    thread.join()

    print(f"\nPTY session terminated. Exit code: {pty_handle.exit_code}")
    if pty_handle.error:
        print(f"Error: {pty_handle.error}")


def main():
    daytona = Daytona()
    sandbox = daytona.create()

    try:
        # interactive PTY session with exit
        interactive_pty_session(sandbox)
        # PTY session killed with .kill()
        kill_pty_session(sandbox)
    except Exception as error:
        print(f"Error executing PTY commands: {error}")
    finally:
        # Cleanup
        print(f"\nDeleting sandbox: {sandbox.id}")
        daytona.delete(sandbox)


if __name__ == "__main__":
    main()
