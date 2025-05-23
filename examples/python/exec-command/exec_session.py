from daytona_sdk import Daytona, SessionExecuteRequest

daytona = Daytona()
sandbox = daytona.create()

exec_session_id = "exec-session-1"
sandbox.process.create_session(exec_session_id)

# Get the session details any time
session = sandbox.process.get_session(exec_session_id)
print(session)

# Execute a first command in the session
execCommand1 = sandbox.process.execute_session_command(exec_session_id, SessionExecuteRequest(command="export FOO=BAR"))
if execCommand1.exit_code != 0:
    print(f"Error: {execCommand1.exit_code} {execCommand1.output}")

# Get the session details again to see the command has been executed
session = sandbox.process.get_session(exec_session_id)
print(session)

# Get the command details
session_command = sandbox.process.get_session_command(exec_session_id, execCommand1.cmd_id)
print(session_command)

# Execute a second command in the session and see that the environment variable is set
execCommand2 = sandbox.process.execute_session_command(exec_session_id, SessionExecuteRequest(command="echo $FOO"))
if execCommand2.exit_code != 0:
    print(f"Error: {execCommand2.exit_code} {execCommand2.output}")
else:
    print(execCommand2.output)

print("Now getting logs for the second command")
logs = sandbox.process.get_session_command_logs(exec_session_id, execCommand2.cmd_id)
print(logs)

# You can also list all active sessions
sessions = sandbox.process.list_sessions()
print(sessions)

# And of course you can delete the session at any time
sandbox.process.delete_session(exec_session_id)

daytona.delete(sandbox)
