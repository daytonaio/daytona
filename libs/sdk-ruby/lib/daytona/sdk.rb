# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

require 'json'
require 'logger'

require 'daytona_api_client'
require 'daytona_toolbox_api_client'
require 'toml'
require 'websocket-client-simple'

require_relative 'sdk/version'
require_relative 'sdk/errors'
require_relative 'sdk/file_download_patch'
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
    # The error hierarchy and translation helpers live in `sdk/errors.rb`.
    # This file just provides cross-cutting bits that need to be loaded
    # alongside them.

    def self.logger = @logger ||= Logger.new($stdout, level: Logger::INFO)
  end
end
