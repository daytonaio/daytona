# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""In-sandbox SDK EnvironmentWorker wrapper."""
from __future__ import annotations

import asyncio
import logging
import os

from anthropic import AsyncAnthropic

WORKDIR = "/mnt/session"


def runner_max_idle_seconds() -> float:
    return float(os.environ.get("RUNNER_MAX_IDLE_SECONDS", "300"))


async def main() -> None:
    logging.basicConfig(level=os.environ.get("LOG_LEVEL", "INFO"))

    environment_key = os.environ["ANTHROPIC_ENVIRONMENT_KEY"]
    async with AsyncAnthropic(auth_token=environment_key) as client:
        await client.beta.environments.work.worker(
            environment_key=environment_key,
            workdir=WORKDIR,
            unrestricted_paths=True,
            max_idle=runner_max_idle_seconds(),
        ).handle_item()


if __name__ == "__main__":
    asyncio.run(main())
