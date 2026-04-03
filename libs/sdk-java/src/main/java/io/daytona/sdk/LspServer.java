// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.toolbox.client.api.LspApi;
import io.daytona.toolbox.client.model.CompletionList;
import io.daytona.toolbox.client.model.LspCompletionParams;
import io.daytona.toolbox.client.model.LspDocumentRequest;
import io.daytona.toolbox.client.model.LspPosition;
import io.daytona.toolbox.client.model.LspServerRequest;
import io.daytona.toolbox.client.model.LspSymbol;

import java.util.List;

/**
 * Language Server Protocol (LSP) interface for Sandbox code intelligence operations.
 *
 * <p>This class wraps Toolbox {@link LspApi} operations and maps Toolbox API errors to Daytona SDK
 * exceptions via {@link ExceptionMapper}. It supports starting/stopping language servers,
 * notifying document open/close events, and retrieving completions and symbols.
 */
public class LspServer {
    private final LspApi lspApi;

    /**
     * Supported language IDs for LSP operations.
     *
     * <p>Values mirror the TypeScript SDK {@code LspLanguageId} enum.
     */
    public enum LspLanguageId {
        PYTHON("python"),
        TYPESCRIPT("typescript"),
        JAVASCRIPT("javascript");

        private final String value;

        LspLanguageId(String value) {
            this.value = value;
        }

        /**
         * Returns the wire value used by Toolbox API requests.
         *
         * @return language ID string (for example {@code "python"})
         */
        public String getValue() {
            return value;
        }
    }

    /**
     * Creates an LSP server wrapper using the Toolbox LSP API client.
     *
     * @param lspApi Toolbox LSP API client
     */
    public LspServer(LspApi lspApi) {
        this.lspApi = lspApi;
    }

    /**
     * Starts a language server for the specified language and project root.
     *
     * <p>This must be called before document notifications or code intelligence requests.
     *
     * @param languageId language identifier (for example {@code "python"}, {@code "typescript"})
     * @param pathToProject absolute or relative project root path inside the sandbox
     */
    public void start(String languageId, String pathToProject) {
        ExceptionMapper.runToolbox(() -> lspApi.start(new LspServerRequest()
                .languageId(languageId)
                .pathToProject(pathToProject)));
    }

    /**
     * Stops a language server for the specified language and project root.
     *
     * @param languageId language identifier
     * @param pathToProject absolute or relative project root path inside the sandbox
     */
    public void stop(String languageId, String pathToProject) {
        ExceptionMapper.runToolbox(() -> lspApi.stop(new LspServerRequest()
                .languageId(languageId)
                .pathToProject(pathToProject)));
    }

    /**
     * Notifies the language server that a document has been opened.
     *
     * <p>Use this before requesting completions or symbols for a document to ensure the language
     * server tracks it.
     *
     * @param languageId language identifier
     * @param pathToProject absolute or relative project root path inside the sandbox
     * @param uri document URI (typically {@code file://...})
     */
    public void didOpen(String languageId, String pathToProject, String uri) {
        ExceptionMapper.runToolbox(() -> lspApi.didOpen(new LspDocumentRequest()
                .languageId(languageId)
                .pathToProject(pathToProject)
                .uri(uri)));
    }

    /**
     * Notifies the language server that a document has been closed.
     *
     * @param languageId language identifier
     * @param pathToProject absolute or relative project root path inside the sandbox
     * @param uri document URI (typically {@code file://...})
     */
    public void didClose(String languageId, String pathToProject, String uri) {
        ExceptionMapper.runToolbox(() -> lspApi.didClose(new LspDocumentRequest()
                .languageId(languageId)
                .pathToProject(pathToProject)
                .uri(uri)));
    }

    /**
     * Retrieves completion candidates at a zero-based position in a document.
     *
     * @param languageId language identifier
     * @param pathToProject absolute or relative project root path inside the sandbox
     * @param uri document URI (typically {@code file://...})
     * @param line zero-based line number
     * @param character zero-based character offset on the line
     * @return completion list returned by the language server
     */
    public CompletionList completions(String languageId, String pathToProject, String uri, int line, int character) {
        LspCompletionParams params = new LspCompletionParams()
                .languageId(languageId)
                .pathToProject(pathToProject)
                .uri(uri)
                .position(new LspPosition().line(line).character(character));

        return ExceptionMapper.callToolbox(() -> lspApi.completions(params));
    }

    /**
     * Returns all symbols defined in the specified document.
     *
     * @param languageId language identifier
     * @param pathToProject absolute or relative project root path inside the sandbox
     * @param uri document URI (typically {@code file://...})
     * @return list of document symbols
     */
    public List<LspSymbol> documentSymbols(String languageId, String pathToProject, String uri) {
        return ExceptionMapper.callToolbox(() -> lspApi.documentSymbols(languageId, pathToProject, uri));
    }

    /**
     * Searches workspace-wide symbols matching the provided query.
     *
     * @param query symbol query text
     * @param languageId language identifier
     * @param pathToProject absolute or relative project root path inside the sandbox
     * @return list of matching workspace symbols
     */
    public List<LspSymbol> workspaceSymbols(String query, String languageId, String pathToProject) {
        return ExceptionMapper.callToolbox(() -> lspApi.workspaceSymbols(query, languageId, pathToProject));
    }
}
