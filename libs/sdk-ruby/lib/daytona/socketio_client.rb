# frozen_string_literal: true

# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

require 'websocket-client-simple'
require 'json'
require 'uri'
require 'thread'

module Daytona
  # Minimal Engine.IO/Socket.IO v4 client over raw WebSocket.
  # Supports connect with auth, heartbeat, and event reception.
  class SocketIOClient
    # Engine.IO v4 packet types
    EIO_OPEN    = '0'
    EIO_CLOSE   = '1'
    EIO_PING    = '2'
    EIO_PONG    = '3'
    EIO_MESSAGE = '4'

    # Socket.IO v4 packet types (inside Engine.IO messages)
    SIO_CONNECT       = '0'
    SIO_DISCONNECT    = '1'
    SIO_EVENT         = '2'
    SIO_CONNECT_ERROR = '4'

    attr_reader :connected

    # @param api_url [String] The API URL (e.g., "https://app.daytona.io/api")
    # @param token [String] Auth token (API key or JWT)
    # @param organization_id [String, nil] Organization ID for room joining
    # @param on_event [Proc] Called with (event_name, data_hash) for each Socket.IO event
    # @param on_disconnect [Proc] Called when the connection is lost
    # @param connect_timeout [Numeric] Connection timeout in seconds
    def initialize(api_url:, token:, organization_id: nil, on_event: nil, on_disconnect: nil, connect_timeout: 5)
      @api_url = api_url
      @token = token
      @organization_id = organization_id
      @on_event = on_event
      @on_disconnect = on_disconnect
      @connect_timeout = connect_timeout
      @connected = false
      @mutex = Mutex.new
      @write_mutex = Mutex.new
      @ping_thread = nil
      @ping_interval = 25
      @ws = nil
      @close_requested = false
    end

    # Establish the WebSocket connection and perform Socket.IO handshake.
    # @return [Boolean] true if connection succeeded
    # @raise [StandardError] on connection failure
    def connect
      ws_url = build_ws_url
      connected_queue = Queue.new

      # Capture self because websocket-client-simple uses instance_exec for callbacks
      client = self

      @ws = WebSocket::Client::Simple.connect(ws_url)

      @ws.on :message do |msg|
        client.send(:handle_raw_message, msg.data.to_s, connected_queue)
      end

      @ws.on :error do |_e|
        client.instance_variable_get(:@mutex).synchronize do
          client.instance_variable_set(:@connected, false)
        end
        connected_queue.push(:error) unless client.connected?
      end

      @ws.on :close do
        mutex = client.instance_variable_get(:@mutex)
        was_connected = mutex.synchronize do
          prev = client.instance_variable_get(:@connected)
          client.instance_variable_set(:@connected, false)
          prev
        end
        on_disconnect = client.instance_variable_get(:@on_disconnect)
        close_requested = client.instance_variable_get(:@close_requested)
        on_disconnect&.call if was_connected && !close_requested
      end

      # Wait for connection with timeout
      result = nil
      begin
        Timeout.timeout(@connect_timeout) { result = connected_queue.pop }
      rescue Timeout::Error
        close
        raise "WebSocket connection timed out after #{@connect_timeout}s"
      end

      raise "WebSocket connection failed: #{result}" if result != :connected

      @mutex.synchronize { @connected }
    end

    # @return [Boolean]
    def connected?
      @mutex.synchronize { @connected }
    end

    # Gracefully close the connection.
    def close
      @close_requested = true
      @ping_thread&.kill
      @ping_thread = nil

      send_raw(EIO_CLOSE) if @ws
      @ws&.close
      @mutex.synchronize { @connected = false }
    rescue StandardError
      # Ignore errors during close
    end

    private

    def build_ws_url
      parsed = URI.parse(@api_url)
      ws_scheme = parsed.scheme == 'https' ? 'wss' : 'ws'
      host = parsed.host
      port = parsed.port

      query_parts = ['EIO=4', 'transport=websocket']
      query_parts << "organizationId=#{URI.encode_www_form_component(@organization_id)}" if @organization_id

      port_str = (parsed.scheme == 'https' && port == 443) || (parsed.scheme == 'http' && port == 80) ? '' : ":#{port}"
      "#{ws_scheme}://#{host}#{port_str}/api/socket.io/?#{query_parts.join('&')}"
    end

    def handle_raw_message(raw, connected_queue)
      return if raw.nil? || raw.empty?

      case raw[0]
      when EIO_OPEN
        # Parse open payload for ping interval
        begin
          payload = JSON.parse(raw[1..])
          @ping_interval = (payload['pingInterval'] || 25_000) / 1000.0
        rescue JSON::ParserError
          # Use default ping interval
        end
        # Send Socket.IO CONNECT with auth
        auth = JSON.generate({ token: @token })
        send_raw("#{EIO_MESSAGE}#{SIO_CONNECT}#{auth}")

      when EIO_PING
        send_raw(EIO_PONG)

      when EIO_MESSAGE
        handle_socketio_packet(raw[1..], connected_queue)

      when EIO_CLOSE
        @mutex.synchronize { @connected = false }
      end
    end

    def handle_socketio_packet(data, connected_queue)
      return if data.nil? || data.empty?

      case data[0]
      when SIO_CONNECT
        # Connection acknowledged
        @mutex.synchronize { @connected = true }
        start_ping_thread
        connected_queue&.push(:connected)

      when SIO_CONNECT_ERROR
        # Connection rejected
        error_msg = begin
          payload = JSON.parse(data[1..])
          payload['message'] || 'Unknown error'
        rescue JSON::ParserError
          data[1..]
        end
        @mutex.synchronize { @connected = false }
        connected_queue&.push("Auth rejected: #{error_msg}")

      when SIO_EVENT
        handle_event(data[1..])

      when SIO_DISCONNECT
        @mutex.synchronize { @connected = false }
      end
    end

    def handle_event(json_str)
      return unless @on_event

      # Skip namespace prefix if present (e.g., "/ns,")
      if json_str&.start_with?('/')
        comma_idx = json_str.index(',')
        json_str = json_str[(comma_idx + 1)..] if comma_idx
      end

      event_array = JSON.parse(json_str)
      return unless event_array.is_a?(Array) && event_array.length >= 1

      event_name = event_array[0]
      event_data = event_array[1]

      @on_event.call(event_name, event_data)
    rescue JSON::ParserError
      # Malformed event, ignore
    end

    def start_ping_thread
      @ping_thread&.kill
      @ping_thread = Thread.new do
        loop do
          sleep(@ping_interval)
          break unless connected?

          send_raw(EIO_PING)
        rescue StandardError
          break
        end
      end
    end

    def send_raw(msg)
      @write_mutex.synchronize do
        @ws&.send(msg)
      end
    rescue StandardError
      # Ignore write errors
    end
  end
end
