# frozen_string_literal: true

daytona = Daytona::Daytona.new
sandbox = daytona.create

puts '=== First PTY Session: Interactive Command with Exit ==='
pty_session_id = 'interactive-pty-session'

# Create PTY session
handle = sandbox.process.create_pty_session(id: pty_session_id, pty_size: Daytona::PtySize.new(cols: 120, rows: 30))

thread = Thread.new do
  # Using iterator to handle PTY data
  print("\n--- Using iterator approach to handle PTY output ---")
  handle.each { |data| puts data }
end

# Send interactive command
handle.send_input("printf 'Enter your name: ' && read name && printf 'Hello %s\n' \"$name\"\n")

# Wait and respond
sleep(1)
handle.send_input("Alice\n")

handle.resize(Daytona::PtySize.new(cols: 80, rows: 25))

# Send another command
sleep(1)
handle.send_input("ls -la\n")

# Send exit command
sleep(1)
handle.send_input("exit\n")

thread.join

puts "\nPTY session exited with code: #{handle.exit_code}"
puts "Error: #{handle.error}" if handle.error

puts '=== Second PTY Session: Kill PTY Session ==='

pty_session_id = 'kill-pty-session'

# Create PTY session
handle = sandbox.process.create_pty_session(id: pty_session_id, pty_size: Daytona::PtySize.new(cols: 120, rows: 30))

# Send a long-running command
handle.send_input("while true; do echo \"Running... $(date)\"; sleep 1; done\n")

[
  Thread.new { handle.each { |data| puts data } },
  Thread.new { sleep(3) and handle.kill }
].each(&:join)

puts "PTY session terminated. Exit code: #{handle.exit_code}"
puts "Error: #{handle.error}" if handle.error

# Cleanup
daytona.delete(sandbox)
