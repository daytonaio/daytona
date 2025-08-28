# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from typing import List

from daytona_api_client_async import (
    CompletionList,
    LspCompletionParams,
    LspDocumentRequest,
    LspServerRequest,
    LspSymbol,
    ToolboxApi,
)
from deprecated import deprecated

from .._utils.errors import intercept_errors
from .._utils.path import prefix_relative_path
from ..common.lsp_server import LspLanguageId, Position


class AsyncLspServer:
    """Provides Language Server Protocol functionality for code intelligence to provide
    IDE-like features such as code completion, symbol search, and more.
    """

    def __init__(
        self,
        language_id: LspLanguageId,
        path_to_project: str,
        toolbox_api: ToolboxApi,
        sandbox_id: str,
    ):
        """Initializes a new LSP server instance.

        Args:
            language_id (LspLanguageId): The language server type (e.g., LspLanguageId.TYPESCRIPT).
            path_to_project (str): Absolute path to the project root directory.
            toolbox_api (ToolboxApi): API client for Sandbox operations.
            instance (SandboxInstance): The Sandbox instance this server belongs to.
        """
        self._language_id = str(language_id)
        self._path_to_project = path_to_project
        self._toolbox_api = toolbox_api
        self._sandbox_id = sandbox_id

    @intercept_errors(message_prefix="Failed to start LSP server: ")
    async def start(self) -> None:
        """Starts the language server.

        This method must be called before using any other LSP functionality.
        It initializes the language server for the specified language and project.

        Example:
            ```python
            lsp = sandbox.create_lsp_server("typescript", "workspace/project")
            await lsp.start()  # Initialize the server
            # Now ready for LSP operations
            ```
        """
        await self._toolbox_api.lsp_start(
            self._sandbox_id,
            lsp_server_request=LspServerRequest(
                language_id=self._language_id,
                path_to_project=self._path_to_project,
            ),
        )

    @intercept_errors(message_prefix="Failed to stop LSP server: ")
    async def stop(self) -> None:
        """Stops the language server.

        This method should be called when the LSP server is no longer needed to
        free up system resources.

        Example:
            ```python
            # When done with LSP features
            await lsp.stop()  # Clean up resources
            ```
        """
        await self._toolbox_api.lsp_stop(
            self._sandbox_id,
            lsp_server_request=LspServerRequest(
                language_id=self._language_id,
                path_to_project=self._path_to_project,
            ),
        )

    @intercept_errors(message_prefix="Failed to open file: ")
    async def did_open(self, path: str) -> None:
        """Notifies the language server that a file has been opened.

        This method should be called when a file is opened in the editor to enable
        language features like diagnostics and completions for that file. The server
        will begin tracking the file's contents and providing language features.

        Args:
            path (str): Path to the opened file. Relative paths are resolved based on the project path
            set in the LSP server constructor.

        Example:
            ```python
            # When opening a file for editing
            await lsp.did_open("workspace/project/src/index.ts")
            # Now can get completions, symbols, etc. for this file
            ```
        """
        path = prefix_relative_path(self._path_to_project, path)
        await self._toolbox_api.lsp_did_open(
            self._sandbox_id,
            lsp_document_request=LspDocumentRequest(
                language_id=self._language_id,
                path_to_project=self._path_to_project,
                uri=f"file://{path}",
            ),
        )

    @intercept_errors(message_prefix="Failed to close file: ")
    async def did_close(self, path: str) -> None:
        """Notify the language server that a file has been closed.

        This method should be called when a file is closed in the editor to allow
        the language server to clean up any resources associated with that file.

        Args:
            path (str): Path to the closed file. Relative paths are resolved based on the project path
            set in the LSP server constructor.

        Example:
            ```python
            # When done editing a file
            await lsp.did_close("workspace/project/src/index.ts")
            ```
        """
        await self._toolbox_api.lsp_did_close(
            self._sandbox_id,
            lsp_document_request=LspDocumentRequest(
                language_id=self._language_id,
                path_to_project=self._path_to_project,
                uri=f"file://{prefix_relative_path(self._path_to_project, path)}",
            ),
        )

    @intercept_errors(message_prefix="Failed to get symbols from document: ")
    async def document_symbols(self, path: str) -> List[LspSymbol]:
        """Gets symbol information (functions, classes, variables, etc.) from a document.

        Args:
            path (str): Path to the file to get symbols from. Relative paths are resolved based on the project path
            set in the LSP server constructor.

        Returns:
            List[LspSymbol]: List of symbols in the document. Each symbol includes:
                - name: The symbol's name
                - kind: The symbol's kind (function, class, variable, etc.)
                - location: The location of the symbol in the file

        Example:
            ```python
            # Get all symbols in a file
            symbols = await lsp.document_symbols("workspace/project/src/index.ts")
            for symbol in symbols:
                print(f"{symbol.kind} {symbol.name}: {symbol.location}")
            ```
        """
        return await self._toolbox_api.lsp_document_symbols(
            self._sandbox_id,
            language_id=self._language_id,
            path_to_project=self._path_to_project,
            uri=f"file://{prefix_relative_path(self._path_to_project, path)}",
        )

    @deprecated(
        reason="Method is deprecated. Use `sandbox_symbols` instead. This method will be removed in a future version."
    )
    async def workspace_symbols(self, query: str) -> List[LspSymbol]:
        """Searches for symbols matching the query string across all files
        in the Sandbox.

        Args:
            query (str): Search query to match against symbol names.

        Returns:
            List[LspSymbol]: List of matching symbols from all files.
        """
        return await self.sandbox_symbols(query)

    @intercept_errors(message_prefix="Failed to get symbols from sandbox: ")
    async def sandbox_symbols(self, query: str) -> List[LspSymbol]:
        """Searches for symbols matching the query string across all files
        in the Sandbox.

        Args:
            query (str): Search query to match against symbol names.

        Returns:
            List[LspSymbol]: List of matching symbols from all files. Each symbol
                includes:
                - name: The symbol's name
                - kind: The symbol's kind (function, class, variable, etc.)
                - location: The location of the symbol in the file

        Example:
            ```python
            # Search for all symbols containing "User"
            symbols = await lsp.sandbox_symbols("User")
            for symbol in symbols:
                print(f"{symbol.name} in {symbol.location}")
            ```
        """
        return await self._toolbox_api.lsp_workspace_symbols(
            self._sandbox_id,
            language_id=self._language_id,
            path_to_project=self._path_to_project,
            query=query,
        )

    @intercept_errors(message_prefix="Failed to get completions: ")
    async def completions(self, path: str, position: Position) -> CompletionList:
        """Gets completion suggestions at a position in a file.

        Args:
            path (str): Path to the file. Relative paths are resolved based on the project path
            set in the LSP server constructor.
            position (Position): Cursor position to get completions for.

        Returns:
            CompletionList: List of completion suggestions. The list includes:
                - isIncomplete: Whether more items might be available
                - items: List of completion items, each containing:
                    - label: The text to insert
                    - kind: The kind of completion
                    - detail: Additional details about the item
                    - documentation: Documentation for the item
                    - sortText: Text used to sort the item in the list
                    - filterText: Text used to filter the item
                    - insertText: The actual text to insert (if different from label)

        Example:
            ```python
            # Get completions at a specific position
            pos = Position(line=10, character=15)
            completions = await lsp.completions("workspace/project/src/index.ts", pos)
            for item in completions.items:
                print(f"{item.label} ({item.kind}): {item.detail}")
            ```
        """
        return await self._toolbox_api.lsp_completions(
            self._sandbox_id,
            lsp_completion_params=LspCompletionParams(
                language_id=self._language_id,
                path_to_project=self._path_to_project,
                uri=f"file://{prefix_relative_path(self._path_to_project, path)}",
                position=position,
            ),
        )
