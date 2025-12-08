# frozen_string_literal: true

require 'net/http'

module Daytona
  module Util
    def self.demux(line) # rubocop:disable Metrics/MethodLength
      stdout = ''.dup
      stderr = ''.dup

      until line.empty?
        buff = line.start_with?(STDOUT_PREFIX) ? stdout : stderr
        line = line[3..]

        end_index = [
          line.index(STDOUT_PREFIX),
          line.index(STDERR_PREFIX)
        ].compact.min || line.length
        data = line[...end_index]
        buff << data

        line = line[end_index..]
      end

      [stdout, stderr]
    end

    # @param uri [URI]
    # @param on_chunk [Proc]
    # @param headers [Hash<String, String>]
    # @return [Thread]
    def self.stream_async(uri:, on_chunk:, headers: nil) # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
      Sdk.logger.debug("Starting async stream: #{uri}")
      Thread.new do
        Net::HTTP.start(uri.host, uri.port, use_ssl: uri.scheme == 'https') do |http|
          request = Net::HTTP::Get.new(uri, headers)

          http.request(request) do |response|
            response.read_body do |chunk|
              Sdk.logger.debug("Chunked response received: #{chunk.inspect}")
              on_chunk.call(chunk)
            end
          end
        end
      rescue Net::ReadTimeout => e
        Sdk.logger.debug("Async stream (#{uri}) timeout: #{e.inspect}")
      end
    end

    STDOUT_PREFIX = "\x01\x01\01"
    private_constant :STDOUT_PREFIX

    STDERR_PREFIX = "\x02\x02\02"
    private_constant :STDERR_PREFIX
  end
end
