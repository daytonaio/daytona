# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""
Build the default BYOC sandbox snapshot from Dockerfile.default.

Names the snapshot using host_lib.default_snapshot_name() so the orchestrator
and this builder cannot drift, and skips the build if a snapshot with that
name already exists. Streams build logs.
"""
import sys

import dotenv
import host_lib

from daytona import CreateSnapshotParams, Daytona, DaytonaNotFoundError, Image, Resources

dotenv.load_dotenv(override=True)

DOCKERFILE = host_lib.DEFAULT_DOCKERFILE
RESOURCES = Resources(cpu=2, memory=8, disk=10)


def main() -> int:
    if not DOCKERFILE.is_file():
        print(f"!! {DOCKERFILE} not found", file=sys.stderr)
        return 1

    name = host_lib.default_snapshot_name()

    daytona = Daytona()
    try:
        existing = daytona.snapshot.get(name)
        print(f"snapshot {name} already exists (id {existing.id}); skipping build")
        return 0
    except DaytonaNotFoundError:
        pass

    print(f"building snapshot {name} from {DOCKERFILE} (this is slow)...")
    snapshot = daytona.snapshot.create(
        CreateSnapshotParams(
            name=name,
            image=Image.from_dockerfile(str(DOCKERFILE)),
            resources=RESOURCES,
        ),
        on_logs=lambda chunk: print(chunk, end="", flush=True),
    )
    print(f"\n\ncreated snapshot {snapshot.name} (id {snapshot.id})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
