# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

require 'json'

module Daytona
  module Sdk # rubocop:disable Metrics/ModuleLength
    # ApiError classes raised by the generated OpenAPI clients. Kept here so
    # the error module can resolve them without depending on `sdk.rb`.
    API_ERROR_CLASSES = [DaytonaApiClient::ApiError, DaytonaToolboxApiClient::ApiError].freeze

    # Wire-format `source` values set by the translation layer when a
    # Daytona service stamps them on the wire envelope. `nil` source means
    # the response did not carry a structured envelope (treat as opaque).
    SOURCE_API     = 'DAYTONA_API'
    SOURCE_DAEMON  = 'DAYTONA_DAEMON'
    SOURCE_PROXY   = 'DAYTONA_PROXY'

    # ---------------------------------------------------------------------
    # Base
    # ---------------------------------------------------------------------

    # Base class for every error raised by the Daytona SDK.
    #
    # @example Catching any SDK error
    #   begin
    #     daytona.create(params)
    #   rescue Daytona::Sdk::Error => e
    #     puts e.status_code, e.code, e.source
    #   end
    class Error < StandardError
      attr_reader :headers

      def initialize(message = nil, status_code: nil, code: nil, source: nil, headers: nil)
        super(message)
        @status_code = status_code
        @code = code
        @source = source
        @headers = headers || {}
      end

      def status_code = @status_code || metadata_from_cause[:status_code]

      def code = @code || metadata_from_cause[:code]

      # Returns the originating service, or `nil` if unknown. Falls back to
      # metadata carried by `cause` so a re-raised ApiError keeps its
      # server-stamped source.
      def source = @source || metadata_from_cause[:source]

      private

      def metadata_from_cause
        @metadata_from_cause ||= Sdk.api_error_details(cause)
      end
    end

    # ---------------------------------------------------------------------
    # HTTP status-class errors
    # ---------------------------------------------------------------------

    # HTTP 400 — request was rejected as malformed or invalid.
    class BadRequestError < Error; end

    # @deprecated Use {BadRequestError} instead. Kept as an alias so existing
    #   `rescue Daytona::Sdk::ValidationError` blocks keep working.
    ValidationError = BadRequestError

    # HTTP 401 — authentication failed or credentials missing.
    class AuthenticationError < Error; end

    # HTTP 403 — caller is authenticated but not allowed.
    class ForbiddenError < Error; end

    # HTTP 404 — target resource does not exist.
    class NotFoundError < Error; end

    # HTTP 408 / 504 / client-side timeouts.
    class TimeoutError < Error; end

    # HTTP 409 — request conflicts with current resource state.
    class ConflictError < Error; end

    # HTTP 410 — resource is permanently gone.
    class GoneError < Error; end

    # HTTP 422 — request is well-formed but semantically invalid.
    class UnprocessableEntityError < Error; end

    # HTTP 429 — rate limit was exceeded.
    class RateLimitError < Error; end

    # Generic 5xx fallback when a more specific class doesn't apply.
    class ServerError < Error; end

    # HTTP 500 — server-side bug or unhandled condition.
    class InternalServerError < ServerError; end

    # HTTP 502 — an upstream dependency rejected or dropped the request.
    class BadGatewayError < ServerError; end

    # HTTP 503 — the service is temporarily refusing traffic.
    class ServiceUnavailableError < ServerError; end

    # Transport-level failure (no HTTP response received).
    class ConnectionError < Error; end

    # Transport-level timeout. Subclass of ConnectionError so callers that
    # catch the broader connection-failure category also match.
    class ConnectionTimeoutError < ConnectionError; end

    # ---------------------------------------------------------------------
    # Domain subclasses — each inherits from the HTTP-status parent that
    # matches its server-side status code.
    # ---------------------------------------------------------------------

    # Daemon: git
    class GitAuthFailedError < AuthenticationError; end
    class GitRepoNotFoundError < NotFoundError; end
    class GitBranchNotFoundError < NotFoundError; end
    class GitBranchExistsError < ConflictError; end
    class GitPushRejectedError < ConflictError; end
    class GitDirtyWorktreeError < ConflictError; end
    class GitMergeConflictError < ConflictError; end

    # Daemon: filesystem
    class FileNotFoundError < NotFoundError; end
    class FileAccessDeniedError < ForbiddenError; end

    # Daemon: LSP
    class LspServerNotInitializedError < BadRequestError; end

    # Daemon: process / session
    class ProcessExecutionTimeoutError < TimeoutError; end
    class ProcessNotFoundError < NotFoundError; end
    class SessionEndedError < GoneError; end
    class CommandAlreadyCompletedError < GoneError; end

    # Daemon: computer-use
    class A11yUnavailableError < ServiceUnavailableError; end
    class RecordingStillActiveError < ConflictError; end
    class RecordingFfmpegNotFoundError < ServiceUnavailableError; end

    # ---------------------------------------------------------------------
    # Routing tables
    # ---------------------------------------------------------------------

    STATUS_CODE_TO_ERROR = {
      400 => BadRequestError,
      401 => AuthenticationError,
      403 => ForbiddenError,
      404 => NotFoundError,
      408 => TimeoutError,
      409 => ConflictError,
      410 => GoneError,
      422 => UnprocessableEntityError,
      429 => RateLimitError,
      500 => InternalServerError,
      502 => BadGatewayError,
      503 => ServiceUnavailableError,
      504 => TimeoutError
    }.freeze

    # (source, code) tuple → exception class. Resolved BEFORE the status
    # code, so a server-stamped domain code always wins over the generic
    # status class.
    CODE_TO_ERROR = {
      # Daemon: git
      [SOURCE_DAEMON, 'GIT_AUTH_FAILED'] => GitAuthFailedError,
      [SOURCE_DAEMON, 'GIT_REPO_NOT_FOUND'] => GitRepoNotFoundError,
      [SOURCE_DAEMON, 'GIT_BRANCH_NOT_FOUND'] => GitBranchNotFoundError,
      [SOURCE_DAEMON, 'GIT_BRANCH_EXISTS'] => GitBranchExistsError,
      [SOURCE_DAEMON, 'GIT_PUSH_REJECTED'] => GitPushRejectedError,
      [SOURCE_DAEMON, 'GIT_DIRTY_WORKTREE'] => GitDirtyWorktreeError,
      [SOURCE_DAEMON, 'GIT_MERGE_CONFLICT'] => GitMergeConflictError,

      # Daemon: filesystem
      [SOURCE_DAEMON, 'FILE_NOT_FOUND'] => FileNotFoundError,
      [SOURCE_DAEMON, 'FILE_ACCESS_DENIED'] => FileAccessDeniedError,

      # Daemon: LSP
      [SOURCE_DAEMON, 'LSP_SERVER_NOT_INITIALIZED'] => LspServerNotInitializedError,

      # Daemon: process / session
      [SOURCE_DAEMON, 'PROCESS_EXECUTION_TIMEOUT'] => ProcessExecutionTimeoutError,
      [SOURCE_DAEMON, 'PROCESS_NOT_FOUND'] => ProcessNotFoundError,
      [SOURCE_DAEMON, 'SESSION_ENDED'] => SessionEndedError,
      [SOURCE_DAEMON, 'COMMAND_ALREADY_COMPLETED'] => CommandAlreadyCompletedError,

      # Daemon: computer-use
      [SOURCE_DAEMON, 'A11Y_UNAVAILABLE'] => A11yUnavailableError,
      [SOURCE_DAEMON, 'RECORDING_STILL_ACTIVE'] => RecordingStillActiveError,
      [SOURCE_DAEMON, 'RECORDING_FFMPEG_NOT_FOUND'] => RecordingFfmpegNotFoundError
    }.freeze

    # ---------------------------------------------------------------------
    # Public translation helpers (module-level functions on Daytona::Sdk)
    # ---------------------------------------------------------------------

    # Translate an OpenAPI-client error into the most specific Daytona SDK
    # exception. Accepts an optional `prefix` that's prepended to the
    # message for context (e.g. "Failed to create sandbox"). When `error`
    # is already an `Sdk::Error` (e.g. raised by the streaming transfer
    # helpers for cancel/timeout), its class is preserved and only the
    # message is prefixed.
    def self.wrap_error(error, prefix = nil)
      if error.is_a?(Error)
        message = prefix ? "#{prefix}: #{error.message}" : error.message
        return error.class.new(message, status_code: error.status_code, code: error.code,
                                        source: error.source, headers: error.headers)
      end

      details = api_error_details(error)
      base_message = parsed_message(error) || error.message
      message = prefix ? "#{prefix}: #{base_message}" : base_message
      error_class_for(details).new(message, **details.slice(:status_code, :code, :source, :headers))
    end

    # Extract status code, code, source and headers from a raised OpenAPI
    # error. Returns an empty hash when the error is not one of the
    # generated client types.
    def self.api_error_details(error)
      return {} unless API_ERROR_CLASSES.any? { |c| error.is_a?(c) }

      data = parse_error_body(error.respond_to?(:response_body) ? error.response_body : nil)
      {
        status_code: error.respond_to?(:code) ? error.code : nil,
        code: data[:code],
        source: data[:source],
        headers: error.respond_to?(:response_headers) ? error.response_headers : nil
      }
    end

    # Choose the exception class for a parsed error: (source, code) match
    # wins, then HTTP status code, then the base Error.
    def self.error_class_for(details)
      code = details[:code]
      source = details[:source]
      if code && source
        cls = CODE_TO_ERROR[[source, code]]
        return cls if cls
      end
      STATUS_CODE_TO_ERROR.fetch(details[:status_code], Error)
    end

    def self.parse_error_body(response_body)
      return {} if response_body.nil? || response_body.empty?

      data = JSON.parse(response_body)
      return {} unless data.is_a?(Hash)

      {
        message: string_or_nil(data['message']) || string_or_nil(data['error']),
        code: string_or_nil(data['code'] || data['error_code']),
        source: string_or_nil(data['source'])
      }
    rescue JSON::ParserError
      {}
    end

    # @api private
    def self.parsed_message(error)
      return nil unless API_ERROR_CLASSES.any? { |c| error.is_a?(c) }
      return nil unless error.respond_to?(:response_body)

      parse_error_body(error.response_body)[:message]
    end

    def self.string_or_nil(value)
      value.is_a?(String) && !value.empty? ? value : nil
    end
  end
end
