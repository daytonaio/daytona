/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { CompletionList, LspSymbol, ToolboxApi } from '@daytonaio/api-client'
import { prefixRelativePath } from './utils/Path'

/**
 * Supported language server types.
 */
export enum LspLanguageId {
  PYTHON = 'python',
  TYPESCRIPT = 'typescript',
  JAVASCRIPT = 'javascript',
}

/**
 * Represents a zero-based position within a text document,
 * specified by line number and character offset.
 *
 * @interface
 * @property {number} line - Zero-based line number in the document
 * @property {number} character - Zero-based character offset on the line
 *
 * @example
 * const position: Position = {
 *   line: 10,     // Line 11 (zero-based)
 *   character: 15  // Character 16 on the line (zero-based)
 * };
 */
export type Position = {
  /** Zero-based line number */
  line: number
  /** Zero-based character offset */
  character: number
}

/**
 * Provides Language Server Protocol functionality for code intelligence to provide
 * IDE-like features such as code completion, symbol search, and more.
 *
 * @property {LspLanguageId} languageId - The language server type (e.g., "typescript")
 * @property {string} pathToProject - Absolute path to the project root directory
 * @property {ToolboxApi} toolboxApi - API client for Sandbox operations
 * @property {SandboxInstance} instance - The Sandbox instance this server belongs to
 *
 * @class
 */
export class LspServer {
  constructor(
    private readonly languageId: LspLanguageId,
    private readonly pathToProject: string,
    private readonly toolboxApi: ToolboxApi,
    private readonly sandboxId: string,
  ) {
    if (!Object.values(LspLanguageId).includes(this.languageId)) {
      throw new Error(
        `Invalid languageId: ${this.languageId}. Supported values are: ${Object.values(LspLanguageId).join(', ')}`,
      )
    }
  }

  /**
   * Starts the language server, must be called before using any other LSP functionality.
   * It initializes the language server for the specified language and project.
   *
   * @returns {Promise<void>}
   *
   * @example
   * const lsp = await sandbox.createLspServer('typescript', 'workspace/project');
   * await lsp.start();  // Initialize the server
   * // Now ready for LSP operations
   */
  public async start(): Promise<void> {
    await this.toolboxApi.lspStart(this.sandboxId, {
      languageId: this.languageId,
      pathToProject: this.pathToProject,
    })
  }

  /**
   * Stops the language server, should be called when the LSP server is no longer needed to
   * free up system resources.
   *
   * @returns {Promise<void>}
   *
   * @example
   * // When done with LSP features
   * await lsp.stop();  // Clean up resources
   */
  public async stop(): Promise<void> {
    await this.toolboxApi.lspStop(this.sandboxId, {
      languageId: this.languageId,
      pathToProject: this.pathToProject,
    })
  }

  /**
   * Notifies the language server that a file has been opened, enabling
   * language features like diagnostics and completions for that file. The server
   * will begin tracking the file's contents and providing language features.
   *
   * @param {string} path - Path to the opened file. Relative paths are resolved based on the user's
   * root directory.
   * @returns {Promise<void>}
   *
   * @example
   * // When opening a file for editing
   * await lsp.didOpen('workspace/project/src/index.ts');
   * // Now can get completions, symbols, etc. for this file
   */
  public async didOpen(path: string): Promise<void> {
    await this.toolboxApi.lspDidOpen(this.sandboxId, {
      languageId: this.languageId,
      pathToProject: this.pathToProject,
      uri: 'file://' + prefixRelativePath(this.pathToProject, path),
    })
  }

  /**
   * Notifies the language server that a file has been closed, should be called when a file is closed
   * in the editor to allow the language server to clean up any resources associated with that file.
   *
   * @param {string} path - Path to the closed file. Relative paths are resolved based on the project path
   * set in the LSP server constructor.
   * @returns {Promise<void>}
   *
   * @example
   * // When done editing a file
   * await lsp.didClose('workspace/project/src/index.ts');
   */
  public async didClose(path: string): Promise<void> {
    await this.toolboxApi.lspDidClose(this.sandboxId, {
      languageId: this.languageId,
      pathToProject: this.pathToProject,
      uri: 'file://' + prefixRelativePath(this.pathToProject, path),
    })
  }

  /**
   * Get symbol information (functions, classes, variables, etc.) from a document.
   *
   * @param {string} path - Path to the file to get symbols from. Relative paths are resolved based on the project path
   * set in the LSP server constructor.
   * @returns {Promise<LspSymbol[]>} List of symbols in the document. Each symbol includes:
   *                                 - name: The symbol's name
   *                                 - kind: The symbol's kind (function, class, variable, etc.)
   *                                 - location: The location of the symbol in the file
   *
   * @example
   * // Get all symbols in a file
   * const symbols = await lsp.documentSymbols('workspace/project/src/index.ts');
   * symbols.forEach(symbol => {
   *   console.log(`${symbol.kind} ${symbol.name}: ${symbol.location}`);
   * });
   */
  public async documentSymbols(path: string): Promise<LspSymbol[]> {
    const response = await this.toolboxApi.lspDocumentSymbols(
      this.sandboxId,
      this.languageId,
      this.pathToProject,
      'file://' + prefixRelativePath(this.pathToProject, path),
    )
    return response.data
  }

  /**
   * Searches for symbols matching the query string across the entire Sandbox.
   *
   * @param {string} query - Search query to match against symbol names
   * @returns {Promise<LspSymbol[]>} List of matching symbols from all files. Each symbol includes:
   *                                 - name: The symbol's name
   *                                 - kind: The symbol's kind (function, class, variable, etc.)
   *                                 - location: The location of the symbol in the file
   *
   * @deprecated Use `sandboxSymbols` instead. This method will be removed in a future version.
   */
  public async workspaceSymbols(query: string): Promise<LspSymbol[]> {
    return await this.sandboxSymbols(query)
  }

  /**
   * Searches for symbols matching the query string across the entire Sandbox.
   *
   * @param {string} query - Search query to match against symbol names
   * @returns {Promise<LspSymbol[]>} List of matching symbols from all files. Each symbol includes:
   *                                 - name: The symbol's name
   *                                 - kind: The symbol's kind (function, class, variable, etc.)
   *                                 - location: The location of the symbol in the file
   *
   * @example
   * // Search for all symbols containing "User"
   * const symbols = await lsp.sandboxSymbols('User');
   * symbols.forEach(symbol => {
   *   console.log(`${symbol.name} (${symbol.kind}) in ${symbol.location}`);
   * });
   */
  public async sandboxSymbols(query: string): Promise<LspSymbol[]> {
    const response = await this.toolboxApi.lspWorkspaceSymbols(
      this.sandboxId,
      this.languageId,
      this.pathToProject,
      query,
    )
    return response.data
  }

  /**
   * Gets completion suggestions at a position in a file.
   *
   * @param {string} path - Path to the file. Relative paths are resolved based on the project path
   * set in the LSP server constructor.
   * @param {Position} position - The position in the file where completion was requested
   * @returns {Promise<CompletionList>} List of completion suggestions. The list includes:
   *                                    - isIncomplete: Whether more items might be available
   *                                    - items: List of completion items, each containing:
   *                                      - label: The text to insert
   *                                      - kind: The kind of completion
   *                                      - detail: Additional details about the item
   *                                      - documentation: Documentation for the item
   *                                      - sortText: Text used to sort the item in the list
   *                                      - filterText: Text used to filter the item
   *                                      - insertText: The actual text to insert (if different from label)
   *
   * @example
   * // Get completions at a specific position
   * const completions = await lsp.completions('workspace/project/src/index.ts', {
   *   line: 10,
   *   character: 15
   * });
   * completions.items.forEach(item => {
   *   console.log(`${item.label} (${item.kind}): ${item.detail}`);
   * });
   */
  public async completions(path: string, position: Position): Promise<CompletionList> {
    const response = await this.toolboxApi.lspCompletions(this.sandboxId, {
      languageId: this.languageId,
      pathToProject: this.pathToProject,
      uri: 'file://' + prefixRelativePath(this.pathToProject, path),
      position,
    })
    return response.data
  }
}
