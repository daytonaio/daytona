# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import base64

from ..common.process import CodeRunParams


class SandboxJsCodeToolbox:
    def get_run_command(self, code: str, params: CodeRunParams | None = None) -> str:
        # Prepend argv fix: node - places '-' at argv[1]; splice it out to match legacy node -e behaviour
        # Encode the provided code in base64
        base64_code = base64.b64encode(("process.argv.splice(1, 1);\n" + code).encode()).decode()

        # Build command-line arguments string
        argv = ""
        if params and params.argv:
            argv = " ".join(params.argv)

        # Pipe the base64-encoded code via stdin to avoid OS ARG_MAX limits on large payloads
        # Use node - to read from stdin (node /dev/stdin does not work when stdin is a pipe)
        return f"printf '%s' '{base64_code}' | base64 -d | node - {argv}"
