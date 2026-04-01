# frozen_string_literal: true

require 'json'
require 'logger'

require 'daytona_api_client'
require 'daytona_toolbox_api_client'
require 'toml'
require 'websocket-client-simple'

require_relative 'sdk/version'
require_relative 'config'
require_relative 'otel'
require_relative 'common/charts'
require_relative 'common/code_interpreter'
require_relative 'common/code_language'
require_relative 'common/daytona'
require_relative 'common/file_system'
require_relative 'common/image'
require_relative 'common/git'
require_relative 'common/process'
require_relative 'common/pty'
require_relative 'common/resources'
require_relative 'common/response'
require_relative 'common/snapshot'
require_relative 'code_interpreter'
require_relative 'computer_use'
require_relative 'code_toolbox/sandbox_python_code_toolbox'
require_relative 'code_toolbox/sandbox_ts_code_toolbox'
require_relative 'code_toolbox/sandbox_js_code_toolbox'
require_relative 'daytona'
require_relative 'file_system'
require_relative 'git'
require_relative 'lsp_server'
require_relative 'object_storage'
require_relative 'sandbox'
require_relative 'snapshot_service'
require_relative 'util'
require_relative 'volume'
require_relative 'volume_service'
require_relative 'process'

module Daytona
  module Sdk
    # Base error for all Daytona SDK errors.
    #
    # @attr_reader [Integer, nil] status_code HTTP status code if available
    class Error < StandardError
      attr_reader :status_code

      def initialize(message = nil, status_code: nil)
        super(message)
        @status_code = status_code
      end
    end

    # Raised when a request is malformed or contains invalid parameters (HTTP 400).
    #
    # @example
    #   rescue Daytona::Sdk::BadRequestError => e
    #     puts "Invalid parameters: #{e.message}"
    class BadRequestError < Error; end

    # Raised when API credentials are missing or invalid (HTTP 401).
    #
    # @example
    #   rescue Daytona::Sdk::AuthenticationError
    #     puts "Invalid or missing API key"
    class AuthenticationError < Error; end

    # Raised when the authenticated user lacks permission (HTTP 403).
    #
    # @example
    #   rescue Daytona::Sdk::ForbiddenError
    #     puts "Not authorized for this operation"
    class ForbiddenError < Error; end

    # Raised when a requested resource is not found (HTTP 404).
    #
    # @example
    #   rescue Daytona::Sdk::NotFoundError
    #     puts "Resource does not exist"
    class NotFoundError < Error; end

    # Raised when an operation conflicts with existing state (HTTP 409).
    #
    # @example
    #   rescue Daytona::Sdk::ConflictError
    #     puts "A resource with that name already exists"
    class ConflictError < Error; end

    # Raised for semantic validation failures (HTTP 422).
    #
    # @example
    #   rescue Daytona::Sdk::ValidationError => e
    #     puts "Validation failed: #{e.message}"
    class ValidationError < Error; end

    # Raised when the rate limit is exceeded (HTTP 429).
    #
    # @example
    #   rescue Daytona::Sdk::RateLimitError
    #     puts "Rate limit exceeded, back off and retry"
    class RateLimitError < Error; end

    # Raised for unexpected server-side failures (HTTP 5xx).
    #
    # @example
    #   rescue Daytona::Sdk::ServerError
    #     puts "Server error, retry later"
    class ServerError < Error; end

    # Raised when a polling operation exceeds the configured timeout.
    class TimeoutError < Error; end

    # Raised when the SDK cannot reach the Daytona API due to network issues.
    #
    # @example
    #   rescue Daytona::Sdk::ConnectionError
    #     puts "Cannot reach Daytona API, check connectivity"
    class ConnectionError < Error; end

    # Convert a DaytonaApiClient::ApiError or DaytonaToolboxApiClient::ApiError
    # to the appropriate Daytona::Sdk error subclass.
    #
    # @param api_error [DaytonaApiClient::ApiError, DaytonaToolboxApiClient::ApiError]
    # @param message_prefix [String] optional prefix for the error message
    # @return [Daytona::Sdk::Error] the mapped SDK error
    def self.map_api_error(api_error, message_prefix: '')
      code = api_error.respond_to?(:code) ? api_error.code : nil
      raw = api_error.message.to_s

      # Try to extract a cleaner message from a JSON response body
      message = begin
        data = JSON.parse(raw)
        data['message'] || data['error'] || raw
      rescue StandardError
        raw
      end

      full_message = message_prefix.empty? ? message : "#{message_prefix}#{message}"

      klass = case code
              when 400 then BadRequestError
              when 401 then AuthenticationError
              when 403 then ForbiddenError
              when 404 then NotFoundError
              when 409 then ConflictError
              when 422 then ValidationError
              when 429 then RateLimitError
              when (500..) then ServerError
              else Error
              end

      klass.new(full_message, status_code: code)
    end

    def self.logger = @logger ||= Logger.new($stdout, level: Logger::INFO)
  end
end
