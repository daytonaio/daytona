# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0


class GitCommitResponse:
    """Response from the git commit.

    Attributes:
        sha (str): The SHA of the commit
    """

    def __init__(self, sha: str):
        self.sha = sha
