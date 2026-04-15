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
      @subscription_timers = {}
      @subscription_ttls = {}
      @reconnect_thread = nil
      @connect_thread = nil
      @last_event_at = Time.now
      @reconnecting = false
      @connecting = false
      @close_requested = false
      @closed = false
      @max_reconnects = 100
    end

    # Idempotent: ensure a connection attempt is in progress or already established.
    # Non-blocking. Starts a background Thread to connect if not already connected
    # and no attempt is currently running.
    # @return [void]
    def ensure_connected
      @mutex.synchronize do
        return if @closed || @connected || @connecting || @connect_thread&.alive?

        @connect_thread = Thread.new do
          connect
        rescue StandardError
          # Callers check connected? when they need it
        ensure
          @mutex.synchronize do
            @connect_thread = nil if @connect_thread == Thread.current
          end
        end
      end
    end

    # Establish the Socket.IO connection.
    # @return [void]
    # @raise [StandardError] on connection failure
    def connect
      @mutex.synchronize do
        return if @closed || @connected || @connecting

        @connecting = true
      end

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
    ensure
      @mutex.synchronize { @connecting = false }
    end

    # Subscribe to specific events for a resource.
    # @param resource_id [String] The ID of the resource (e.g. sandbox ID, volume ID).
    # @param events [Array<String>] List of Socket.IO event names to listen for.
    # @yield [event_name, data] Called with raw event name and data hash.
    # @return [Proc] Unsubscribe function.
    DISCONNECT_DELAY = 30

    def subscribe(resource_id, events:, ttl: 0, &handler)
      ensure_connected

      timer_to_stop = nil

      @mutex.synchronize do
        return -> {} if @closed

        # Register any new events with the Socket.IO client (idempotent)
        register_events(events)

        @listeners[resource_id] ||= []
        @listeners[resource_id] << handler

        @disconnect_timer&.kill
        @disconnect_timer = nil

        if ttl.positive?
          @subscription_ttls[resource_id] = ttl
          timer_to_stop = start_subscription_timer_locked(resource_id)
        else
          timer_to_stop = cancel_subscription_timer_locked(resource_id)
          @subscription_ttls.delete(resource_id)
        end
      end

      stop_thread(timer_to_stop)

      lambda {
        return if @close_requested || @closed

        should_schedule = false
        timer_to_stop = nil
        @mutex.synchronize do
          @listeners[resource_id]&.delete(handler)
          if @listeners[resource_id] && @listeners[resource_id].empty?
            timer_to_stop = unsubscribe_resource_locked(resource_id)
          end
          should_schedule = @listeners.empty?

          schedule_delayed_disconnect_locked if should_schedule
        end

        stop_thread(timer_to_stop)
      }
    end

    def refresh_subscription(resource_id)
      timer_to_stop = nil

      @mutex.synchronize do
        return false unless @subscription_ttls.key?(resource_id)

        timer_to_stop = start_subscription_timer_locked(resource_id)
      end

      stop_thread(timer_to_stop)
      true
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
      thread_state = @mutex.synchronize do
        begin_disconnect_locked(permanent: true, skip_thread: Thread.current)
      end

      finalize_disconnect(thread_state)
    end

    private

    # Register Socket.IO event handlers (idempotent - each event is registered once).
    # The SocketIOClient dispatches all events via the on_event callback, so we just
    # need to track which events we care about for filtering in handle_event.
    def register_events(events)
      events.each { |evt| @registered_events.add(evt) }
    end

    def start_subscription_timer_locked(resource_id)
      ttl = @subscription_ttls[resource_id]
      return unless ttl&.positive?

      previous_timer = cancel_subscription_timer_locked(resource_id)
      timer = Thread.new do
        sleep(ttl)
        @mutex.synchronize do
          current_timer = @subscription_timers[resource_id]
          next unless current_timer == Thread.current

          @subscription_timers.delete(resource_id)
          @subscription_ttls.delete(resource_id)
          @listeners.delete(resource_id)
          schedule_delayed_disconnect_locked if @listeners.empty?
        end
      end
      timer.abort_on_exception = false
      @subscription_timers[resource_id] = timer
      previous_timer
    end

    def cancel_subscription_timer_locked(resource_id)
      @subscription_timers.delete(resource_id)
    end

    def cancel_subscription_timers_locked
      timers = @subscription_timers.values
      @subscription_timers = {}
      @subscription_ttls = {}
      timers
    end

    def unsubscribe_resource_locked(resource_id)
      @listeners.delete(resource_id)
      @subscription_ttls.delete(resource_id)
      cancel_subscription_timer_locked(resource_id)
    end

    def schedule_delayed_disconnect_locked
      @disconnect_timer&.kill
      @disconnect_timer = Thread.new do
        sleep(DISCONNECT_DELAY)
        thread_state = @mutex.synchronize do
          next unless @listeners.empty?

          begin_disconnect_locked(permanent: false, skip_thread: Thread.current)
        end

        finalize_disconnect(thread_state) if thread_state
      end
      @disconnect_timer.abort_on_exception = false
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

    def begin_disconnect_locked(permanent:, skip_thread:)
      @closed = true if permanent
      @close_requested = true

      thread_state = {
        client: @client,
        reconnect_thread: @reconnect_thread,
        connect_thread: @connect_thread,
        disconnect_timer: @disconnect_timer,
        subscription_timers: cancel_subscription_timers_locked,
        skip_thread:
      }

      @client = nil
      @reconnect_thread = nil
      @connect_thread = nil
      @disconnect_timer = nil
      @connected = false
      @connecting = false
      @listeners.clear
      @registered_events.clear

      thread_state
    end

    def finalize_disconnect(thread_state)
      return unless thread_state

      [thread_state[:reconnect_thread], thread_state[:connect_thread], thread_state[:disconnect_timer],
       *thread_state[:subscription_timers]].each do |thread|
        next unless thread
        next if thread == thread_state[:skip_thread]

        stop_thread(thread)
      end

      thread_state[:client]&.close
    rescue StandardError
      nil
    end

    def stop_thread(thread)
      return unless thread
      return if thread == Thread.current

      thread.kill
      thread.join
    rescue StandardError
      nil
    end
  end
end
