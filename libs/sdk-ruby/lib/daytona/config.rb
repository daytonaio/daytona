# frozen_string_literal: true

# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

require 'dotenv'

module Daytona
  class Config
    API_URL = 'https://app.daytona.io/api'

    # API key for authentication with the Daytona API
    #
    # @return [String, nil] Daytona API key
    attr_accessor :api_key

    # JWT token for authentication with the Daytona API
    #
    # @return [String, nil] Daytona JWT token
    attr_accessor :jwt_token

    # URL of the Daytona API
    #
    # @return [String, nil] Daytona API URL
    attr_accessor :api_url

    # Organization ID for authentication with the Daytona API
    #
    # @return [String, nil] Daytona API URL
    attr_accessor :organization_id

    # Target environment for sandboxes
    #
    # @return [String, nil] Daytona target
    attr_accessor :target

    # Experimental configuration options
    #
    # @return [Hash, nil] Experimental configuration hash
    attr_accessor :_experimental

    # Initializes a new Daytona::Config object.
    #
    # @param api_key [String, nil] Daytona API key. Defaults to ENV['DAYTONA_API_KEY'].
    # @param jwt_token [String, nil] Daytona JWT token. Defaults to ENV['DAYTONA_JWT_TOKEN'].
    # @param api_url [String, nil] Daytona API URL. Defaults to ENV['DAYTONA_API_URL'] or Daytona::Config::API_URL.
    # @param organization_id [String, nil] Daytona organization ID. Defaults to ENV['DAYTONA_ORGANIZATION_ID'].
    # @param target [String, nil] Daytona target. Defaults to ENV['DAYTONA_TARGET'].
    # @param _experimental [Hash, nil] Experimental configuration options.
    def initialize( # rubocop:disable Metrics/ParameterLists
      api_key: nil,
      jwt_token: nil,
      api_url: nil,
      organization_id: nil,
      target: nil,
      _experimental: nil
    )
      @env_reader = daytona_env_reader

      @api_key = api_key || @env_reader.call('DAYTONA_API_KEY')
      @jwt_token = jwt_token || @env_reader.call('DAYTONA_JWT_TOKEN')
      @api_url = api_url || @env_reader.call('DAYTONA_API_URL') || API_URL
      @target = target || @env_reader.call('DAYTONA_TARGET')
      @organization_id = organization_id || @env_reader.call('DAYTONA_ORGANIZATION_ID')
      @_experimental = _experimental
    end

    # Reads a DAYTONA_-prefixed environment variable using the same precedence
    # as the Config initializer: runtime ENV first, then .env.local, then .env.
    # Only names starting with DAYTONA_ are accepted.
    #
    # @param name [String] The environment variable name. Must start with DAYTONA_.
    # @return [String, nil] The value of the environment variable, or nil if not set.
    # @raise [ArgumentError] If name does not start with DAYTONA_.
    def read_env(name)
      @env_reader.call(name)
    end

    private

    # Returns a lambda that looks up DAYTONA_-prefixed env vars without writing to ENV.
    # Files are parsed once; lookups check runtime env first, then .env.local, then .env.
    def daytona_env_reader
      file_vars = {}
      env_file = File.join(Dir.pwd, '.env')
      file_vars.merge!(daytona_filter(Dotenv.parse(env_file))) if File.exist?(env_file)
      env_local_file = File.join(Dir.pwd, '.env.local')
      file_vars.merge!(daytona_filter(Dotenv.parse(env_local_file))) if File.exist?(env_local_file)

      lambda do |name|
        raise ArgumentError, "Variable must start with 'DAYTONA_', got '#{name}'" unless name.start_with?('DAYTONA_')

        ENV.key?(name) ? ENV[name] : file_vars[name]
      end
    end

    def daytona_filter(env_hash)
      env_hash.select { |k, _| k.start_with?('DAYTONA_') }
    end
  end
end
