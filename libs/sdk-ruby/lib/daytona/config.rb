# frozen_string_literal: true

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
      # Load environment variables from .env and .env.local files
      # Files are loaded from the current working directory (where the code is executed)
      load_env_files

      @api_key = api_key || ENV.fetch('DAYTONA_API_KEY', nil)
      @jwt_token = jwt_token || ENV.fetch('DAYTONA_JWT_TOKEN', nil)
      @api_url = api_url || ENV.fetch('DAYTONA_API_URL', API_URL)
      @target = target || ENV.fetch('DAYTONA_TARGET', nil)
      @organization_id = organization_id || ENV.fetch('DAYTONA_ORGANIZATION_ID', nil)
      @_experimental = _experimental
    end

    private

    # Load only Daytona-specific environment variables from .env and .env.local files
    # Only loads variables that are not already set in the runtime environment
    # .env.local overrides .env
    # Files are loaded from the current working directory
    def load_env_files
      # Daytona-specific variables we want to load
      daytona_vars = %w[
        DAYTONA_API_KEY
        DAYTONA_API_URL
        DAYTONA_TARGET
        DAYTONA_JWT_TOKEN
        DAYTONA_ORGANIZATION_ID
      ]

      env_file = File.join(Dir.pwd, '.env')
      env_local_file = File.join(Dir.pwd, '.env.local')

      # Parse .env files using dotenv (doesn't set ENV automatically)
      env_from_file = {}
      env_from_file.merge!(Dotenv.parse(env_file)) if File.exist?(env_file)
      env_from_file.merge!(Dotenv.parse(env_local_file)) if File.exist?(env_local_file)

      # Only set Daytona-specific variables that aren't already in runtime
      daytona_vars.each do |var|
        ENV[var] = env_from_file[var] if env_from_file.key?(var) && !ENV.key?(var)
      end
    end
  end
end
