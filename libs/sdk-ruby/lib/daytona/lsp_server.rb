# frozen_string_literal: true

module Daytona
  class LspServer
    module Language
      ALL = [
        JAVASCRIPT = :javascript,
        PYTHON = :python,
        TYPESCRIPT = :typescript

      ].freeze
    end

    # Represents a zero-based position in a text document,
    # specified by line number and character offset.
    Position = Data.define(:line, :character)

    # @return [Symbol]
    attr_reader :language_id

    # @return [String]
    attr_reader :path_to_project

    # @return [DaytonaToolboxApiClient::LspApi]
    attr_reader :toolbox_api

    # @return [String]
    attr_reader :sandbox_id

    # @param language_id [Symbol]
    # @param path_to_project [String]
    # @param toolbox_api [DaytonaToolboxApiClient::LspApi]
    # @param sandbox_id [String]
    def initialize(language_id:, path_to_project:, toolbox_api:, sandbox_id:)
      @language_id = language_id
      @path_to_project = path_to_project
      @toolbox_api = toolbox_api
      @sandbox_id = sandbox_id
    end

    # Gets completion suggestions at a position in a file
    #
    # @param path [String]
    # @param position [Daytona::LspServer::Position]
    # @return [DaytonaApiClient::CompletionList]
    def completions(path:, position:)
      toolbox_api.lsp_completions(
        DaytonaToolboxApiClient::LspCompletionParams.new(
          language_id:,
          path_to_project:,
          uri: uri(path),
          position: DaytonaApiClient::Position.new(line: position.line, character: position.character)
        )
      )
    end

    # Notify the language server that a file has been closed.
    # This method should be called when a file is closed in the editor to allow
    # the language server to clean up any resources associated with that file.
    #
    # @param path [String]
    # @return [void]
    def did_close(path)
      toolbox_api.lsp_did_close(
        DaytonaToolboxApiClient::LspDocumentRequest.new(language_id:, path_to_project:, uri: uri(path))
      )
    end

    # Notifies the language server that a file has been opened.
    # This method should be called when a file is opened in the editor to enable
    # language features like diagnostics and completions for that file. The server
    # will begin tracking the file's contents and providing language features.
    #
    # @param path [String]
    # @return [void]
    def did_open(path)
      toolbox_api.lsp_did_open(
        DaytonaToolboxApiClient::LspDocumentRequest.new(language_id:, path_to_project:, uri: uri(path))
      )
    end

    # Gets symbol information (functions, classes, variables, etc.) from a document.
    #
    # @param path [String]
    # @return [Array<DaytonaToolboxApiClient::LspSymbol]
    def document_symbols(path) = toolbox_api.lsp_document_symbols(language_id, path_to_project, uri(path))

    # Searches for symbols matching the query string across all files
    # in the Sandbox.
    #
    # @param query [String]
    # @return [Array<DaytonaToolboxApiClient::LspSymbol]
    def sandbox_symbols(query) = toolbox_api.lsp_workspace_symbols(language_id, path_to_project, query)

    # Starts the language server.
    # This method must be called before using any other LSP functionality.
    # It initializes the language server for the specified language and project.
    #
    # @return [void]
    def start
      toolbox_api.lsp_start(
        DaytonaToolboxApiClient::LspServerRequest.new(language_id:, path_to_project:)
      )
    end

    # Stops the language server.
    # This method should be called when the LSP server is no longer needed to
    # free up system resources.
    #
    # @return [void]
    def stop
      toolbox_api.lsp_stop(
        DaytonaToolboxApiClient::LspServerRequest.new(language_id:, path_to_project:)
      )
    end

    private

    # Convert path to file uri.
    #
    # @param path [String]
    # @return [String]
    def uri(path) = "file://#{path}"
  end
end
