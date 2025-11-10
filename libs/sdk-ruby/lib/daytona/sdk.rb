# frozen_string_literal: true

require 'logger'

require 'dotenv'
Dotenv.load('.env.local', '.env')
require 'daytona_api_client'
require 'toml'
require 'websocket-client-simple'

require_relative 'sdk/version'
require_relative 'config'
require_relative 'common/charts'
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
require_relative 'computer_use'
require_relative 'code_toolbox/sandbox_python_code_toolbox'
require_relative 'code_toolbox/sandbox_ts_code_toolbox'
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
    class Error < StandardError; end

    def self.logger = @logger ||= Logger.new($stdout, level: Logger::INFO)
  end
end
