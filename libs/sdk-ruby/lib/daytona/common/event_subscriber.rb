# frozen_string_literal: true

# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

require_relative 'socketio_client'

module Daytona
  # Manages a Socket.IO connection and dispatches events to per-resource handlers.
  # Generic — works for sandboxes, volumes, snapshots, runners, etc.
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
      @registered_events = Set.new
      @mutex = Mutex.new
      @disconnect_timer = nil
      @reconnect_thread = nil
      @last_event_at = Time.now
      @reconnecting = false
      @close_requested = false
      @max_reconnects = 100
    end

    # Idempotent: ensure a connection attempt is in progress or already established.
    # Non-blocking. Starts a background Thread to connect if not already connected
    # and no attempt is currently running.
    # @return [void]
    def ensure_connected
      return if @connected
      return if @connect_thread&.alive?

      @connect_thread = Thread.new do
        connect
      rescue StandardError
        # Callers check connected? when they need it
      end
    end

    # Establish the Socket.IO connection.
    # @return [void]
    # @raise [StandardError] on connection failure
    def connect
      return if @connected

      # Close any existing stale connection before creating a fresh one
      @client&.close rescue nil # rubocop:disable Style/RescueModifier

      @client = SocketIOClient.new(
        api_url: @api_url,
        token: @token,
        organization_id: @organization_id,
        on_event: method(:handle_event),
        on_disconnect: method(:handle_disconnect)
      )

      @close_requested = false
      @client.connect
      @connected = true
      @failed = false
      @fail_error = nil
    rescue StandardError => e
      @failed = true
      @fail_error = "WebSocket connection failed: #{e.message}"
    end

    # Subscribe to specific events for a resource.
    # @param resource_id [String] The ID of the resource (e.g. sandbox ID, volume ID).
    # @param events [Array<String>] List of Socket.IO event names to listen for.
    # @yield [event_name, data] Called with raw event name and data hash.
    # @return [Proc] Unsubscribe function.
    DISCONNECT_DELAY = 30

    def subscribe(resource_id, events:, &handler)
      # Cancel any pending delayed disconnect
      @disconnect_timer&.kill
      @disconnect_timer = nil

      # Register any new events with the Socket.IO client (idempotent)
      register_events(events)

      @mutex.synchronize do
        @listeners[resource_id] ||= []
        @listeners[resource_id] << handler
      end

      lambda {
        return if @close_requested

        should_schedule = false
        @mutex.synchronize do
          @listeners[resource_id]&.delete(handler)
          @listeners.delete(resource_id) if @listeners[resource_id] && @listeners[resource_id].empty?
          should_schedule = @listeners.empty?
        end

        # Schedule delayed disconnect when no resources are listening anymore
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
    attr_reader :fail_error

    # Disconnect and clean up.
    def disconnect
      @close_requested = true
      @reconnect_thread&.kill
      @reconnect_thread = nil
      @disconnect_timer&.kill
      @disconnect_timer = nil
      @client&.close
      @connected = false
      @mutex.synchronize { @listeners.clear }
      @registered_events.clear
    end

    private

    # Register Socket.IO event handlers (idempotent - each event is registered once).
    # The SocketIOClient dispatches all events via the on_event callback, so we just
    # need to track which events we care about for filtering in handle_event.
    def register_events(events)
      events.each { |evt| @registered_events.add(evt) }
    end

    def handle_event(event_name, data)
      @last_event_at = Time.now

      # Only dispatch events that have been registered
      return unless @registered_events.include?(event_name)

      resource_id = extract_id_from_event(data)
      return unless resource_id

      dispatch(resource_id, event_name, data)
    end

    # Extract resource ID from an event payload.
    # Handles two payload shapes:
    #   - Wrapper: {sandbox: {id: ...}, ...} -> nested resource ID
    #   - Direct: {id: ...} -> top-level ID
    def extract_id_from_event(data)
      return nil unless data.is_a?(Hash)

      %w[sandbox volume snapshot runner].each do |key|
        nested = data[key]
        next unless nested.is_a?(Hash)

        sid = nested['id']
        return sid if sid.is_a?(String)
      end

      top_id = data['id']
      return top_id if top_id.is_a?(String)

      nil
    end

    def dispatch(resource_id, event_name, data)
      handlers = @mutex.synchronize { @listeners[resource_id]&.dup || [] }
      handlers.each do |handler|
        handler.call(event_name, data)
      rescue StandardError
        # Don't let handler errors break other handlers
      end
    end

    def handle_disconnect
      @connected = false
      return if @close_requested

      @reconnect_thread = Thread.new { reconnect_loop }
    end

    def reconnect_loop
      @mutex.synchronize do
        return if @reconnecting

        @reconnecting = true
      end

      attempt = 0
      while attempt < @max_reconnects
        if @close_requested
          @mutex.synchronize { @reconnecting = false }
          return
        end

        delay = [2**attempt, 30].min
        sleep(delay)
        if @close_requested
          @mutex.synchronize { @reconnecting = false }
          return
        end

        begin
          connect
          if @connected
            @mutex.synchronize { @reconnecting = false }
            return
          end
        rescue StandardError
          # Continue retrying
        end

        attempt += 1
      end

      # All attempts failed
      @mutex.synchronize do
        @failed = true
        @fail_error = "WebSocket reconnection failed after #{@max_reconnects} attempts"
        @reconnecting = false
      end
    end
  end
end
