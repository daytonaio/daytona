# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import warnings
from unittest.mock import MagicMock

from daytona.common.lsp_server import LspCompletionPosition, LspLanguageId


def _make_lsp(language_id: str | LspLanguageId = LspLanguageId.PYTHON, path_to_project: str = "/workspace/project"):
    from daytona._sync.lsp_server import LspServer

    api_client = MagicMock()
    return LspServer(language_id, path_to_project, api_client), api_client


class TestLspServer:
    def test_init_stores_language_and_project_path(self):
        lsp, _api_client = _make_lsp("typescript", "/workspace/ts")

        assert lsp._language_id == "typescript"
        assert lsp._path_to_project == "/workspace/ts"

    def test_start_builds_request(self):
        lsp, api_client = _make_lsp()

        lsp.start()

        request = api_client.start.call_args.kwargs["request"]
        assert request.language_id == "python"
        assert request.path_to_project == "/workspace/project"

    def test_stop_builds_request(self):
        lsp, api_client = _make_lsp()

        lsp.stop()

        request = api_client.stop.call_args.kwargs["request"]
        assert request.language_id == "python"
        assert request.path_to_project == "/workspace/project"

    def test_did_open_prefixes_uri(self):
        lsp, api_client = _make_lsp()

        lsp.did_open("/workspace/project/main.py")

        request = api_client.did_open.call_args.kwargs["request"]
        assert request.uri == "file:///workspace/project/main.py"

    def test_did_close_prefixes_uri(self):
        lsp, api_client = _make_lsp()

        lsp.did_close("relative/file.py")

        request = api_client.did_close.call_args.kwargs["request"]
        assert request.uri == "file://relative/file.py"

    def test_document_symbols_delegates(self):
        lsp, api_client = _make_lsp()
        api_client.document_symbols.return_value = [MagicMock(name="symbol")]

        result = lsp.document_symbols("app.py")

        assert len(result) == 1
        api_client.document_symbols.assert_called_once_with(
            language_id="python",
            path_to_project="/workspace/project",
            uri="file://app.py",
        )

    def test_sandbox_symbols_delegates(self):
        lsp, api_client = _make_lsp()
        api_client.workspace_symbols.return_value = [MagicMock(name="symbol")]

        result = lsp.sandbox_symbols("User")

        assert len(result) == 1
        api_client.workspace_symbols.assert_called_once_with(
            language_id="python",
            path_to_project="/workspace/project",
            query="User",
        )

    def test_workspace_symbols_warns_and_delegates(self):
        lsp, api_client = _make_lsp()
        api_client.workspace_symbols.return_value = []

        with warnings.catch_warnings(record=True) as caught:
            warnings.simplefilter("always")
            result = lsp.workspace_symbols("Thing")

        assert result == []
        assert any("deprecated" in str(w.message).lower() for w in caught)
        api_client.workspace_symbols.assert_called_once_with(
            language_id="python",
            path_to_project="/workspace/project",
            query="Thing",
        )

    def test_completions_builds_position_request(self):
        lsp, api_client = _make_lsp()
        api_client.completions.return_value = MagicMock(items=[MagicMock(label="print")])

        result = lsp.completions("main.py", LspCompletionPosition(line=10, character=15))

        assert len(result.items) == 1
        request = api_client.completions.call_args.kwargs["request"]
        assert request.uri == "file://main.py"
        assert request.position.line == 10
        assert request.position.character == 15
