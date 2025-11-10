# frozen_string_literal: true

daytona = Daytona::Daytona.new
sandbox = daytona.create

session_id = 'exec-session-1'
sandbox.process.create_session(session_id)

# Get the session details
session = sandbox.process.get_session(session_id)
puts session

# Execute first command in the session
first_command = sandbox.process.execute_session_command(
  session_id:,
  req: Daytona::SessionExecuteRequest.new(command: 'export FOO=BAR')
)

if first_command.exit_code != 0
  puts "Error: #{first_command.exit_code} #{first_command.stderr}"
else
  puts first_command.output
end

# Get the session details again to see the command has been executed
session = sandbox.process.get_session(session_id)
puts session.commands

# Get the command details
command = sandbox.process.get_session_command(session_id:, command_id: first_command.cmd_id)
puts command

# Execute second command in the session and observe the environment variable is set
second_command = sandbox.process.execute_session_command(
  session_id:,
  req: Daytona::SessionExecuteRequest.new(command: 'echo $FOO')
)

if second_command.exit_code != 0
  puts "Error: #{second_command.exit_code} #{second_command.stderr}"
else
  puts second_command.output
end

# Get logs for the second command
logs = sandbox.process.get_session_command_logs(session_id:, command_id: second_command.cmd_id)
puts "[STDOUT] #{logs.stdout}"
puts "[STDERR] #{logs.stderr}"

# List active sessions
sessions = sandbox.process.list_sessions
puts sessions

# Delete the sessio
sandbox.process.delete_session(session_id)

# Cleanup resources
daytona.delete(sandbox)
