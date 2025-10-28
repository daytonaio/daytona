import time

from daytona import (
    CreateSnapshotParams,
    Daytona,
    Image,
)


def main():
    daytona = Daytona()

    snapshot_name = f"kepler-test:{int(time.time())}"

    with open("file.txt", "w") as f:
        f.write("Hello kepler 1!")

    image = (
        Image.debian_slim("3.12")
        .add_local_file("file.txt", "./file.txt")
    )

    print(f"=== Creating Snapshot: {snapshot_name} ===")
    daytona.snapshot.create(
        CreateSnapshotParams(
            name=snapshot_name,
            image=image,
        ),
        on_logs=print,
    )


if __name__ == "__main__":
    main()