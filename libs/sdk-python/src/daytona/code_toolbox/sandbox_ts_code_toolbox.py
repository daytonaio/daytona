# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import base64
from typing import Optional

from ..common.process import CodeRunParams


class SandboxTsCodeToolbox:
    def get_run_command(self, code: str, params: Optional[CodeRunParams] = None) -> str:
        # Encode the provided code in base64
        base64_code = base64.b64encode(code.encode()).decode()

        # Build command-line arguments string
        argv = ""
        if params and params.argv:
            argv = " ".join(params.argv)

        # Combine everything into the final command for TypeScript
        return (
            f""" sh -c 'echo {base64_code} | base64 --decode | npx ts-node -O """
            f""""{{\\\"module\\\":\\\"CommonJS\\\"}}" -e "$(cat)" x {argv} 2>&1 | grep -vE """
            f""""npm notice"' """
        )
