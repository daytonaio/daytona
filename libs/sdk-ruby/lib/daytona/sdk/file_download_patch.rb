# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

# Patches the OpenAPI-generated `download_file` and `call_api` helpers in both
# API clients so that streaming responses (return_type == 'File') do not lose
# the response body when the server returns an error status.
#
# Background:
#   For File downloads the generated `download_file` attaches Typhoeus
#   `on_body` callbacks to stream the response into a Tempfile. The presence
#   of `on_body` prevents Typhoeus from populating `response.body`, which
#   means the daemon's structured error envelope (code/source/message) is
#   discarded when the response is an error. The generated `call_api` then
#   raises `ApiError` with an empty `response_body`.
#
# Fix:
#   * In `download_file`, defer tempfile creation until `on_headers` and
#     inspect the status code. For 2xx responses, behave exactly like the
#     upstream code. For non-2xx responses, accumulate the body into a
#     buffer stored as an instance variable on the ApiClient.
#   * In `call_api`, rescue the `ApiError`, and if our buffer is populated,
#     re-raise with `response_body` set to the buffered envelope.

require 'tempfile'

module Daytona
  module Sdk
    module FileDownloadPatch
      def self.apply!(api_client_class, api_error_class)
        api_client_class.class_eval do
          define_method(:download_file) do |request, &block|
            tempfile = nil
            encoding = nil
            stream_to_tempfile = false
            error_body = String.new.b
            @_daytona_error_body = nil

            request.on_headers do |response|
              stream_to_tempfile = response.code && response.code >= 200 && response.code < 300
              next unless stream_to_tempfile

              content_disposition = response.headers['Content-Disposition']
              if content_disposition && content_disposition =~ /filename=/i
                filename = content_disposition[/filename=['"]?([^'"\s]+)['"]?/, 1]
                prefix = sanitize_filename(filename)
              else
                prefix = 'download-'
              end
              prefix += '-' unless prefix.end_with?('-')
              encoding = response.body.encoding
              tempfile = Tempfile.open(prefix, @config.temp_folder_path, encoding: encoding)
            end

            request.on_body do |chunk|
              if stream_to_tempfile
                chunk.force_encoding(encoding)
                tempfile.write(chunk)
              else
                error_body << chunk.b
              end
            end

            request.on_complete do
              if stream_to_tempfile
                if tempfile.nil?
                  raise api_error_class.new(
                    "Failed to create the tempfile based on the HTTP response from the server: #{request.inspect}"
                  )
                end
                tempfile.close
                @config.logger.info(
                  "Temp file written to #{tempfile.path}, please copy the file to a proper folder " \
                  "with e.g. `FileUtils.cp(tempfile.path, '/new/file/path')` otherwise the temp file " \
                  'will be deleted automatically with GC. It\'s also recommended to delete the temp file ' \
                  'explicitly with `tempfile.delete`'
                )
                block&.call(tempfile)
              else
                @_daytona_error_body = error_body unless error_body.empty?
              end
            end
          end

          alias_method :_daytona_orig_call_api, :call_api

          define_method(:call_api) do |http_method, path, opts = {}|
            _daytona_orig_call_api(http_method, path, opts)
          rescue api_error_class => e
            if opts[:return_type] == 'File' && @_daytona_error_body && !@_daytona_error_body.empty?
              new_err = api_error_class.new(
                code: e.code,
                response_headers: e.response_headers,
                response_body: @_daytona_error_body
              )
              @_daytona_error_body = nil
              raise new_err, e.message
            end
            raise
          ensure
            @_daytona_error_body = nil
          end
        end
      end
    end
  end
end

if Object.const_defined?(:DaytonaApiClient) &&
   DaytonaApiClient.const_defined?(:ApiClient) &&
   DaytonaApiClient.const_defined?(:ApiError)
  Daytona::Sdk::FileDownloadPatch.apply!(DaytonaApiClient::ApiClient, DaytonaApiClient::ApiError)
end

if Object.const_defined?(:DaytonaToolboxApiClient) &&
   DaytonaToolboxApiClient.const_defined?(:ApiClient) &&
   DaytonaToolboxApiClient.const_defined?(:ApiError)
  Daytona::Sdk::FileDownloadPatch.apply!(
    DaytonaToolboxApiClient::ApiClient, DaytonaToolboxApiClient::ApiError
  )
end
