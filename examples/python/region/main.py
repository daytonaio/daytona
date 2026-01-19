import time

from daytona import CreateSandboxFromSnapshotParams, CreateSnapshotParams, Daytona, DaytonaConfig, Image


def main():
    daytona = Daytona(DaytonaConfig(target="us"))

    snapshot1 = f"us-{int(time.time() * 1000)}"
    print(f"Creating snapshot {snapshot1}")
    try:
        _ = daytona.snapshot.create(
            CreateSnapshotParams(
                name=snapshot1,
                image=Image.debian_slim("3.12"),
                region_id="us",
            )
        )
    except Exception as e:
        print(e)
    print("--------------------------------")

    snapshot2 = f"eu-{int(time.time() * 1000)}"
    print(f"Creating snapshot {snapshot2}")
    try:
        _ = daytona.snapshot.create(
            CreateSnapshotParams(
                name=snapshot2,
                image=Image.debian_slim("3.13"),
                region_id="eu",
            )
        )
    except Exception as e:
        print("error", e)
    print("--------------------------------")

    print(f"Creating sandbox from snapshot {snapshot1}")
    try:
        sandbox = daytona.create(CreateSandboxFromSnapshotParams(snapshot=snapshot1))
        daytona.delete(sandbox)
    except Exception as e:
        print(e)
    print("--------------------------------")

    print(f"Creating sandbox from snapshot {snapshot2}")
    try:
        sandbox = daytona.create(CreateSandboxFromSnapshotParams(snapshot=snapshot2))
        daytona.delete(sandbox)
    except Exception as e:
        print("error", e)


if __name__ == "__main__":
    main()
