from daytona import Daytona, ListSandboxesQuery, SandboxListSortDirection, SandboxListSortField, SandboxState


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

    for sb in daytona.list(
        ListSandboxesQuery(
            limit=10,
            labels={"env": "dev"},
            states=[SandboxState.STARTED],
            sort=SandboxListSortField.CREATEDAT,
            order=SandboxListSortDirection.DESC,
        )
    ):
        print(sb.id)

    print("Removing sandbox")
    daytona.delete(sandbox)
    print("Sandbox removed")


if __name__ == "__main__":
    main()
