from daytona import CreateSandboxFromSnapshotParams, Daytona


def main():
    daytona = Daytona()

    # Default interval
    sandbox1 = daytona.create()
    print(sandbox1.auto_archive_interval)

    # Set interval to 1 hour
    sandbox1.set_auto_archive_interval(60)
    print(sandbox1.auto_archive_interval)

    # Max interval
    sandbox2 = daytona.create(params=CreateSandboxFromSnapshotParams(auto_archive_interval=0))
    print(sandbox2.auto_archive_interval)

    # 1 day interval
    sandbox3 = daytona.create(params=CreateSandboxFromSnapshotParams(auto_archive_interval=1440))
    print(sandbox3.auto_archive_interval)

    sandbox1.delete()
    sandbox2.delete()
    sandbox3.delete()


if __name__ == "__main__":
    main()
