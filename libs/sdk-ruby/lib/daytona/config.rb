# frozen_string_literal: true

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

    # Initializes a new Daytona::Config object.
    #
    # @param api_key [String, nil] Daytona API key. Defaults to ENV['DAYTONA_API_KEY'].
    # @param jwt_token [String, nil] Daytona JWT token. Defaults to ENV['DAYTONA_JWT_TOKEN'].
    # @param api_url [String, nil] Daytona API URL. Defaults to ENV['DAYTONA_API_URL'] or Daytona::Config::API_URL.
    # @param organization_id [String, nil] Daytona organization ID. Defaults to ENV['DAYTONA_ORGANIZATION_ID'].
    # @param target [String, nil] Daytona target. Defaults to ENV['DAYTONA_TARGET'].
    def initialize(
      api_key: ENV.fetch('DAYTONA_API_KEY', nil),
      jwt_token: ENV.fetch('DAYTONA_JWT_TOKEN', nil),
      api_url: ENV.fetch('DAYTONA_API_URL', API_URL),
      organization_id: ENV.fetch('DAYTONA_ORGANIZATION_ID', nil),
      target: ENV.fetch('DAYTONA_TARGET', nil)
    )
      @api_key = api_key
      @jwt_token = jwt_token
      @api_url = api_url
      @target = target
      @organization_id = organization_id
    end
  end
end
