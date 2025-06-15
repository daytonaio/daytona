from daytona import Daytona


def main():
    daytona = Daytona()

    print("Creating sandbox")
    sandbox = daytona.create()
    print("Sandbox created")

    sandbox.set_labels(
        {
            "public": True,
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

    sandboxes = daytona.list()
    print("Total sandboxes count:", len(sandboxes))

    print(f"Printing sandboxes[0] -> id: {sandboxes[0].id} state: {sandboxes[0].state}")

    print("Removing sandbox")
    daytona.delete(sandbox)
    print("Sandbox removed")


if __name__ == "__main__":
    main()
