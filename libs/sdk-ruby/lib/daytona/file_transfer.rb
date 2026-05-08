# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

require 'json'
require 'stringio'
require 'tempfile'
require 'typhoeus'

module Daytona
  # Progress information for a streaming download.
  DownloadProgress = Struct.new(:bytes_received, :total_bytes, keyword_init: true)

  # Progress information for a streaming upload.
  UploadProgress = Struct.new(:bytes_sent, keyword_init: true)

  class MultipartDownloadStreamParser
    attr_reader :error_message
    attr_reader :part_total_bytes
    attr_writer :boundary_token

    def initialize(&on_file_chunk)
      @on_file_chunk = on_file_chunk
      @boundary_token = nil
      @buffer = String.new.b
      @state = :preamble
      @part_name = nil
      @part_total_bytes = nil
      @error_buffer = String.new.b
    end

    def <<(chunk)
      @buffer << chunk.b
      process!
    end

    def finish!
      process!

      return if @state == :done || @buffer.empty?

      emit(@buffer)
      finalize_part!
      @buffer = String.new.b
      @state = :done
    end

    private

    def process!
      loop do
        advanced = case @state
                   when :preamble then consume_preamble?
                   when :headers then consume_headers?
                   when :body then consume_body?
                   else false
                   end

        break unless advanced
      end
    end

    def consume_preamble?
      start_marker = "#{boundary}\r\n".b
      index = @buffer.index(start_marker)
      return retain_tail?(start_marker.bytesize - 1) unless index

      @buffer = remaining_bytes(index + start_marker.bytesize)
      @state = :headers
      true
    end

    def consume_headers?
      index = @buffer.index("\r\n\r\n".b)
      return false unless index

      headers = @buffer.byteslice(0, index)
      @buffer = remaining_bytes(index + 4)
      @part_name = headers[/Content-Disposition:\s*[^\r\n]*\bname="([^"]+)"/i, 1] ||
                   raise(Sdk::Error, 'Invalid multipart response')
      @part_total_bytes = headers[/Content-Length:\s*(\d+)/i, 1]&.to_i

      @state = :body
      true
    end

    def consume_body? # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
      marker = "\r\n#{boundary}".b
      index = @buffer.index(marker)

      if index
        emit(@buffer.byteslice(0, index))
        @buffer = remaining_bytes(index + marker.bytesize)
        finalize_part!
        @state = :done
        return true
      end

      flushable = @buffer.bytesize - marker.bytesize + 1
      return false if flushable <= 0

      emit(@buffer.byteslice(0, flushable))
      @buffer = remaining_bytes(flushable)
      false
    end

    def emit(data)
      return if data.nil? || data.empty?

      case @part_name
      when 'file'
        @on_file_chunk.call(data)
      when 'error'
        @error_buffer << data
      end
    end

    def finalize_part!
      return unless @part_name == 'error'

      @error_message = extract_error_message(@error_buffer)
    end

    def extract_error_message(payload)
      parsed = JSON.parse(payload)
      parsed['message'] || parsed['error'] || payload
    rescue JSON::ParserError
      payload
    end

    def retain_tail?(size)
      @buffer = @buffer.byteslice(-size, size) || String.new.b if size.positive? && @buffer.bytesize > size
      false
    end

    def remaining_bytes(offset) = @buffer.byteslice(offset, @buffer.bytesize - offset) || String.new.b
    def boundary = "--#{@boundary_token}".b
  end

  module FileTransfer # rubocop:disable Metrics/ModuleLength
    def self.extract_multipart_boundary(content_type)
      match = content_type&.match(/boundary=(?:"([^"]+)"|([^;]+))/i)
      return unless match

      match.captures.compact.first
    end

    def self.assign_download_boundary(parser, content_type)
      boundary = extract_multipart_boundary(content_type)
      raise Sdk::Error, 'Missing multipart boundary in download response' unless boundary

      parser.boundary_token = boundary
    end

    # rubocop:disable Metrics/AbcSize, Metrics/MethodLength, Metrics/CyclomaticComplexity, Metrics/PerceivedComplexity
    def self.stream_download(api_client:, remote_path:, timeout:, on_progress: nil, cancel_event: nil, &block)
      config = api_client.config
      bytes_received = 0
      parser = nil
      wrapped_block = proc do |chunk|
        raise Sdk::Error, "Download cancelled: #{remote_path}" if cancel_event&.set?

        if on_progress
          bytes_received += chunk.bytesize
          on_progress.call(DownloadProgress.new(
                             bytes_received: bytes_received,
                             total_bytes: parser&.part_total_bytes
                           ))
        end
        block.call(chunk)
      end
      parser = MultipartDownloadStreamParser.new(&wrapped_block)
      response = nil

      request = Typhoeus::Request.new(
        "#{config.base_url}/files/bulk-download",
        method: :post,
        headers: api_client.default_headers.dup.merge(
          'Accept' => 'multipart/form-data',
          'Content-Type' => 'application/json'
        ),
        body: JSON.generate(paths: [remote_path]),
        timeout: timeout,
        ssl_verifypeer: config.verify_ssl,
        ssl_verifyhost: config.verify_ssl_host ? 2 : 0
      )

      request.on_headers do |stream_response|
        assign_download_boundary(parser, stream_response.headers['Content-Type'])
      end

      # Returning +:abort+ from the on_body callback tells libcurl to tear down the
      # connection immediately, which is how cancellation actually severs the
      # transfer rather than just stopping our own bookkeeping.
      request.on_body do |chunk|
        next :abort if cancel_event&.set?

        parser << chunk
      end

      request.on_complete do |completed_response|
        response = completed_response
        parser.finish!
      end

      request.run

      raise Sdk::Error, "Download cancelled: #{remote_path}" if cancel_event&.set?
      raise Sdk::Error, parser.error_message if parser.error_message
      raise Sdk::Error, "HTTP #{response.code}" if response && !response.success?
    end
    # rubocop:enable Metrics/AbcSize, Metrics/MethodLength, Metrics/CyclomaticComplexity, Metrics/PerceivedComplexity

    # Uploads +source+ to /files/bulk-upload via Typhoeus (libcurl), which streams the
    # request body straight from disk without buffering it in memory. Local file paths
    # are uploaded directly; in-memory IOs/bytes are first drained to a tempfile so we
    # have a stable file handle for libcurl.
    #
    # The daemon owns atomicity (writes to a sibling tempfile then renames), so a
    # client-side abort just leaves no destination file at all.
    #
    # @param api_client The OpenAPI-generated toolbox API client (auth/base-url only).
    # @param remote_path [String] Destination path in the sandbox.
    # @param source [String, IO] Local file path or any IO-like object responding to +read(n)+.
    # @param timeout [Integer] Typhoeus timeout in seconds (0 disables).
    # @param on_progress [Proc, nil] Optional callback invoked with +Daytona::UploadProgress+
    #   as libcurl reports real network upload progress.
    # @param cancel_event [#set?, nil] Optional cancellation token. Checked while staging
    #   non-file sources and during the libcurl transfer itself.
    # rubocop:disable Metrics/MethodLength, Metrics/ParameterLists
    def self.stream_upload(api_client:, remote_path:, source:, timeout:, on_progress: nil, cancel_event: nil)
      with_upload_file(source, cancel_event, remote_path) do |upload_path|
        config = api_client.config
        progress_callback = upload_progress_callback(on_progress, cancel_event)
        response = with_open_upload_file(upload_path) do |file|
          upload_request(
            api_client: api_client,
            config: config,
            remote_path: remote_path,
            file: file,
            timeout: timeout,
            progress_callback: progress_callback
          ).run
        end
        raise_upload_error(response, cancel_event, remote_path)
      end
    end
    # rubocop:enable Metrics/MethodLength, Metrics/ParameterLists

    # Yields a path on disk that holds the source's bytes, ready for libcurl to stream.
    # Local files are passed through unchanged; everything else is drained into a
    # tempfile that gets unlinked when we return.
    # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
    def self.with_upload_file(source, cancel_event, remote_path)
      raise Sdk::Error, "Upload cancelled: #{remote_path}" if cancel_event&.set?

      return yield(source) if source.is_a?(String) && File.exist?(source)

      tmp = Tempfile.new(['daytona-upload-', File.extname(remote_path).to_s])
      tmp.binmode
      begin
        drain_source_to(source, tmp, cancel_event, remote_path)
        tmp.flush
        tmp.close
        yield(tmp.path)
      ensure
        tmp.close unless tmp.closed?
        begin
          tmp.unlink
        rescue StandardError
          # tempfile already gone, nothing to do
        end
      end
    end
    # rubocop:enable Metrics/AbcSize, Metrics/MethodLength

    def self.drain_source_to(source, sink, cancel_event, remote_path)
      io, owns_io = open_drain_source(source)
      begin
        while (chunk = io.read(64 * 1024))
          break if chunk.empty?
          raise Sdk::Error, "Upload cancelled: #{remote_path}" if cancel_event&.set?

          sink.write(chunk)
        end
      ensure
        io.close if owns_io && io.respond_to?(:close)
      end
    end

    def self.with_open_upload_file(upload_path)
      file = File.open(upload_path, 'rb')
      yield(file)
    ensure
      file.close if file && !file.closed?
    end

    # rubocop:disable Metrics/MethodLength, Metrics/ParameterLists
    def self.upload_request(api_client:, config:, remote_path:, file:, timeout:, progress_callback:)
      Typhoeus::Request.new(
        "#{config.base_url}/files/bulk-upload",
        method: :post,
        headers: api_client.default_headers.dup.tap { |h| h.delete('Content-Type') },
        body: {
          'files[0].path' => remote_path,
          'files[0].file' => file
        },
        timeout: timeout,
        ssl_verifypeer: config.verify_ssl,
        ssl_verifyhost: config.verify_ssl_host ? 2 : 0,
        noprogress: false,
        progressfunction: progress_callback,
        xferinfofunction: progress_callback
      )
    end
    # rubocop:enable Metrics/MethodLength, Metrics/ParameterLists

    def self.upload_progress_callback(on_progress, cancel_event)
      last_bytes_sent = -1

      proc do |_clientp, _dltotal, _dlnow, _ultotal, ulnow|
        next 1 if cancel_event&.set?

        bytes_sent = ulnow.to_i
        if on_progress && bytes_sent > last_bytes_sent
          last_bytes_sent = bytes_sent
          on_progress.call(UploadProgress.new(bytes_sent: bytes_sent))
        end

        0
      end
    end

    def self.raise_upload_error(response, _cancel_event, remote_path)
      raise Sdk::Error, "Upload timed out: #{remote_path}" if response.timed_out?
      raise Sdk::Error, "Upload cancelled: #{remote_path}" if response.return_code == :aborted_by_callback
      raise Sdk::Error, "HTTP #{response.code}: #{response.body}" unless response.success?
    end

    def self.open_drain_source(source)
      return [source, false] if source.respond_to?(:read)
      return [StringIO.new(source.b), true] if source.is_a?(String)

      raise Sdk::Error, "Unsupported upload source: #{source.class}"
    end
  end
end
