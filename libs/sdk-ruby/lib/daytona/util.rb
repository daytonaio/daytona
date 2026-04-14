# frozen_string_literal: true

# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

require 'net/http'

module Daytona
  module Util
    STDOUT_PREFIX = "\x01\x01\01"
    private_constant :STDOUT_PREFIX

    STDERR_PREFIX = "\x02\x02\02"
    private_constant :STDERR_PREFIX

    PREFIX_LEN = STDOUT_PREFIX.bytesize
    private_constant :PREFIX_LEN

    def self.demux(line)
      out_parts = []
      err_parts = []
      state = nil
      pos = 0

      while pos < line.bytesize
        si = line.index(STDOUT_PREFIX, pos)
        ei = line.index(STDERR_PREFIX, pos)

        if si && (ei.nil? || si < ei)
          next_idx = si
          next_state = :stdout
        elsif ei
          next_idx = ei
          next_state = :stderr
        else
          case state
          when :stdout then out_parts << line[pos..]
          when :stderr then err_parts << line[pos..]
          end
          break
        end

        if pos < next_idx
          chunk = line[pos...next_idx]
          case state
          when :stdout then out_parts << chunk
          when :stderr then err_parts << chunk
          end
        end

        state = next_state
        pos = next_idx + PREFIX_LEN
      end

      [out_parts.join, err_parts.join]
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
  end
end
