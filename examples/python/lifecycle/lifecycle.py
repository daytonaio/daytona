from daytona import Daytona, Resources


def main():
    daytona = Daytona()

    print("Creating sandbox")
    sandbox = daytona.create()
    print("Sandbox created")

    _ = sandbox.set_labels(
        {
            "public": "true",
        }
    )

    print("Stopping sandbox")
    daytona.stop(sandbox)
    print("Sandbox stopped")

    print("Starting sandbox")
    daytona.start(sandbox)
    print("Sandbox started")

    print("Getting existing sandbox")
    existing_sandbox = daytona.get(sandbox.id)
    print("Get existing sandbox")

    response = existing_sandbox.process.exec('echo "Hello World from exec!"', cwd="/home/daytona", timeout=10)
    if response.exit_code != 0:
        print(f"Error: {response.exit_code} {response.result}")
    else:
        print(response.result)

    result = daytona.list()
    print("Total sandboxes count:", result.total)

    print(f"Printing first sandbox -> id: {result.items[0].id} state: {result.items[0].state}")

    # Hot resize: can only increase CPU and memory on a running sandbox
    # The resize() method automatically waits for the resize operation to complete
    print("Resizing sandbox (hot resize)...")
    sandbox.resize(Resources(cpu=2, memory=2), hot=True, timeout=120)

    # After resize completes, the sandbox object is automatically updated with new resources
    print(f"Resize complete! New resources: CPU={sandbox.cpu}, Memory={sandbox.memory}GB, Disk={sandbox.disk}GB")

    print("Removing sandbox")
    daytona.delete(sandbox)
    print("Sandbox removed")


if __name__ == "__main__":
    main()
