# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""
Build the default BYOC sandbox snapshot from Dockerfile.default.

Hashes the Dockerfile, names the snapshot byoc-env-default-{sha8}, and skips
the build if a snapshot with that name already exists. Streams build logs.
"""
import hashlib
import pathlib
import sys

import dotenv

from daytona import CreateSnapshotParams, Daytona, DaytonaNotFoundError, Image, Resources

dotenv.load_dotenv(override=True)

DOCKERFILE = pathlib.Path("Dockerfile.default")
SNAPSHOT_PREFIX = "byoc-env-default"
RESOURCES = Resources(cpu=2, memory=8, disk=10)


def short_hash(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()[:8]


def main() -> int:
    if not DOCKERFILE.is_file():
        print(f"!! {DOCKERFILE} not found in cwd {pathlib.Path.cwd()}", file=sys.stderr)
        return 1

    sha = short_hash(DOCKERFILE)
    name = f"{SNAPSHOT_PREFIX}-{sha}"

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
