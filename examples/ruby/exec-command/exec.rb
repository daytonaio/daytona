#!/usr/bin/env ruby
# frozen_string_literal: true

require 'daytona'

def basic_exec(sandbox)
  # Run some Ruby code directly
  code_result = sandbox.process.code_run(code: 'puts "Hello World from code!"')
  if code_result.exit_code != 0
    puts "Error running code: #{code_result.exit_code}"
  else
    puts code_result.result
  end

  # Run OS command
  cmd_result = sandbox.process.exec(command: 'echo "Hello World from CMD!"')
  if cmd_result.exit_code != 0
    puts "Error running command: #{cmd_result.exit_code}"
  else
    puts cmd_result.result
  end
end

def session_exec(sandbox)
  # Exec session
  # Session allows for multiple commands to be executed in the same context
  sandbox.process.create_session('exec-session-1')

  # Get the session details any time
  session = sandbox.process.get_session('exec-session-1')
  puts "session: #{session.inspect}"

  # Execute a first command in the session
  command = sandbox.process.execute_session_command(
    session_id: 'exec-session-1',
    req: { command: 'export FOO=BAR' }
  )

  # Get the session details again to see the command has been executed
  session_updated = sandbox.process.get_session('exec-session-1')
  puts "sessionUpdated: #{session_updated.inspect}"

  # Get the command details
  session_command = sandbox.process.get_session_command(
    session_id: 'exec-session-1',
    command_id: command.cmd_id
  )
  puts "sessionCommand: #{session_command.inspect}"

  # Execute a second command in the session and see that the environment variable is set
  response = sandbox.process.execute_session_command(
    session_id: 'exec-session-1',
    req: { command: 'echo $FOO' }
  )
  puts "FOO=#{response.stdout}"

  # We can also get the logs for the command any time after it is executed
  logs = sandbox.process.get_session_command_logs(
    session_id: 'exec-session-1',
    command_id: response.cmd_id
  )
  puts "[STDOUT]: #{logs[:stdout]}"
  puts "[STDERR]: #{logs[:stderr]}"

  # We can also delete the session
  sandbox.process.delete_session('exec-session-1')
end

def session_exec_logs_async(sandbox)
  puts 'Executing long running command in a session and streaming logs asynchronously...'

  session_id = 'exec-session-async-logs'
  sandbox.process.create_session(session_id)

  command = sandbox.process.execute_session_command(
    session_id:,
    req: {
      command: 'counter=1; while (( counter <= 3 )); do echo "Count: $counter"; ((counter++)); sleep 2; done; non-existent-command',
      run_async: true
    }
  )

  sandbox.process.get_session_command_logs_async(
    session_id:,
    command_id: command.cmd_id,
    on_stdout: ->(stdout) { puts "[STDOUT]: #{stdout}" },
    on_stderr: ->(stderr) { puts "[STDERR]: #{stderr}" }
  )
end

def stateful_code_interpreter(sandbox) # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
  log_stdout = ->(msg) { print "[STDOUT] #{msg.output}" }
  log_stderr = ->(msg) { print "[STDERR] #{msg.output}" }
  log_error = lambda do |err|
    print "[ERROR] #{err.name}: #{err.value}\n"
    print "#{err.traceback}\n" unless err.traceback.empty?
  end

  puts "\n#{'=' * 60}"
  puts 'Stateful Code Interpreter'
  puts '=' * 60

  puts '=' * 10 + ' Statefulness in the default context ' + '=' * 10
  result = sandbox.code_interpreter.run_code(
    "counter = 1\nprint(f'Initialized counter = {counter}')"
  )
  print "[STDOUT] #{result.stdout}"

  sandbox.code_interpreter.run_code(
    "counter += 1\nprint(f'Counter after second call = {counter}')",
    on_stdout: log_stdout,
    on_stderr: log_stderr,
    on_error: log_error
  )

  puts '=' * 10 + ' Context isolation ' + '=' * 10
  ctx = sandbox.code_interpreter.create_context
  begin
    sandbox.code_interpreter.run_code(
      "value = 'stored in isolated context'\nprint(f'Isolated context value: {value}')",
      context: ctx,
      on_stdout: log_stdout,
      on_stderr: log_stderr,
      on_error: log_error
    )

    puts '-' * 3 + ' Print value from same context ' + '-' * 3
    ctx_result = sandbox.code_interpreter.run_code(
      "print(f'Value still available: {value}')",
      context: ctx
    )
    print "[STDOUT] #{ctx_result.stdout}"

    puts '-' * 3 + ' Print value from different context ' + '-' * 3
    sandbox.code_interpreter.run_code(
      'print(value)',
      on_stdout: log_stdout,
      on_stderr: log_stderr,
      on_error: log_error
    )
  ensure
    sandbox.code_interpreter.delete_context(ctx)
  end

  puts '=' * 10 + ' Timeout handling ' + '=' * 10
  begin
    code = <<~PYTHON
      import time
      print('Starting long running task...')
      time.sleep(5)
      print('Finished!')
    PYTHON

    sandbox.code_interpreter.run_code(
      code,
      timeout: 1,
      on_stdout: log_stdout,
      on_stderr: log_stderr,
      on_error: log_error
    )
  rescue Daytona::Sdk::Error => e
    puts "Timed out as expected: #{e.message}" if e.message.include?('timed out')
  end
end

def main # rubocop:disable Metrics/MethodLength
  daytona = Daytona::Daytona.new

  # First, create a sandbox
  image = Daytona::Image.base('ubuntu:22.04').run_commands(
    'apt-get update',
    'apt-get install -y --no-install-recommends python3 python3-pip python3-venv',
    'apt-get install -y --no-install-recommends coreutils',
    'apt-get install -y --no-install-recommends ruby'
  )

  params = Daytona::CreateSandboxFromImageParams.new(
    image:,
    language: 'python',
    auto_stop_interval: 60,
    auto_archive_interval: 60,
    auto_delete_interval: 120,
    resources: Daytona::Resources.new(cpu: 2, memory: 2, disk: 10)
  )

  sandbox = daytona.create(
    params,
    on_snapshot_create_logs: ->(chunk) { print chunk }
  )

  begin
    basic_exec(sandbox)
    session_exec(sandbox)
    session_exec_logs_async(sandbox)
    stateful_code_interpreter(sandbox)
  rescue StandardError => e
    puts "Error executing commands: #{e.message}"
    puts e.backtrace
  ensure
    # Cleanup
    daytona.delete(sandbox)
  end
end

main if __FILE__ == $PROGRAM_NAME
