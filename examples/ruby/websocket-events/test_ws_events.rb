# frozen_string_literal: true

# Test WebSocket event subscription for sandbox lifecycle.
#
# Tests:
# 1. Event subscriber connects on first sandbox creation
# 2. Sandbox state auto-updates via WebSocket events
# 3. wait_for_sandbox_start/stop use WebSocket events
# 4. get() sandboxes also subscribe to events

require 'daytona'

def test1_subscriber_connects(daytona)
  puts '--- Test 1: Subscriber connects on Daytona construction ---'

  subscriber = daytona.instance_variable_get(:@event_subscriber)
  raise 'Subscriber should exist after Daytona.new' if subscriber.nil?
  puts '  PASS: Subscriber initialized in constructor'

  sandbox = daytona.create
  raise 'Subscriber should be connected after create' unless subscriber.connected?

  puts "  PASS: Subscriber connected (state=#{sandbox.state})"

  daytona.delete(sandbox)
  puts '  PASS: Test 1 complete'
  puts
end

def test2_event_driven_lifecycle(daytona)
  puts '--- Test 2: Event-driven start/stop lifecycle ---'

  sandbox = daytona.create
  puts "  Created sandbox: #{sandbox.id}, state=#{sandbox.state}"

  t0 = Time.now
  daytona.stop(sandbox)
  stop_time = Time.now - t0
  puts "  Stopped in #{stop_time.round(2)}s, state=#{sandbox.state}"
  raise "Expected stopped, got #{sandbox.state}" unless sandbox.state == 'stopped'

  t0 = Time.now
  daytona.start(sandbox)
  start_time = Time.now - t0
  puts "  Started in #{start_time.round(2)}s, state=#{sandbox.state}"
  raise "Expected started, got #{sandbox.state}" unless sandbox.state == 'started'

  daytona.delete(sandbox)
  puts '  PASS: Test 2 complete'
  puts
end

def test3_auto_update_from_events(daytona)
  puts '--- Test 3: Auto-update sandbox state from events ---'

  sandbox = daytona.create
  puts "  Created: state=#{sandbox.state}"

  # Stop and verify state update
  daytona.stop(sandbox)
  puts "  After stop: state=#{sandbox.state}"
  raise "Expected stopped, got #{sandbox.state}" unless sandbox.state == 'stopped'

  daytona.delete(sandbox)
  puts '  PASS: Test 3 complete'
  puts
end

def test4_get_subscribes(daytona)
  puts '--- Test 4: get() sandboxes subscribe to events ---'

  sandbox = daytona.create

  # Get the same sandbox via get()
  sandbox2 = daytona.get(sandbox.id)
  puts "  Got sandbox via get(): state=#{sandbox2.state}"

  # Stop via original
  daytona.stop(sandbox)
  sleep(1)
  puts "  After stop - original: state=#{sandbox.state}, get'd: state=#{sandbox2.state}"
  unless %w[stopped stopping].include?(sandbox2.state)
    raise "get'd sandbox state should update, got #{sandbox2.state}"
  end

  daytona.delete(sandbox)
  puts '  PASS: Test 4 complete'
  puts
end

puts '=' * 60
puts 'WebSocket Event Subscription Test Suite (Ruby)'
puts '=' * 60
puts

daytona = Daytona::Daytona.new

test1_subscriber_connects(daytona)
test2_event_driven_lifecycle(daytona)
test3_auto_update_from_events(daytona)
test4_get_subscribes(daytona)

puts '=' * 60
puts 'ALL TESTS PASSED'
puts '=' * 60
