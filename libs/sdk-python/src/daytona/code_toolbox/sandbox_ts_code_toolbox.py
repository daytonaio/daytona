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
        # ts-node does not support reading from stdin via - or /dev/stdin when stdin is a pipe,
        # so write to a temp file, execute it, then clean up
        # Capture the exit code before filtering to preserve ts-node's exit status
        parts = [
            "_f=/tmp/dtn_$$.ts",
            f"printf '%s' '{base64_code}' | base64 -d > \"$_f\"",
            f'_dtn_out=$(npx ts-node -O \'{{"module":"CommonJS"}}\' "$_f" {argv} 2>&1)',
            "_dtn_ec=$?",
            'rm -f "$_f"',
            "printf '%s\\n' \"$_dtn_out\" | grep -v 'npm notice'",
            "exit $_dtn_ec",
        ]
        return "; ".join(parts)
