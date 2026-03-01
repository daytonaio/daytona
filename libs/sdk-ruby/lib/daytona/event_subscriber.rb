# frozen_string_literal: true

# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

require_relative 'socketio_client'

module Daytona
  # Manages a Socket.IO connection and dispatches sandbox events to per-sandbox handlers.
  class EventSubscriber
    # @param api_url [String]
    # @param token [String]
    # @param organization_id [String, nil]
    def initialize(api_url:, token:, organization_id: nil)
      @api_url = api_url
      @token = token
      @organization_id = organization_id
      @client = nil
      @connected = false
      @failed = false
      @fail_error = nil
      @listeners = {}
      @mutex = Mutex.new
      @disconnect_timer = nil
      @reconnecting = false
      @close_requested = false
      @max_reconnects = 10
    end

    # Establish the Socket.IO connection.
    # @return [void]
    # @raise [StandardError] on connection failure
    def connect
      return if @connected

      @client = SocketIOClient.new(
        api_url: @api_url,
        token: @token,
        organization_id: @organization_id,
        on_event: method(:handle_event),
        on_disconnect: method(:handle_disconnect)
      )

      @client.connect
      @connected = true
      @failed = false
      @fail_error = nil
    rescue StandardError => e
      @failed = true
      @fail_error = "WebSocket connection failed: #{e.message}"
      raise
    end

    # Subscribe to events for a specific sandbox.
    # @param sandbox_id [String]
    # @yield [event_type, data] Called with event type and data hash
    # @return [Proc] Unsubscribe function
    DISCONNECT_DELAY = 30

    def subscribe(sandbox_id, &handler)
      # Cancel any pending delayed disconnect
      @disconnect_timer&.kill
      @disconnect_timer = nil

      @mutex.synchronize do
        @listeners[sandbox_id] ||= []
        @listeners[sandbox_id] << handler
      end

      -> {
        should_schedule = false
        @mutex.synchronize do
          @listeners[sandbox_id]&.delete(handler)
          @listeners.delete(sandbox_id) if @listeners[sandbox_id]&.empty?
          should_schedule = @listeners.empty?
        end

        # Schedule delayed disconnect when no sandboxes are listening anymore
        if should_schedule
          @disconnect_timer = Thread.new do
            sleep(DISCONNECT_DELAY)
            empty = @mutex.synchronize { @listeners.empty? }
            disconnect if empty
          end
        end
      }
    end

    # @return [Boolean]
    def connected?
      @connected
    end

    # @return [Boolean]
    def failed?
      @failed
    end

    # @return [String, nil]
    def fail_error
      @fail_error
    end

    # Disconnect and clean up.
    def disconnect
      @close_requested = true
      @client&.close
      @connected = false
      @mutex.synchronize { @listeners.clear }
    end

    private

    def handle_event(event_name, data)
      sandbox_id = extract_sandbox_id(event_name, data)
      return unless sandbox_id

      event_type = case event_name
                   when 'sandbox.state.updated' then 'state.updated'
                   when 'sandbox.desired-state.updated' then 'desired-state.updated'
                   when 'sandbox.created' then 'created'
                   else return
                   end

      dispatch(sandbox_id, event_type, data)
    end

    def extract_sandbox_id(event_name, data)
      return nil unless data.is_a?(Hash)

      case event_name
      when 'sandbox.state.updated', 'sandbox.desired-state.updated'
        data.dig('sandbox', 'id')
      when 'sandbox.created'
        data['id']
      end
    end

    def dispatch(sandbox_id, event_type, data)
      handlers = @mutex.synchronize { @listeners[sandbox_id]&.dup || [] }
      handlers.each do |handler|
        handler.call(event_type, data)
      rescue StandardError
        # Don't let handler errors break other handlers
      end
    end

    def handle_disconnect
      @connected = false
      return if @close_requested

      Thread.new { reconnect_loop }
    end

    def reconnect_loop
      return if @reconnecting

      @reconnecting = true

      @max_reconnects.times do |attempt|
        return if @close_requested

        delay = [2**attempt, 30].min
        sleep(delay)
        return if @close_requested

        connect
        @reconnecting = false
        return
      rescue StandardError
        # Continue retrying
      end

      # All attempts failed
      @failed = true
      @fail_error = "WebSocket reconnection failed after #{@max_reconnects} attempts"
      @reconnecting = false
    end
  end
end
