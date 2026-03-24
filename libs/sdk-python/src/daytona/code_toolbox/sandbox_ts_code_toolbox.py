# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import base64

from ..common.process import CodeRunParams


class SandboxTsCodeToolbox:
    def get_run_command(self, code: str, params: CodeRunParams | None = None) -> str:
        # Prepend argv fix: ts-node places the script path at argv[1]; splice it out to match legacy node -e behaviour
        # Encode the provided code in base64
        base64_code = base64.b64encode(("process.argv.splice(1, 1);\n" + code).encode()).decode()

        # Build command-line arguments string
        argv = ""
        if params and params.argv:
            argv = " ".join(params.argv)

        # Pipe the base64-encoded code via stdin to avoid OS ARG_MAX limits on large payloads
        # ts-node does not support - for stdin; use shell PID ($$) for the temp file — each code_run spawns its own
        # shell process so $$ is unique across concurrent calls; cleaned up before exit
        # npm_config_loglevel=error suppresses npm notice/warn output at source, preserving streaming and real errors
        parts = [
            "_f=/tmp/dtn_$$.ts",
            f"printf '%s' '{base64_code}' | base64 -d > \"$_f\"",
            (
                f"npm_config_loglevel=error npx ts-node -T --ignore-diagnostics 5107"
                f' -O \'{{"module":"CommonJS"}}\' "$_f" {argv}'
            ),
            "_dtn_ec=$?",
            'rm -f "$_f"',
            "exit $_dtn_ec",
        ]
        return "; ".join(parts)
