# frozen_string_literal: true

# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

require 'securerandom'

module Daytona
  class EventSubscriptionManager
    SUBSCRIPTION_TTL = 300
    private_constant :SUBSCRIPTION_TTL

    def initialize(dispatcher)
      @dispatcher = dispatcher
      @subscriptions = {}
      @mutex = Mutex.new
      @closed = false
    end

    def subscribe(resource_id:, handler:, events:)
      @mutex.synchronize do
        # Reject after shutdown to prevent use-after-close
        return nil if @closed
      end

      unsubscribe = @dispatcher.subscribe(resource_id, events:, &handler)
      sub_id = SecureRandom.hex(16)
      timer_to_stop = nil
      rollback_unsubscribe = nil

      @mutex.synchronize do
        if @closed
          # Rollback dispatcher subscription on failure
          rollback_unsubscribe = unsubscribe
          next
        end

        @subscriptions[sub_id] = { unsubscribe:, timer: nil }
        timer_to_stop = start_timer_locked(sub_id)
      end

      if rollback_unsubscribe
        rollback_unsubscribe.call
        return nil
      end

      stop_thread(timer_to_stop)
      sub_id
    end

    def refresh(sub_id)
      @mutex.synchronize do
        # Reject after shutdown to prevent use-after-close
        return false if @closed
      end

      timer_to_stop = nil

      @mutex.synchronize do
        return false unless @subscriptions.key?(sub_id)

        timer_to_stop = start_timer_locked(sub_id)
      end

      stop_thread(timer_to_stop)
      true
    end

    def unsubscribe(sub_id)
      subscription = nil
      timer = nil

      @mutex.synchronize do
        subscription = @subscriptions.delete(sub_id)
        if subscription
          timer = subscription[:timer]
          subscription[:timer] = nil
        end
      end

      return unless subscription

      stop_thread(timer)
      subscription[:unsubscribe].call
    end

    def shutdown
      subscriptions = nil

      @mutex.synchronize do
        @closed = true
        subscriptions = @subscriptions.values
        @subscriptions = {}
      end

      subscriptions.each do |subscription|
        stop_thread(subscription[:timer])
        subscription[:unsubscribe].call
      end
    end

    private

    def start_timer_locked(sub_id)
      subscription = @subscriptions[sub_id]
      return unless subscription

      previous_timer = cancel_timer_locked(sub_id)
      timer = Thread.new do
        sleep(SUBSCRIPTION_TTL)

        unsubscribe = @mutex.synchronize do
          current_subscription = @subscriptions[sub_id]
          next unless current_subscription && current_subscription[:timer] == Thread.current

          @subscriptions.delete(sub_id)&.dig(:unsubscribe)
        end

        unsubscribe&.call
      end
      timer.abort_on_exception = false
      subscription[:timer] = timer
      previous_timer
    end

    def cancel_timer_locked(sub_id)
      subscription = @subscriptions[sub_id]
      return unless subscription

      timer = subscription[:timer]
      subscription[:timer] = nil
      timer
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
