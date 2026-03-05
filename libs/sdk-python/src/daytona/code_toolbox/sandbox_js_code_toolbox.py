# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import base64

from ..common.process import CodeRunParams


class SandboxJsCodeToolbox:
    def get_run_command(self, code: str, params: CodeRunParams | None = None) -> str:
        # Encode the provided code in base64
        base64_code = base64.b64encode(code.encode()).decode()

        # Build command-line arguments string
        argv = ""
        if params and params.argv:
            argv = " ".join(params.argv)

        # Pipe the base64-encoded code via stdin to avoid OS ARG_MAX limits on large payloads
        # Use /dev/stdin instead of -e "$(cat)" which would expand as a process arg and hit ARG_MAX
        # Capture the exit code before filtering to preserve node's exit status
        return (
            f"_dtn_out=$(echo '{base64_code}' | base64 -d | node /dev/stdin {argv} 2>&1); _dtn_ec=$?; "
            f"printf '%s\\n' \"$_dtn_out\" | grep -v 'npm notice'; exit $_dtn_ec"
        )
