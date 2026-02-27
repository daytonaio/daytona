"""Build a Daytona snapshot for finqa_env.

Creates a snapshot from the openenv-base image with finqa_env installed
and the FinQA dataset pre-downloaded from HuggingFace. Sandboxes created
from this snapshot start in ~2s with no network dependencies.

Usage:
    python build_snapshot.py
    python build_snapshot.py --snapshot-name my-finqa
"""

import argparse
import os

from daytona import CreateSnapshotParams, Daytona, DaytonaConfig, Image
from dotenv import load_dotenv

OPENENV_REPO = "https://github.com/meta-pytorch/OpenEnv.git"
SERVER_CMD = "cd /app/env && uvicorn finqa_env.server.app:app --host 0.0.0.0 --port 8000"


def main():
    load_dotenv()

    parser = argparse.ArgumentParser(description="Build a Daytona snapshot for finqa_env.")
    parser.add_argument(
        "--snapshot-name",
        default="openenv-finqa",
        help="Name for the snapshot (default: openenv-finqa).",
    )
    args = parser.parse_args()

    daytona = Daytona(DaytonaConfig(api_key=os.environ.get("DAYTONA_API_KEY")))

    image = (
        Image.base("ghcr.io/meta-pytorch/openenv-base:latest")
        .workdir("/app/env")
        .run_commands(
            "apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*",
            f"git clone --depth 1 {OPENENV_REPO} /tmp/openenv",
            "cp -r /tmp/openenv/envs/finqa_env/. /app/env/",
            "rm -rf /tmp/openenv",
        )
        .run_commands(
            "pip install --no-cache-dir -e .",
            # Pre-download FinQA dataset from HuggingFace (same as Dockerfile)
            'python -c "'
            "from huggingface_hub import snapshot_download; "
            "snapshot_download('snorkelai/finqa-data', repo_type='dataset', local_dir='/app/env/data')"
            '"',
        )
        .env(
            {
                "PYTHONUNBUFFERED": "1",
                "FINQA_DATA_PATH": "/app/env/data",
            }
        )
        .cmd(["sh", "-c", SERVER_CMD])
    )

    print(f"Building snapshot '{args.snapshot_name}'...")
    print("This may take 3-5 minutes on first build.\n")

    snapshot = daytona.snapshot.create(
        CreateSnapshotParams(name=args.snapshot_name, image=image),
        on_logs=lambda line: print(f"  {line}"),
    )

    print(f"\nSnapshot ready: {snapshot.name} (state={snapshot.state})")
    print("Use it with: python run.py")


if __name__ == "__main__":
    main()
