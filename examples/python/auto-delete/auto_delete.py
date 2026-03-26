from daytona import CreateSandboxFromSnapshotParams, Daytona


def main():
    daytona = Daytona()

    # Auto-delete is disabled by default
    sandbox1 = daytona.create()
    print(sandbox1.auto_delete_interval)

    # Auto-delete after the Sandbox has been stopped for 1 hour
    sandbox1.set_auto_delete_interval(60)
    print(sandbox1.auto_delete_interval)

    # Delete immediately upon stopping
    sandbox1.set_auto_delete_interval(0)
    print(sandbox1.auto_delete_interval)

    # Disable auto-delete
    sandbox1.set_auto_delete_interval(-1)
    print(sandbox1.auto_delete_interval)

    # Auto-delete after the Sandbox has been stopped for 1 day
    sandbox2 = daytona.create(params=CreateSandboxFromSnapshotParams(auto_delete_interval=1440))
    print(sandbox2.auto_delete_interval)


if __name__ == "__main__":
    main()
