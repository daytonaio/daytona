# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import base64

from ..common.process import CodeRunParams


class SandboxTsCodeToolbox:
    def get_run_command(self, code: str, params: CodeRunParams | None = None) -> str:
        # Encode the provided code in base64
        base64_code = base64.b64encode(code.encode()).decode()

        # Build command-line arguments string
        argv = ""
        if params and params.argv:
            argv = " ".join(params.argv)

        # Pipe the base64-encoded code via stdin to avoid OS ARG_MAX limits on large payloads
        # Use /dev/stdin instead of -e "$(cat)" which would expand as a process arg and hit ARG_MAX
        return (
            f"""echo '{base64_code}' | base64 --decode | npx ts-node -O """
            f"""'{{"module":"CommonJS"}}' /dev/stdin {argv} 2>&1 | grep -vE 'npm notice'"""
        )
